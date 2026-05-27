package health

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Ready(ctx context.Context) bool {
	if s.db == nil {
		return true
	}

	return s.db.Ping(ctx) == nil
}
