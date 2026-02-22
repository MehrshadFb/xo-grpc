package session

import "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type Session struct {
	Token    string
	GameID   string
	PlayerID string
	Mark     game.Mark // X or O
}
