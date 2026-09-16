//revive:disable:package-comments
package transport

import "time"

// Clock is what the ready transport needs of time: the current instant, and a
// channel that fires after a duration. A test supplies one it controls.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// systemClock is the Clock production uses.
type systemClock struct{}

func (systemClock) Now() time.Time                         { return time.Now() }
func (systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
