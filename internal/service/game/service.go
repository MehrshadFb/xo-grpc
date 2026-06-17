package gamesvc

import (
	"errors"
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

func (s *Service) RequestRematch(gameID, playerToken string) (*RequestRematchResult, error) {
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

	var result *RequestRematchResult

	for attempt := 0; attempt < 3; attempt++ {
		result, err = s.requestRematchOnce(gameID, playerSession.Mark)
		if errors.Is(err, repository.ErrConflict) {
			continue
		}
		if err != nil {
			return nil, err
		}
		break
	}

	if err != nil {
		return nil, err
	}

	if s.hub != nil && result.Changed {
		eventType := realtime.EventTypeRematchRequested
		if result.Started {
			eventType = realtime.EventTypeRoundStarted
		}

		s.hub.Publish(result.Game.ID, realtime.Event{
			Type: eventType,
			Game: result.Game,
		})
	}

	return result, nil
}

func (s *Service) requestRematchOnce(gameID string, mark domaingame.Mark) (*RequestRematchResult, error) {
	g, err := s.games.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	rematchResult, err := g.RequestRematch(mark)
	if err != nil {
		return nil, err
	}

	if rematchResult == domaingame.RematchNoop {
		return &RequestRematchResult{
			Game: g,
		}, nil
	}

	if err := s.games.Update(g); err != nil {
		return nil, err
	}

	return &RequestRematchResult{
		Game:    g,
		Started: rematchResult == domaingame.RematchStarted,
		Changed: true,
	}, nil
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
