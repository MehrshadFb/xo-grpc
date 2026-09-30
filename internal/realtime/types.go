package realtime

import domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type EventType int

const (
	EventTypeUnspecified EventType = iota
	EventTypePlayerJoined
	EventTypeMoveMade
	EventTypeGameOver
	EventTypeRematchRequested
	EventTypeRoundStarted
	EventTypePlayerLeft
	EventTypePlayerReturned
)

type Event struct {
	Type           EventType
	Game           *domaingame.Game
	GameOverReason string
	PlayerMark     domaingame.Mark
}

type Subscriber chan Event
