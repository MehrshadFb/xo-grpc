package memory

import (
	"sync"
	"github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

type Store struct {
	mu sync.RWMutex
	byID       map[string]*game.Game
	byJoinCode map[string]string
}

func NewStore() *Store {
	return &Store{
		byID:       make(map[string]*game.Game),
		byJoinCode: make(map[string]string),
	}
}
