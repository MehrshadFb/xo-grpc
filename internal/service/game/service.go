package gamesvc

import (
	"strings"

	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
)

type Service struct {
	store    *memory.Store
	sessions *session.Manager
}

func NewService(store *memory.Store, sessions *session.Manager) *Service {
	return &Service{
		store:    store,
		sessions: sessions,
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

	g, err := s.store.GetByID(gameID)
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

	g, err := s.store.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	if err := g.ApplyMove(playerSession.Mark, cellIndex); err != nil {
		return nil, err
	}

	if err := s.store.Update(g); err != nil {
		return nil, err
	}

	return &MakeMoveResult{Game: g}, nil
}