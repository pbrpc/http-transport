//revive:disable:package-comments
package transport

import (
	"time"

	"github.com/cenkalti/backoff/v7"
)

// scheduleStub is a BackOff answering a fixed sequence of delays, so a test
// knows what the transport should wait. It records how many were made and
// whether it was reset.
type scheduleStub struct {
	delays []time.Duration
	calls  int
	resets int
}

func (s *scheduleStub) NextBackOff() time.Duration {
	delay := s.delays[min(s.calls, len(s.delays)-1)]
	s.calls++

	return delay
}

func (s *scheduleStub) Reset() { s.resets++ }

// fixedSchedule makes every host follow the given delays.
func fixedSchedule(delays ...time.Duration) BackOffFactory {
	return func() backoff.BackOff { return &scheduleStub{delays: delays} }
}
