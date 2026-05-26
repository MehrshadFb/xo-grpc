package session_test

import (
	"testing"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/service/session"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
)

func TestManager_CreateAndGet(t *testing.T) {
	repo := memory.NewSessionRepository()
	m := session.NewManager(repo)

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
	repo := memory.NewSessionRepository()
	m := session.NewManager(repo)

	_, err := m.Get("missing")
	if err != session.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestManager_ValidateGame_Mismatch(t *testing.T) {
	repo := memory.NewSessionRepository()
	m := session.NewManager(repo)

	token, err := m.Create("game1", "player1", game.MarkX)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.ValidateGame(token, "game2")
	if err != session.ErrSessionGameMismatch {
		t.Fatalf("expected ErrSessionGameMismatch, got %v", err)
	}
}
