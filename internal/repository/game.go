package repository

import "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type GameRepository interface {
	Create(g *game.Game) error
	GetByID(id string) (*game.Game, error)
	GetByJoinCode(joinCode string) (*game.Game, error)
	Update(g *game.Game) error
}
