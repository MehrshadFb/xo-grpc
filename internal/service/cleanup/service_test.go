package cleanup

import (
	"context"
	"testing"
	"time"
)

type fakeDeleter struct {
	waitingBefore    time.Time
	inProgressBefore time.Time
	endedBefore      time.Time
}

func (f *fakeDeleter) DeleteStale(_ context.Context, waitingBefore, inProgressBefore, endedBefore time.Time) (int64, error) {
	f.waitingBefore = waitingBefore
	f.inProgressBefore = inProgressBefore
	f.endedBefore = endedBefore
	return 3, nil
}

func TestRunOnce_UsesPolicyCutoffs(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	deleter := &fakeDeleter{}

	service := NewService(deleter, DefaultPolicy)
	service.now = func() time.Time { return now }

	deleted, err := service.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce error: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 deleted, got %d", deleted)
	}

	if want := now.Add(-time.Hour); !deleter.waitingBefore.Equal(want) {
		t.Fatalf("waiting cutoff: expected %v, got %v", want, deleter.waitingBefore)
	}
	if want := now.Add(-24 * time.Hour); !deleter.inProgressBefore.Equal(want) {
		t.Fatalf("in-progress cutoff: expected %v, got %v", want, deleter.inProgressBefore)
	}
	if want := now.Add(-7 * 24 * time.Hour); !deleter.endedBefore.Equal(want) {
		t.Fatalf("ended cutoff: expected %v, got %v", want, deleter.endedBefore)
	}
}
