package game

func (g *Game) SetPlayerX(id, name string) {
	g.PlayerX = &Player{ID: id, DisplayName: name, Mark: MarkX}
}

func (g *Game) SetPlayerO(id, name string) {
	g.PlayerO = &Player{ID: id, DisplayName: name, Mark: MarkO}
}
