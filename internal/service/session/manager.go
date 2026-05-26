package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
	domainsession "github.com/MehrshadFb/xo-grpc/internal/domain/session"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
)

type Manager struct {
	repo repository.SessionRepository
}

func NewManager(repo repository.SessionRepository) *Manager {
	return &Manager{
		repo: repo,
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

	s := domainsession.Session{
		Token:    token,
		GameID:   gameID,
		PlayerID: playerID,
		Mark:     mark,
	}

	if err := m.repo.Create(s); err != nil {
		return "", err
	}

	return token, nil
}

func (m *Manager) Get(token string) (domainsession.Session, error) {
	return m.repo.Get(token)
}

func (m *Manager) ValidateGame(token, gameID string) (domainsession.Session, error) {
	s, err := m.Get(token)
	if err != nil {
		return domainsession.Session{}, err
	}
	if s.GameID != gameID {
		return domainsession.Session{}, ErrSessionGameMismatch
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
