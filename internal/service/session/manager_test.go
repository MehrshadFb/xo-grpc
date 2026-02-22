package session

import (
	"testing"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

func TestManager_CreateAndGet(t *testing.T) {
	m := NewManager()

	token, err := m.Create("game1", "player1", game.MarkX)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	s, err := m.Get(token)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if s.GameID != "game1" || s.PlayerID != "player1" || s.Mark != game.MarkX {
		t.Fatalf("unexpected session: %+v", s)
	}
}

func TestManager_Get_NotFound(t *testing.T) {
	m := NewManager()

	_, err := m.Get("missing")
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestManager_ValidateGame_Mismatch(t *testing.T) {
	m := NewManager()

	token, err := m.Create("game1", "player1", game.MarkX)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.ValidateGame(token, "game2")
	if err != ErrSessionGameMismatch {
		t.Fatalf("expected ErrSessionGameMismatch, got %v", err)
	}
}
