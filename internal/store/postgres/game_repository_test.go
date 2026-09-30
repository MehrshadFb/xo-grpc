package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MehrshadFb/xo-grpc/internal/database"
	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
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

	applyAndPersistMove := func(mark domaingame.Mark, cell int) *domaingame.Game {
		t.Helper()

		game, err := repo.GetByID("game1")
		if err != nil {
			t.Fatalf("GetByID before move: %v", err)
		}
		if err := game.ApplyMove(mark, cell); err != nil {
			t.Fatalf("ApplyMove %v at %d: %v", mark, cell, err)
		}
		if err := repo.Update(game); err != nil {
			t.Fatalf("Update after move %v at %d: %v", mark, cell, err)
		}

		return game
	}

	applyAndPersistMove(domaingame.MarkO, 0)
	applyAndPersistMove(domaingame.MarkX, 3)
	applyAndPersistMove(domaingame.MarkO, 1)
	latest = applyAndPersistMove(domaingame.MarkX, 5)
	if latest.Status != domaingame.StatusFinished || latest.XWins != 1 {
		t.Fatalf("expected finished game with X score 1, got status=%v X=%d", latest.Status, latest.XWins)
	}
	if _, err := latest.RequestRematch(domaingame.MarkX); err != nil {
		t.Fatalf("RequestRematch X: %v", err)
	}

	if err := repo.Update(latest); err != nil {
		t.Fatalf("Update after rematch request: %v", err)
	}

	updated, err = repo.GetByID("game1")
	if err != nil {
		t.Fatalf("GetByID after rematch request: %v", err)
	}

	if updated.XWins != 1 || updated.OWins != 0 || updated.Draws != 0 {
		t.Fatalf("unexpected persisted score X=%d O=%d D=%d", updated.XWins, updated.OWins, updated.Draws)
	}
	if !updated.RematchXRequested || updated.RematchORequested {
		t.Fatalf("unexpected persisted rematch flags X=%v O=%v", updated.RematchXRequested, updated.RematchORequested)
	}
	if updated.RoundNumber != 1 {
		t.Fatalf("expected round 1 before restart, got %d", updated.RoundNumber)
	}
}

func TestGameRepository_DeleteStale(t *testing.T) {
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

	games := []struct {
		id     string
		status string
		idle   time.Duration
	}{
		{"waiting-old", "WAITING", 2 * time.Hour},
		{"waiting-new", "WAITING", 10 * time.Minute},
		{"playing-old", "IN_PROGRESS", 48 * time.Hour},
		{"playing-new", "IN_PROGRESS", 2 * time.Hour},
		{"finished-old", "FINISHED", 8 * 24 * time.Hour},
		{"finished-new", "FINISHED", 2 * 24 * time.Hour},
	}

	for i, tc := range games {
		g := domaingame.NewGame(tc.id, fmt.Sprintf("CODE%d", i))
		g.SetPlayerX("player-"+tc.id, "Alice")
		if err := repo.Create(g); err != nil {
			t.Fatalf("create %s: %v", tc.id, err)
		}

		_, err := pool.Exec(ctx, `
			UPDATE games SET status = $2, updated_at = NOW() - $3::interval WHERE id = $1
		`, tc.id, tc.status, fmt.Sprintf("%d seconds", int(tc.idle.Seconds())))
		if err != nil {
			t.Fatalf("age %s: %v", tc.id, err)
		}
	}

	now := time.Now()
	deleted, err := repo.DeleteStale(ctx, now.Add(-time.Hour), now.Add(-24*time.Hour), now.Add(-7*24*time.Hour))
	if err != nil {
		t.Fatalf("DeleteStale error: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 deleted games, got %d", deleted)
	}

	for _, tc := range games {
		_, err := repo.GetByID(tc.id)
		shouldExist := !strings.HasSuffix(tc.id, "-old")

		if shouldExist && err != nil {
			t.Fatalf("expected %s to remain, got error %v", tc.id, err)
		}
		if !shouldExist && !errors.Is(err, memory.ErrGameNotFound) {
			t.Fatalf("expected %s to be deleted, got error %v", tc.id, err)
		}
	}

	var orphanPlayers int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM players WHERE game_id LIKE '%-old'`).Scan(&orphanPlayers); err != nil {
		t.Fatalf("count players: %v", err)
	}
	if orphanPlayers != 0 {
		t.Fatalf("expected players of deleted games to cascade, found %d", orphanPlayers)
	}
}
