package lobby

import "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type CreateGameResult struct {
	Game        *game.Game
	PlayerToken string
	PlayerID    string
}

type JoinGameResult struct {
	Game        *game.Game
	PlayerToken string
	PlayerID    string
}
