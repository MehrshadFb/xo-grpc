package postgres

import (
	"context"
	"errors"
	"fmt"

	domainsession "github.com/MehrshadFb/xo-grpc/internal/domain/session"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool: pool,
	}
}

var _ repository.SessionRepository = (*SessionRepository)(nil)

func (r *SessionRepository) Create(s domainsession.Session) error {
	ctx := context.Background()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (
			token,
			game_id,
			player_id,
			mark
		)
		VALUES ($1, $2, $3, $4)
	`,
		s.Token,
		s.GameID,
		s.PlayerID,
		markToString(s.Mark),
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

func (r *SessionRepository) Get(token string) (domainsession.Session, error) {
	ctx := context.Background()

	var (
		s    domainsession.Session
		mark string
	)

	err := r.pool.QueryRow(ctx, `
		SELECT token, game_id, player_id, mark
		FROM sessions
		WHERE token = $1
	`, token).Scan(
		&s.Token,
		&s.GameID,
		&s.PlayerID,
		&mark,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainsession.Session{}, session.ErrSessionNotFound
		}

		return domainsession.Session{}, fmt.Errorf("select session: %w", err)
	}

	s.Mark = stringToMark(mark)

	return s, nil
}
