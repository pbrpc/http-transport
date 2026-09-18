//revive:disable:package-comments
package transport

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/pbrpc/testing/mocks/roundtripper"
)

func TestTransportWithReadiness(t *testing.T) {
	t.Run("substitutes the system clock and exponential schedule", func(t *testing.T) {
		base := roundtripper.Fail(errors.New("unexpected request"))

		transport, ok := WithReadiness(base, nil, nil).(*readyTransport)
		if !ok {
			t.Fatal("expected a readyTransport")
		}

		if _, ok := transport.clock.(systemClock); !ok {
			t.Errorf("clock = %T, want the system clock", transport.clock)
		}
		if _, ok := transport.newBackOff().(*backoff.ExponentialBackOff); !ok {
			t.Error("expected the library's exponential schedule")
		}
	})

	t.Run("keeps what it is given", func(t *testing.T) {
		base := roundtripper.Record(
			roundtripper.Respond(http.StatusNoContent, nil, ""),
		)
		clock := newClockStub()

		transport := WithReadiness(base, clock, fixedSchedule(time.Second)).(*readyTransport)

		if transport.base != base {
			t.Error("expected the base transport it was given")
		}
		if transport.clock != clock {
			t.Error("expected the clock it was given")
		}
		if _, ok := transport.newBackOff().(*scheduleStub); !ok {
			t.Error("expected the schedule it was given")
		}
	})
}
