package lobby

import (
	"errors"
	"strings"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
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

func (s *Service) CreateGame(displayName string) (*CreateGameResult, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, errors.New("display name is required")
	}

	// 1) Generate IDs
	gameID, err := newGameID()
	if err != nil {
		return nil, err
	}

	joinCode, err := newJoinCode()
	if err != nil {
		return nil, err
	}

	playerID, err := newPlayerID()
	if err != nil {
		return nil, err
	}

	// 2) Create domain game
	g := game.NewGame(gameID, joinCode)

	// 3) Set Player X (creator)
	g.SetPlayerX(playerID, displayName)

	// 4) Persist game
	if err := s.store.Create(g); err != nil {
		return nil, err
	}

	// 5) Create session token for Player X
	token, err := s.sessions.Create(gameID, playerID, game.MarkX)
	if err != nil {
		return nil, err
	}

	return &CreateGameResult{
		Game:        g,
		PlayerToken: token,
		PlayerID:    playerID,
	}, nil
}