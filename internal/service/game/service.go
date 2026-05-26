package gamesvc

import (
	"strings"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
)

type Service struct {
	games    repository.GameRepository
	sessions *session.Manager
	hub      *realtime.Hub
}

func NewService(games repository.GameRepository, sessions *session.Manager, hub *realtime.Hub) *Service {
	return &Service{
		games:    games,
		sessions: sessions,
		hub:      hub,
	}
}

func (s *Service) GetState(gameID, playerToken string) (*GetStateResult, error) {
	gameID = strings.TrimSpace(gameID)
	playerToken = strings.TrimSpace(playerToken)

	if gameID == "" {
		return nil, ErrInvalidGameID
	}
	if playerToken == "" {
		return nil, ErrInvalidToken
	}

	if _, err := s.sessions.ValidateGame(playerToken, gameID); err != nil {
		return nil, err
	}

	g, err := s.games.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	return &GetStateResult{Game: g}, nil
}

func (s *Service) MakeMove(gameID, playerToken string, cellIndex int) (*MakeMoveResult, error) {
	gameID = strings.TrimSpace(gameID)
	playerToken = strings.TrimSpace(playerToken)

	if gameID == "" {
		return nil, ErrInvalidGameID
	}
	if playerToken == "" {
		return nil, ErrInvalidToken
	}

	playerSession, err := s.sessions.ValidateGame(playerToken, gameID)
	if err != nil {
		return nil, err
	}

	g, err := s.games.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	if err := g.ApplyMove(playerSession.Mark, cellIndex); err != nil {
		return nil, err
	}

	if err := s.games.Update(g); err != nil {
		return nil, err
	}

	if s.hub != nil {
		eventType := realtime.EventTypeMoveMade
		gameOverReason := ""

		if g.Status == domaingame.StatusFinished {
			eventType = realtime.EventTypeGameOver
			if g.IsDraw {
				gameOverReason = "draw"
			} else {
				gameOverReason = "win"
			}
		}

		s.hub.Publish(g.ID, realtime.Event{
			Type:           eventType,
			Game:           g,
			GameOverReason: gameOverReason,
		})
	}

	return &MakeMoveResult{Game: g}, nil
}
