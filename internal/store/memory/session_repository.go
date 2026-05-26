package memory

import (
	"sync"

	domainsession "github.com/MehrshadFb/xo-grpc/internal/domain/session"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
)

type SessionRepository struct {
	mu      sync.RWMutex
	byToken map[string]domainsession.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		byToken: make(map[string]domainsession.Session),
	}
}

var _ repository.SessionRepository = (*SessionRepository)(nil)

func (r *SessionRepository) Create(s domainsession.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byToken[s.Token] = s
	return nil
}

func (r *SessionRepository) Get(token string) (domainsession.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.byToken[token]
	if !ok {
		return domainsession.Session{}, session.ErrSessionNotFound
	}

	return s, nil
}
