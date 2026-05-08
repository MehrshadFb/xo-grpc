// References:
// - https://go101.org/article/channel.html
// - https://go.dev/tour/concurrency/2

package realtime

import (
	"sync"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

type Hub struct {
	mu sync.RWMutex

	// gameID -> subscribers
	subscribers map[string]map[Subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[Subscriber]struct{}),
	}
}

func (h *Hub) Subscribe(gameID string) Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()

	// if we don't have a buffer size, the sender will block until the channel is read from
	// if we have a buffer size, the sender will block until the channel is full
	sub := make(Subscriber, 8) // create a new channel with a buffer size of 8

	if _, ok := h.subscribers[gameID]; !ok {
		h.subscribers[gameID] = make(map[Subscriber]struct{})
	}

	h.subscribers[gameID][sub] = struct{}{}

	return sub
}

func (h *Hub) Unsubscribe(gameID string, sub Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs, ok := h.subscribers[gameID]
	if !ok {
		return
	}

	delete(subs, sub) // remove the subscriber from the subscribers map
	close(sub) // close the channel to notify the subscriber that it is no longer needed

	if len(subs) == 0 {
		delete(h.subscribers, gameID) // remove gameID from subscribers map when no subscribers are left
	}
}

func (h *Hub) Publish(gameID string, g *domaingame.Game) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.subscribers[gameID]
	if !ok {
		return
	}

	for sub := range subs {
		select {
		case sub <- g:
		default:
			// skip slow subscriber
		}
	}
}