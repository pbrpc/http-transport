//revive:disable:package-comments
package transport

import (
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v7"
)

// hostState is a host that could not be reached: the schedule its attempts
// follow, and when the next one may go.
type hostState struct {
	backoff backoff.BackOff
	next    time.Time
}

// readyTransport paces requests to hosts that cannot be reached. An HTTP
// transport tries a dial on every request, so the loops that reopen a stream
// after it drops would spin against a peer that is down. This gates them: after
// a failure to reach a host, every request to that host waits until its
// schedule allows the next attempt, and a request that gets through clears it.
//
// A request is held before it is sent and never sent twice, so a streaming
// request body is as safe here as a unary one. Hosts are kept apart: one being
// down holds nothing bound for another.
type readyTransport struct {
	base       http.RoundTripper
	clock      Clock
	newBackOff BackOffFactory

	// mu guards hosts. Requests arrive on their own goroutines, and a BackOff
	// is not safe for concurrent use.
	mu    sync.Mutex
	hosts map[string]*hostState
}

// WithReadiness wraps base with pacing. base nil means the standard transport,
// clock nil means the system clock; newBackOff nil means the backoff library's
// exponential schedule with its defaults.
func WithReadiness(
	base http.RoundTripper,
	clock Clock,
	newBackOff BackOffFactory,
) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	if clock == nil {
		clock = systemClock{}
	}

	if newBackOff == nil {
		newBackOff = func() backoff.BackOff { return backoff.NewExponentialBackOff() }
	}

	return &readyTransport{
		base:       base,
		clock:      clock,
		newBackOff: newBackOff,
		hosts:      map[string]*hostState{},
	}
}
