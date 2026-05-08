package gamesvc

import (
	"testing"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/lobby"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
)

func setupStartedGame(t *testing.T) (*Service, *lobby.CreateGameResult, *lobby.JoinGameResult) {
	t.Helper()

	store := memory.NewStore()
	sessions := session.NewManager()

	lobbySvc := lobby.NewService(store, sessions)
	gameSvc := NewService(store, sessions, nil)

	created, err := lobbySvc.CreateGame("Alice")
	if err != nil {
		t.Fatalf("CreateGame error: %v", err)
	}

	joined, err := lobbySvc.JoinGame(created.Game.JoinCode, "Bob")
	if err != nil {
		t.Fatalf("JoinGame error: %v", err)
	}

	return gameSvc, created, joined
}

func TestGetState(t *testing.T) {
	gameSvc, created, _ := setupStartedGame(t)

	result, err := gameSvc.GetState(created.Game.ID, created.PlayerToken)
	if err != nil {
		t.Fatalf("GetState error: %v", err)
	}

	if result.Game.ID != created.Game.ID {
		t.Fatalf("expected game ID %q, got %q", created.Game.ID, result.Game.ID)
	}
	if result.Game.Status != domaingame.StatusInProgress {
		t.Fatalf("expected StatusInProgress, got %v", result.Game.Status)
	}
}

func TestMakeMove_XMovesFirst(t *testing.T) {
	gameSvc, created, _ := setupStartedGame(t)

	result, err := gameSvc.MakeMove(created.Game.ID, created.PlayerToken, 4)
	if err != nil {
		t.Fatalf("MakeMove error: %v", err)
	}

	if result.Game.Board[4] != domaingame.MarkX {
		t.Fatalf("expected board[4] to be X, got %v", result.Game.Board[4])
	}
	if result.Game.NextTurn != domaingame.MarkO {
		t.Fatalf("expected next turn O, got %v", result.Game.NextTurn)
	}
}

func TestMakeMove_RejectsWrongTurn(t *testing.T) {
	gameSvc, created, joined := setupStartedGame(t)

	// O tries to move first, but X starts.
	_, err := gameSvc.MakeMove(created.Game.ID, joined.PlayerToken, 0)
	if err != domaingame.ErrNotPlayersTurn {
		t.Fatalf("expected ErrNotPlayersTurn, got %v", err)
	}
}

func TestMakeMove_RejectsInvalidToken(t *testing.T) {
	gameSvc, created, _ := setupStartedGame(t)

	_, err := gameSvc.MakeMove(created.Game.ID, "bad-token", 0)
	if err != session.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestMakeMove_OMovesAfterX(t *testing.T) {
	gameSvc, created, joined := setupStartedGame(t)

	if _, err := gameSvc.MakeMove(created.Game.ID, created.PlayerToken, 0); err != nil {
		t.Fatalf("X MakeMove error: %v", err)
	}

	result, err := gameSvc.MakeMove(created.Game.ID, joined.PlayerToken, 1)
	if err != nil {
		t.Fatalf("O MakeMove error: %v", err)
	}

	if result.Game.Board[1] != domaingame.MarkO {
		t.Fatalf("expected board[1] to be O, got %v", result.Game.Board[1])
	}
	if result.Game.NextTurn != domaingame.MarkX {
		t.Fatalf("expected next turn X, got %v", result.Game.NextTurn)
	}
}