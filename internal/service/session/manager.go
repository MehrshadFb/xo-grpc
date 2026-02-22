package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

type Manager struct {
	mu sync.RWMutex
	byToken map[string]Session
}

func NewManager() *Manager {
	return &Manager{
		byToken: make(map[string]Session),
	}
}

func (m *Manager) Create(gameID, playerID string, mark game.Mark) (string, error) {
	if gameID == "" || playerID == "" {
		return "", errors.New("gameID and playerID are required")
	}
	if mark != game.MarkX && mark != game.MarkO {
		return "", errors.New("invalid mark")
	}

	token, err := randomToken(24) // 24 bytes -> 48 hex chars
	if err != nil {
		return "", err
	}

	s := Session{
		Token:    token,
		GameID:   gameID,
		PlayerID: playerID,
		Mark:     mark,
	}

	m.mu.Lock()
	m.byToken[token] = s
	m.mu.Unlock()

	return token, nil
}

func (m *Manager) Get(token string) (Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.byToken[token]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return s, nil
}

func (m *Manager) ValidateGame(token, gameID string) (Session, error) {
	s, err := m.Get(token)
	if err != nil {
		return Session{}, err
	}
	if s.GameID != gameID {
		return Session{}, ErrSessionGameMismatch
	}
	return s, nil
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
