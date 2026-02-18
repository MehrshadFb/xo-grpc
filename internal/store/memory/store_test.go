package memory

import (
	"testing"

	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

func TestStore_CreateAndGetByID(t *testing.T) {
	s := NewStore()

	g := game.NewGame("g1", "CODE1")
	if err := s.Create(g); err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	got, err := s.GetByID("g1")
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil game")
	}
	if got.ID != "g1" {
		t.Fatalf("expected ID g1, got %q", got.ID)
	}
	if got.JoinCode != "CODE1" {
		t.Fatalf("expected JoinCode CODE1, got %q", got.JoinCode)
	}
}

func TestStore_GetByID_NotFound(t *testing.T) {
	s := NewStore()

	_, err := s.GetByID("missing")
	if err != ErrGameNotFound {
		t.Fatalf("expected ErrGameNotFound, got %v", err)
	}
}

func TestStore_Create_DuplicateGameID(t *testing.T) {
	s := NewStore()

	if err := s.Create(game.NewGame("g1", "CODE1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Create(game.NewGame("g1", "CODE2")); err != ErrDuplicateGameID {
		t.Fatalf("expected ErrDuplicateGameID, got %v", err)
	}
}

func TestStore_Create_DuplicateJoinCode(t *testing.T) {
	s := NewStore()

	if err := s.Create(game.NewGame("g1", "CODE1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Create(game.NewGame("g2", "CODE1")); err != ErrDuplicateJoinCode {
		t.Fatalf("expected ErrDuplicateJoinCode, got %v", err)
	}
}

func TestStore_GetByJoinCode(t *testing.T) {
	s := NewStore()

	if err := s.Create(game.NewGame("g1", "CODE1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := s.GetByJoinCode("CODE1")
	if err != nil {
		t.Fatalf("GetByJoinCode() unexpected error: %v", err)
	}
	if got.ID != "g1" {
		t.Fatalf("expected ID g1, got %q", got.ID)
	}
}

func TestStore_GetByJoinCode_NotFound(t *testing.T) {
	s := NewStore()

	_, err := s.GetByJoinCode("NOPE")
	if err != ErrJoinCodeNotFound {
		t.Fatalf("expected ErrJoinCodeNotFound, got %v", err)
	}
}

func TestStore_Update(t *testing.T) {
	s := NewStore()

	g := game.NewGame("g1", "CODE1")
	if err := s.Create(g); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate the game and update
	_ = g.Start()
	if err := s.Update(g); err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	got, err := s.GetByID("g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != game.StatusInProgress {
		t.Fatalf("expected StatusInProgress, got %v", got.Status)
	}
}

func TestStore_Update_NotFound(t *testing.T) {
	s := NewStore()

	g := game.NewGame("missing", "CODEX")
	if err := s.Update(g); err != ErrGameNotFound {
		t.Fatalf("expected ErrGameNotFound, got %v", err)
	}
}
