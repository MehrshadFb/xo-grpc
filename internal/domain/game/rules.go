package game

func (g *Game) Start() error {
	if g.Status == StatusFinished {
		return ErrGameFinished
	}
	// If we want to allow restarting an aborted game later, we can adjust this.
	if g.Status != StatusWaiting {
		// Already started - treat as no-op for simplicity
		return nil
	}
	if g.PlayerO == nil {
		return ErrPlayerOMissing
	}

	g.Status = StatusInProgress
	g.NextTurn = MarkX // X always starts
	g.Version++
	return nil
}

func (g *Game) ApplyMove(player Mark, cell int) error {
	// Input validation
	if cell < 0 || cell > 8 {
		return ErrInvalidCellIndex
	}
	if player != MarkX && player != MarkO {
		return ErrInvalidPlayerMark
	}

	// Game State checks
	if g.Status == StatusFinished {
		return ErrGameFinished
	}
	if g.Status != StatusInProgress {
		return ErrGameNotInProgress
	}

	// Turn check
	if g.NextTurn != player {
		return ErrNotPlayersTurn
	}

	// Cell occupancy check
	if g.Board[cell] != MarkEmpty {
		return ErrCellOccupied
	}

	// Apply move
	g.Board[cell] = player
	g.MoveNumber++
	g.Version++

	// Check outcome
	w := winner(g.Board)
	if w == MarkX || w == MarkO {
		g.Status = StatusFinished
		g.Winner = w
		g.IsDraw = false
		g.recordOutcome()
		return nil
	}
	if isDraw(g.Board) {
		g.Status = StatusFinished
		g.Winner = MarkEmpty
		g.IsDraw = true
		g.recordOutcome()
		return nil
	}

	// Continue game
	g.Winner = MarkEmpty
	g.IsDraw = false

	// Switch turn
	if player == MarkX {
		g.NextTurn = MarkO
	} else {
		g.NextTurn = MarkX
	}
	return nil
}

type RematchResult int

const (
	RematchNoop RematchResult = iota
	RematchRequested
	RematchStarted
)

func (g *Game) RequestRematch(player Mark) (RematchResult, error) {
	if player != MarkX && player != MarkO {
		return RematchNoop, ErrInvalidPlayerMark
	}
	if g.Status != StatusFinished {
		return RematchNoop, ErrGameNotFinished
	}
	if g.PlayerO == nil {
		return RematchNoop, ErrPlayerOMissing
	}

	switch player {
	case MarkX:
		if g.RematchXRequested {
			return RematchNoop, nil
		}
		g.RematchXRequested = true
	case MarkO:
		if g.RematchORequested {
			return RematchNoop, nil
		}
		g.RematchORequested = true
	}

	g.Version++

	if !g.RematchXRequested || !g.RematchORequested {
		return RematchRequested, nil
	}

	g.resetRound()
	return RematchStarted, nil
}

func (g *Game) recordOutcome() {
	switch {
	case g.IsDraw:
		g.Draws++
	case g.Winner == MarkX:
		g.XWins++
	case g.Winner == MarkO:
		g.OWins++
	}
}

func (g *Game) resetRound() {
	for i := range g.Board {
		g.Board[i] = MarkEmpty
	}

	g.Status = StatusInProgress
	g.NextTurn = MarkX // [TODO] x or o? how to determine? people won last round?
	g.Winner = MarkEmpty
	g.IsDraw = false
	g.MoveNumber = 0
	g.RoundNumber++
	g.RematchXRequested = false
	g.RematchORequested = false
}

// returns MarkX or MarkO if there is a winner, otherwise MarkEmpty
func winner(b [9]Mark) Mark {
	lines := [8][3]int{
		{0, 1, 2}, // row 0
		{3, 4, 5}, // row 1
		{6, 7, 8}, // row 2
		{0, 3, 6}, // col 0
		{1, 4, 7}, // col 1
		{2, 5, 8}, // col 2
		{0, 4, 8}, // diag \
		{2, 4, 6}, // diag /
	}
	for _, line := range lines {
		a, c, d := line[0], line[1], line[2]
		if b[a] == MarkEmpty {
			continue
		}
		if b[a] == b[c] && b[c] == b[d] {
			return b[a]
		}
	}
	return MarkEmpty
}

// returns true if the board is full and there is no winner
func isDraw(b [9]Mark) bool {
	for _, cell := range b {
		if cell == MarkEmpty {
			return false
		}
	}
	// full board, no empties
	return winner(b) == MarkEmpty
}
