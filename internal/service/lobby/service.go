package lobby

import (
	"strings"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/realtime"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
)

type Service struct {
	store    *memory.Store
	sessions *session.Manager
	hub      *realtime.Hub
}

func NewService(store *memory.Store, sessions *session.Manager, hub *realtime.Hub) *Service {
	return &Service{
		store:    store,
		sessions: sessions,
		hub:      hub,
	}
}

func (s *Service) CreateGame(displayName string) (*CreateGameResult, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, ErrEmptyDisplayName
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

func (s *Service) JoinGame(joinCode, displayName string) (*JoinGameResult, error) {
	joinCode = strings.TrimSpace(joinCode)
	displayName = strings.TrimSpace(displayName)

	if displayName == "" {
		return nil, ErrEmptyDisplayName
	}
	if joinCode == "" {
		return nil, ErrInvalidJoinCode
	}

	g, err := s.store.GetByJoinCode(joinCode)
	if err != nil {
		if err == memory.ErrJoinCodeNotFound {
			return nil, ErrInvalidJoinCode
		}
		return nil, err
	}

	if g.Status != game.StatusWaiting {
		return nil, ErrGameNotWaiting
	}
	if g.PlayerO != nil {
		return nil, ErrGameFull
	}

	playerID, err := newPlayerID()
	if err != nil {
		return nil, err
	}

	g.SetPlayerO(playerID, displayName)

	if err := g.Start(); err != nil {
		return nil, err
	}

	if err := s.store.Update(g); err != nil {
		return nil, err
	}
	
	if s.hub != nil {
		s.hub.Publish(g.ID, realtime.Event{
			Type: realtime.EventTypePlayerJoined,
			Game: g,
		})
	}

	token, err := s.sessions.Create(g.ID, playerID, game.MarkO)
	if err != nil {
		return nil, err
	}

	return &JoinGameResult{
		Game:        g,
		PlayerToken: token,
		PlayerID:    playerID,
	}, nil
}