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
	sessionRepo := memory.NewSessionRepository()
	sessions := session.NewManager(sessionRepo)

	lobbySvc := lobby.NewService(store, sessions, nil)
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

func TestRequestRematch_RequiresBothPlayersAndKeepsScore(t *testing.T) {
	gameSvc, created, joined := setupStartedGame(t)

	moves := []struct {
		token string
		cell  int
	}{
		{created.PlayerToken, 0},
		{joined.PlayerToken, 3},
		{created.PlayerToken, 1},
		{joined.PlayerToken, 4},
		{created.PlayerToken, 2},
	}

	for _, move := range moves {
		if _, err := gameSvc.MakeMove(created.Game.ID, move.token, move.cell); err != nil {
			t.Fatalf("MakeMove: %v", err)
		}
	}

	xRequested, err := gameSvc.RequestRematch(created.Game.ID, created.PlayerToken)
	if err != nil {
		t.Fatalf("RequestRematch X: %v", err)
	}
	if xRequested.Started {
		t.Fatalf("expected first rematch request to wait for opponent")
	}
	if !xRequested.Changed {
		t.Fatalf("expected first rematch request to change state")
	}
	if !xRequested.Game.RematchXRequested || xRequested.Game.RematchORequested {
		t.Fatalf("unexpected rematch flags X=%v O=%v", xRequested.Game.RematchXRequested, xRequested.Game.RematchORequested)
	}

	duplicate, err := gameSvc.RequestRematch(created.Game.ID, created.PlayerToken)
	if err != nil {
		t.Fatalf("duplicate RequestRematch X: %v", err)
	}
	if duplicate.Changed {
		t.Fatalf("duplicate rematch request should be idempotent")
	}

	started, err := gameSvc.RequestRematch(created.Game.ID, joined.PlayerToken)
	if err != nil {
		t.Fatalf("RequestRematch O: %v", err)
	}
	if !started.Started {
		t.Fatalf("expected second rematch request to start next round")
	}
	if started.Game.Status != domaingame.StatusInProgress {
		t.Fatalf("expected in progress, got %v", started.Game.Status)
	}
	if started.Game.RoundNumber != 2 {
		t.Fatalf("expected round 2, got %d", started.Game.RoundNumber)
	}
	if started.Game.XWins != 1 || started.Game.OWins != 0 || started.Game.Draws != 0 {
		t.Fatalf("unexpected score X=%d O=%d D=%d", started.Game.XWins, started.Game.OWins, started.Game.Draws)
	}
}

func TestRequestRematch_RejectsBeforeGameFinished(t *testing.T) {
	gameSvc, created, _ := setupStartedGame(t)

	_, err := gameSvc.RequestRematch(created.Game.ID, created.PlayerToken)
	if err != domaingame.ErrGameNotFinished {
		t.Fatalf("expected ErrGameNotFinished, got %v", err)
	}
}
