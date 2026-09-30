package realtime

import (
	"log/slog"
	"sync"
	"time"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

type presenceKey struct {
	gameID string
	mark   domaingame.Mark
}

type pendingLeave struct {
	timer *time.Timer
}

// Presence tracks which players have an open watch stream.
// When a player's last stream closes and they don't reconnect within the grace period:
// It publishes PlayerLeft; when they come back, it publishes PlayerReturned.
type Presence struct {
	mu sync.Mutex

	hub      *Hub
	grace    time.Duration
	loadGame func(gameID string) (*domaingame.Game, error)

	streams map[presenceKey]int           // open watch streams per player
	pending map[presenceKey]*pendingLeave // grace timers for players with no streams
	away    map[presenceKey]struct{}      // players announced as left
}

func NewPresence(hub *Hub, grace time.Duration, loadGame func(gameID string) (*domaingame.Game, error)) *Presence {
	return &Presence{
		hub:      hub,
		grace:    grace,
		loadGame: loadGame,
		streams:  make(map[presenceKey]int),
		pending:  make(map[presenceKey]*pendingLeave),
		away:     make(map[presenceKey]struct{}),
	}
}

func (p *Presence) Connect(gameID string, mark domaingame.Mark) {
	key := presenceKey{gameID: gameID, mark: mark}

	p.mu.Lock()
	p.streams[key]++

	// reconnected within the grace period - nobody is told
	if leave, ok := p.pending[key]; ok {
		leave.timer.Stop()
		delete(p.pending, key)
	}

	_, wasAway := p.away[key]
	delete(p.away, key)
	p.mu.Unlock()

	if wasAway {
		p.publish(key, EventTypePlayerReturned)
	}
}

func (p *Presence) Disconnect(gameID string, mark domaingame.Mark) {
	key := presenceKey{gameID: gameID, mark: mark}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.streams[key]--
	if p.streams[key] > 0 {
		return // still has another tab open
	}
	delete(p.streams, key)

	leave := &pendingLeave{}
	leave.timer = time.AfterFunc(p.grace, func() { p.expire(key, leave) })
	p.pending[key] = leave
}

func (p *Presence) IsAway(gameID string, mark domaingame.Mark) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, ok := p.away[presenceKey{gameID: gameID, mark: mark}]
	return ok
}

func (p *Presence) expire(key presenceKey, leave *pendingLeave) {
	p.mu.Lock()

	if p.pending[key] != leave || p.streams[key] > 0 {
		p.mu.Unlock()
		return
	}
	delete(p.pending, key)
	p.away[key] = struct{}{}
	p.mu.Unlock()

	p.publish(key, EventTypePlayerLeft)
}

func (p *Presence) publish(key presenceKey, eventType EventType) {
	g, err := p.loadGame(key.gameID)
	if err != nil {
		slog.Error("presence: load game failed", "game_id", key.gameID, "error", err)
		return
	}

	p.hub.Publish(key.gameID, Event{
		Type:       eventType,
		Game:       g,
		PlayerMark: key.mark,
	})
}
