package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/MehrshadFb/xo-grpc/internal/database"
	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
	domainsession "github.com/MehrshadFb/xo-grpc/internal/domain/session"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
)

func TestSessionRepository_CreateAndGet(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping postgres integration test")
	}

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `
		TRUNCATE TABLE sessions, players, games RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	// create required game row
	_, err = pool.Exec(ctx, `
		INSERT INTO games (
			id,
			join_code,
			status,
			board,
			next_turn,
			winner,
			is_draw,
			move_number,
			version
		)
		VALUES (
			'game1',
			'CODE1',
			'WAITING',
			'["EMPTY","EMPTY","EMPTY","EMPTY","EMPTY","EMPTY","EMPTY","EMPTY","EMPTY"]',
			'X',
			'EMPTY',
			false,
			0,
			1
		)
	`)
	if err != nil {
		t.Fatalf("insert game: %v", err)
	}

	// create required player row
	_, err = pool.Exec(ctx, `
		INSERT INTO players (
			id,
			game_id,
			display_name,
			mark
		)
		VALUES (
			'player1',
			'game1',
			'Alice',
			'X'
		)
	`)
	if err != nil {
		t.Fatalf("insert player: %v", err)
	}

	repo := NewSessionRepository(pool)

	expected := domainsession.Session{
		Token:    "token123",
		GameID:   "game1",
		PlayerID: "player1",
		Mark:     game.MarkX,
	}

	if err := repo.Create(expected); err != nil {
		t.Fatalf("Create: %v", err)
	}

	actual, err := repo.Get("token123")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if actual.Token != expected.Token {
		t.Fatalf("expected token %q, got %q", expected.Token, actual.Token)
	}

	if actual.GameID != expected.GameID {
		t.Fatalf("expected gameID %q, got %q", expected.GameID, actual.GameID)
	}

	if actual.PlayerID != expected.PlayerID {
		t.Fatalf("expected playerID %q, got %q", expected.PlayerID, actual.PlayerID)
	}

	if actual.Mark != expected.Mark {
		t.Fatalf("expected mark %v, got %v", expected.Mark, actual.Mark)
	}
}

func TestSessionRepository_Get_NotFound(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping postgres integration test")
	}

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	repo := NewSessionRepository(pool)

	_, err = repo.Get("missing")
	if err != session.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}
