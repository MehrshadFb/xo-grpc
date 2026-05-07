package gamesvc

import domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type GetStateResult struct {
	Game *domaingame.Game
}

type MakeMoveResult struct {
	Game *domaingame.Game
}