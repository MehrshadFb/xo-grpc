package game

import "testing"

func TestNewGame_InitialState(t *testing.T) {
	g := NewGame("g1", "CODE")

	if g.Status != StatusWaiting {
		t.Fatalf("expected StatusWaiting, got %v", g.Status)
	}
	if g.NextTurn != MarkX {
		t.Fatalf("expected NextTurn MarkX, got %v", g.NextTurn)
	}
	if g.Version != 1 {
		t.Fatalf("expected Version 1, got %d", g.Version)
	}
	for i, c := range g.Board {
		if c != MarkEmpty {
			t.Fatalf("expected board[%d] empty, got %v", i, c)
		}
	}
}

func TestStart_TransitionsToInProgress(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")

	if err := g.Start(); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}

	if g.Status != StatusInProgress {
		t.Fatalf("expected StatusInProgress, got %v", g.Status)
	}
	if g.NextTurn != MarkX {
		t.Fatalf("expected NextTurn MarkX, got %v", g.NextTurn)
	}
	if g.Version != 2 {
		t.Fatalf("expected Version 2 after Start, got %d", g.Version)
	}
}

func TestApplyMove_BasicTurnFlow(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	// X plays center
	if err := g.ApplyMove(MarkX, 4); err != nil {
		t.Fatalf("ApplyMove(X,4) unexpected error: %v", err)
	}
	if g.Board[4] != MarkX {
		t.Fatalf("expected board[4]=X, got %v", g.Board[4])
	}
	if g.NextTurn != MarkO {
		t.Fatalf("expected NextTurn O, got %v", g.NextTurn)
	}
	if g.MoveNumber != 1 {
		t.Fatalf("expected MoveNumber 1, got %d", g.MoveNumber)
	}

	// O plays corner
	if err := g.ApplyMove(MarkO, 0); err != nil {
		t.Fatalf("ApplyMove(O,0) unexpected error: %v", err)
	}
	if g.Board[0] != MarkO {
		t.Fatalf("expected board[0]=O, got %v", g.Board[0])
	}
	if g.NextTurn != MarkX {
		t.Fatalf("expected NextTurn X, got %v", g.NextTurn)
	}
	if g.MoveNumber != 2 {
		t.Fatalf("expected MoveNumber 2, got %d", g.MoveNumber)
	}
}

func TestApplyMove_InvalidCellIndex(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	if err := g.ApplyMove(MarkX, -1); err != ErrInvalidCellIndex {
		t.Fatalf("expected ErrInvalidCellIndex, got %v", err)
	}
	if err := g.ApplyMove(MarkX, 9); err != ErrInvalidCellIndex {
		t.Fatalf("expected ErrInvalidCellIndex, got %v", err)
	}
}

func TestApplyMove_CellOccupied(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	if err := g.ApplyMove(MarkX, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := g.ApplyMove(MarkO, 1); err != ErrCellOccupied {
		t.Fatalf("expected ErrCellOccupied, got %v", err)
	}
}

func TestApplyMove_WrongTurn(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	// X's turn first; O tries to play
	if err := g.ApplyMove(MarkO, 0); err != ErrNotPlayersTurn {
		t.Fatalf("expected ErrNotPlayersTurn, got %v", err)
	}
}

func TestApplyMove_GameNotInProgress(t *testing.T) {
	g := NewGame("g1", "CODE")
	// Note: we did NOT call Start()

	if err := g.ApplyMove(MarkX, 0); err != ErrGameNotInProgress {
		t.Fatalf("expected ErrGameNotInProgress, got %v", err)
	}
}

func TestApplyMove_XWinsTopRow(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	// X:0, O:3, X:1, O:4, X:2 => X wins
	moves := []struct {
		mark Mark
		cell int
	}{
		{MarkX, 0},
		{MarkO, 3},
		{MarkX, 1},
		{MarkO, 4},
		{MarkX, 2},
	}

	for _, m := range moves {
		if err := g.ApplyMove(m.mark, m.cell); err != nil {
			t.Fatalf("unexpected error applying %v to %d: %v", m.mark, m.cell, err)
		}
	}

	if g.Status != StatusFinished {
		t.Fatalf("expected StatusFinished, got %v", g.Status)
	}
	if g.Winner != MarkX {
		t.Fatalf("expected Winner X, got %v", g.Winner)
	}
	if g.IsDraw {
		t.Fatalf("expected IsDraw false")
	}
}

func TestApplyMove_Draw(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	// A known draw sequence (no 3-in-a-row):
	// X O X
	// X O O
	// O X X
	moves := []struct {
		mark Mark
		cell int
	}{
		{MarkX, 0},
		{MarkO, 1},
		{MarkX, 2},
		{MarkO, 4},
		{MarkX, 3},
		{MarkO, 5},
		{MarkX, 7},
		{MarkO, 6},
		{MarkX, 8},
	}

	for _, m := range moves {
		if err := g.ApplyMove(m.mark, m.cell); err != nil {
			t.Fatalf("unexpected error applying %v to %d: %v", m.mark, m.cell, err)
		}
	}

	if g.Status != StatusFinished {
		t.Fatalf("expected StatusFinished, got %v", g.Status)
	}
	if !g.IsDraw {
		t.Fatalf("expected IsDraw true")
	}
	if g.Winner != MarkEmpty {
		t.Fatalf("expected Winner empty on draw, got %v", g.Winner)
	}
}

func TestApplyMove_NoMovesAfterFinished(t *testing.T) {
	g := NewGame("g1", "CODE")
	g.SetPlayerO("p2", "bob")
	_ = g.Start()

	// Force a quick win for X
	_ = g.ApplyMove(MarkX, 0)
	_ = g.ApplyMove(MarkO, 3)
	_ = g.ApplyMove(MarkX, 1)
	_ = g.ApplyMove(MarkO, 4)
	_ = g.ApplyMove(MarkX, 2)

	if g.Status != StatusFinished {
		t.Fatalf("expected finished game")
	}

	if err := g.ApplyMove(MarkO, 8); err != ErrGameFinished {
		t.Fatalf("expected ErrGameFinished, got %v", err)
	}
}
