package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/MehrshadFb/xo-grpc/internal/database"
	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
)

func TestGameRepository_CreateGetAndUpdate(t *testing.T) {
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

	repo := NewGameRepository(pool)
	sessionRepo := NewSessionRepository(pool)
	sessions := session.NewManager(sessionRepo)

	g := domaingame.NewGame("game1", "CODE1")
	g.SetPlayerX("player-x", "Alice")

	if err := repo.Create(g); err != nil {
		t.Fatalf("Create: %v", err)
	}

	byID, err := repo.GetByID("game1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if byID.ID != "game1" {
		t.Fatalf("expected game1, got %q", byID.ID)
	}
	if byID.JoinCode != "CODE1" {
		t.Fatalf("expected CODE1, got %q", byID.JoinCode)
	}
	if byID.PlayerX == nil || byID.PlayerX.DisplayName != "Alice" {
		t.Fatalf("expected PlayerX Alice, got %+v", byID.PlayerX)
	}

	byCode, err := repo.GetByJoinCode("CODE1")
	if err != nil {
		t.Fatalf("GetByJoinCode: %v", err)
	}
	if byCode.ID != "game1" {
		t.Fatalf("expected game1 by join code, got %q", byCode.ID)
	}

	playerXToken, err := sessions.Create("game1", "player-x", domaingame.MarkX)
	if err != nil {
		t.Fatalf("Create player X session: %v", err)
	}

	latest, err := repo.GetByID("game1")
	if err != nil {
		t.Fatalf("GetByID latest: %v", err)
	}

	latest.SetPlayerO("player-o", "Bob")
	if err := latest.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := repo.Update(latest); err != nil {
		t.Fatalf("Update after Start: %v", err)
	}

	if _, err := sessions.Get(playerXToken); err != nil {
		t.Fatalf("expected player X session to survive update: %v", err)
	}

	latest, err = repo.GetByID("game1")
	if err != nil {
		t.Fatalf("GetByID after start: %v", err)
	}

	if err := latest.ApplyMove(domaingame.MarkX, 4); err != nil {
		t.Fatalf("ApplyMove: %v", err)
	}

	if err := repo.Update(latest); err != nil {
		t.Fatalf("Update after move: %v", err)
	}

	updated, err := repo.GetByID("game1")
	if err != nil {
		t.Fatalf("GetByID updated: %v", err)
	}

	if updated.PlayerO == nil || updated.PlayerO.DisplayName != "Bob" {
		t.Fatalf("expected PlayerO Bob, got %+v", updated.PlayerO)
	}
	if updated.Status != domaingame.StatusInProgress {
		t.Fatalf("expected in progress, got %v", updated.Status)
	}
	if updated.Board[4] != domaingame.MarkX {
		t.Fatalf("expected board[4] X, got %v", updated.Board[4])
	}
	if updated.NextTurn != domaingame.MarkO {
		t.Fatalf("expected next turn O, got %v", updated.NextTurn)
	}
	if updated.Version != latest.Version {
		t.Fatalf("expected version %d, got %d", latest.Version, updated.Version)
	}
}
