package lobby

import (
	"testing"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
)

func TestCreateGame(t *testing.T) {
	store := memory.NewStore()
	sessions := session.NewManager()

	service := NewService(store, sessions)

	result, err := service.CreateGame("Alice")
	if err != nil {
		t.Fatalf("CreateGame error: %v", err)
	}

	if result.Game == nil {
		t.Fatalf("expected game, got nil")
	}

	if result.Game.PlayerX == nil {
		t.Fatalf("expected PlayerX to be set")
	}

	if result.Game.PlayerX.DisplayName != "Alice" {
		t.Fatalf("expected display name Alice, got %q", result.Game.PlayerX.DisplayName)
	}

	if result.PlayerToken == "" {
		t.Fatalf("expected non-empty player token")
	}

	// Ensure store has the game
	stored, err := store.GetByID(result.Game.ID)
	if err != nil {
		t.Fatalf("expected game in store, got error %v", err)
	}
	if stored.ID != result.Game.ID {
		t.Fatalf("stored game mismatch")
	}

	// Ensure session exists
	session, err := sessions.Get(result.PlayerToken)
	if err != nil {
		t.Fatalf("expected session, got error %v", err)
	}
	if session.GameID != result.Game.ID {
		t.Fatalf("session gameID mismatch")
	}
}

func TestJoinGame(t *testing.T) {
	store := memory.NewStore()
	sessions := session.NewManager()
	service := NewService(store, sessions)

	created, err := service.CreateGame("Alice")
	if err != nil {
		t.Fatalf("CreateGame error: %v", err)
	}

	joined, err := service.JoinGame(created.Game.JoinCode, "Bob")
	if err != nil {
		t.Fatalf("JoinGame error: %v", err)
	}

	if joined.Game == nil {
		t.Fatalf("expected game, got nil")
	}
	if joined.Game.PlayerO == nil {
		t.Fatalf("expected PlayerO to be set")
	}
	if joined.Game.PlayerO.DisplayName != "Bob" {
		t.Fatalf("expected PlayerO display name Bob, got %q", joined.Game.PlayerO.DisplayName)
	}
	if joined.Game.Status != game.StatusInProgress {
		t.Fatalf("expected game status in progress, got %v", joined.Game.Status)
	}
	if joined.PlayerToken == "" {
		t.Fatalf("expected non-empty player token")
	}

	sess, err := sessions.Get(joined.PlayerToken)
	if err != nil {
		t.Fatalf("expected session for joined player, got error %v", err)
	}
	if sess.GameID != joined.Game.ID {
		t.Fatalf("session gameID mismatch")
	}
	if sess.Mark != game.MarkO {
		t.Fatalf("expected joined player mark O, got %v", sess.Mark)
	}
}

func TestJoinGame_InvalidJoinCode(t *testing.T) {
	store := memory.NewStore()
	sessions := session.NewManager()
	service := NewService(store, sessions)

	_, err := service.JoinGame("NOPE", "Bob")
	if err != ErrInvalidJoinCode {
		t.Fatalf("expected ErrInvalidJoinCode, got %v", err)
	}
}

func TestJoinGame_GameAlreadyFull(t *testing.T) {
	store := memory.NewStore()
	sessions := session.NewManager()
	service := NewService(store, sessions)

	created, err := service.CreateGame("Alice")
	if err != nil {
		t.Fatalf("CreateGame error: %v", err)
	}

	_, err = service.JoinGame(created.Game.JoinCode, "Bob")
	if err != nil {
		t.Fatalf("first JoinGame error: %v", err)
	}

	_, err = service.JoinGame(created.Game.JoinCode, "Charlie")
	if err != ErrGameNotWaiting && err != ErrGameFull {
		t.Fatalf("expected ErrGameNotWaiting or ErrGameFull, got %v", err)
	}
}

func TestJoinGame_EmptyDisplayName(t *testing.T) {
	store := memory.NewStore()
	sessions := session.NewManager()
	service := NewService(store, sessions)

	created, err := service.CreateGame("Alice")
	if err != nil {
		t.Fatalf("CreateGame error: %v", err)
	}

	_, err = service.JoinGame(created.Game.JoinCode, "   ")
	if err != ErrEmptyDisplayName {
		t.Fatalf("expected ErrEmptyDisplayName, got %v", err)
	}
}