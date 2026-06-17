package game

func NewGame(id string, joinCode string) *Game {
	g := &Game{
		ID:          id,
		JoinCode:    joinCode,
		Status:      StatusWaiting,
		NextTurn:    MarkX,
		Version:     1,
		RoundNumber: 1,
	}

	for i := range g.Board {
		g.Board[i] = MarkEmpty
	}

	return g
}
