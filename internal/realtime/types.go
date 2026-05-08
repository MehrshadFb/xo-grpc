package realtime

import domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type EventType int

const (
	EventTypeUnspecified EventType = iota
	EventTypePlayerJoined
	EventTypeMoveMade
	EventTypeGameOver
)

type Event struct {
	Type           EventType
	Game           *domaingame.Game
	GameOverReason string
}

type Subscriber chan Event