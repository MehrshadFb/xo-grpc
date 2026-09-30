package realtime

import (
	"testing"
	"time"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

const testGrace = 50 * time.Millisecond

func newTestPresence(t *testing.T) (*Presence, Subscriber) {
	t.Helper()

	hub := NewHub()
	sub := hub.Subscribe("game1")
	t.Cleanup(func() { hub.Unsubscribe("game1", sub) })

	presence := NewPresence(hub, testGrace, func(gameID string) (*domaingame.Game, error) {
		return domaingame.NewGame(gameID, "CODE1"), nil
	})

	return presence, sub
}

func expectEvent(t *testing.T, sub Subscriber, eventType EventType, mark domaingame.Mark) {
	t.Helper()

	select {
	case event := <-sub:
		if event.Type != eventType || event.PlayerMark != mark {
			t.Fatalf("expected event %v for %v, got %v for %v", eventType, mark, event.Type, event.PlayerMark)
		}
	case <-time.After(10 * testGrace):
		t.Fatalf("timed out waiting for event %v", eventType)
	}
}

func expectNoEvent(t *testing.T, sub Subscriber) {
	t.Helper()

	select {
	case event := <-sub:
		t.Fatalf("expected no event, got %v for %v", event.Type, event.PlayerMark)
	case <-time.After(3 * testGrace):
	}
}

func TestPresence_ReconnectWithinGraceIsSilent(t *testing.T) {
	presence, sub := newTestPresence(t)

	presence.Connect("game1", domaingame.MarkX)
	presence.Disconnect("game1", domaingame.MarkX)
	presence.Connect("game1", domaingame.MarkX)

	expectNoEvent(t, sub)
	if presence.IsAway("game1", domaingame.MarkX) {
		t.Fatalf("expected X not to be away")
	}
}

func TestPresence_LeftAfterGraceThenReturned(t *testing.T) {
	presence, sub := newTestPresence(t)

	presence.Connect("game1", domaingame.MarkO)
	presence.Disconnect("game1", domaingame.MarkO)

	expectEvent(t, sub, EventTypePlayerLeft, domaingame.MarkO)
	if !presence.IsAway("game1", domaingame.MarkO) {
		t.Fatalf("expected O to be away")
	}

	presence.Connect("game1", domaingame.MarkO)

	expectEvent(t, sub, EventTypePlayerReturned, domaingame.MarkO)
	if presence.IsAway("game1", domaingame.MarkO) {
		t.Fatalf("expected O not to be away after returning")
	}
}

func TestPresence_OtherTabStillOpenIsSilent(t *testing.T) {
	presence, sub := newTestPresence(t)

	presence.Connect("game1", domaingame.MarkX)
	presence.Connect("game1", domaingame.MarkX)
	presence.Disconnect("game1", domaingame.MarkX)

	expectNoEvent(t, sub)
}
