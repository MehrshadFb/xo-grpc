package realtime

import domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"

type Subscriber chan *domaingame.Game // channel for transmitting game state updates