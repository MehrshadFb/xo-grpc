package memory

import (
	"errors"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
)

func cloneGame(g *game.Game) *game.Game {
	if g == nil {
		return nil
	}

	clone := *g

	if g.PlayerX != nil {
		playerX := *g.PlayerX
		clone.PlayerX = &playerX
	}

	if g.PlayerO != nil {
		playerO := *g.PlayerO
		clone.PlayerO = &playerO
	}

	return &clone
}

func (s *Store) Create(g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if g == nil {
		return errors.New("nil game")
	}
	if g.ID == "" {
		return errors.New("empty game id")
	}

	if _, exists := s.byID[g.ID]; exists {
		return ErrDuplicateGameID
	}
	if g.JoinCode != "" {
		if _, exists := s.byJoinCode[g.JoinCode]; exists {
			return ErrDuplicateJoinCode
		}
	}

	s.byID[g.ID] = cloneGame(g)
	if g.JoinCode != "" {
		s.byJoinCode[g.JoinCode] = g.ID
	}
	return nil
}

func (s *Store) GetByID(id string) (*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, ok := s.byID[id]
	if !ok {
		return nil, ErrGameNotFound
	}
	return cloneGame(g), nil
}

func (s *Store) GetByJoinCode(code string) (*game.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byJoinCode[code]
	if !ok {
		return nil, ErrJoinCodeNotFound
	}

	g, ok := s.byID[id]
	if !ok {
		return nil, ErrGameNotFound
	}
	return cloneGame(g), nil
}

func (s *Store) Update(g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if g == nil {
		return errors.New("nil game")
	}
	if g.ID == "" {
		return errors.New("empty game id")
	}

	if _, ok := s.byID[g.ID]; !ok {
		return ErrGameNotFound
	}

	s.byID[g.ID] = cloneGame(g)
	return nil
}

var _ repository.GameRepository = (*Store)(nil)
