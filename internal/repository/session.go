package repository

import "github.com/MehrshadFb/xo-grpc/internal/service/session"

type SessionRepository interface {
	Create(session.Session)
	Get(token string) (session.Session, bool)
}
