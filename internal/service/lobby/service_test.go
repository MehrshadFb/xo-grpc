package lobby

import (
	"testing"

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