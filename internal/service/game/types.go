package gamesvc

import domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type GetStateResult struct {
	Game *domaingame.Game
	Mark domaingame.Mark
}

type MakeMoveResult struct {
	Game *domaingame.Game
}

type RequestRematchResult struct {
	Game    *domaingame.Game
	Started bool
	Changed bool
}
