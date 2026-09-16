//revive:disable:package-comments
package transport

import (
	"testing"
	"time"
)

// clockStub is a Clock the test drives: Now is fixed, and After records the
// delay asked for, signals asked, and answers with a channel the test fires.
type clockStub struct {
	now    time.Time
	fire   chan time.Time
	asked  chan struct{}
	waited []time.Duration
}

func newClockStub() *clockStub {
	return &clockStub{
		now:   time.Unix(1_000_000, 0),
		fire:  make(chan time.Time),
		asked: make(chan struct{}, 16),
	}
}

func (c *clockStub) Now() time.Time { return c.now }

func (c *clockStub) After(d time.Duration) <-chan time.Time {
	c.waited = append(c.waited, d)
	c.asked <- struct{}{}

	return c.fire
}

func TestSystemClock(t *testing.T) {
	clock := systemClock{}

	if clock.Now().IsZero() {
		t.Error("expected the current time")
	}

	select {
	case <-clock.After(0):
	case <-t.Context().Done():
		t.Fatal("After never fired")
	}
}
