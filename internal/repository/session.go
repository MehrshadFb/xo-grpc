package repository

import domainsession "github.com/MehrshadFb/xo-grpc/internal/domain/session"

type SessionRepository interface {
	Create(s domainsession.Session) error
	Get(token string) (domainsession.Session, error)
}
