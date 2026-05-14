package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxAttempts = 10
	defaultRetryDelay  = time.Second
)

func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return NewPostgresPoolWithRetry(ctx, databaseURL, defaultMaxAttempts, defaultRetryDelay)
}

func NewPostgresPoolWithRetry(ctx context.Context, databaseURL string, maxAttempts int, retryDelay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err != nil {
			lastErr = fmt.Errorf("create postgres pool: %w", err)
		} else {
			if err := pool.Ping(ctx); err == nil {
				return pool, nil
			} else {
				lastErr = fmt.Errorf("ping postgres: %w", err)
				pool.Close()
			}
		}

		if attempt == maxAttempts {
			break
		}

		// wait for context to be done or for the retry delay to expire
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
		}
	}

	return nil, fmt.Errorf("connect to postgres after %d attempts: %w", maxAttempts, lastErr)
}
