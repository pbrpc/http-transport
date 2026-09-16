//revive:disable:package-comments
package transport

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v7"
)

// RoundTrip waits until the host may be attempted, then sends the request. A
// response, with whatever status, proves the host reachable and clears its
// state; no response starts or extends its schedule.
func (t *readyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host

	if err := t.wait(req, host); err != nil {
		return nil, err
	}

	response, err := t.base.RoundTrip(req)

	t.record(host, err == nil)

	return response, err
}

// wait blocks until host's next attempt may go or the request's context ends.
// A host with no state may be attempted at once.
func (t *readyTransport) wait(req *http.Request, host string) error {
	t.mu.Lock()
	var next time.Time
	if state, found := t.hosts[host]; found {
		next = state.next
	}
	t.mu.Unlock()

	delay := next.Sub(t.clock.Now())
	if delay <= 0 {
		return nil
	}

	select {
	case <-t.clock.After(delay):
		return nil
	case <-req.Context().Done():
		return req.Context().Err()
	}
}

// record notes whether host was reached. Reached forgets the host; not reached
// schedules its next attempt, starting its schedule if this was the first
// failure. A schedule that answers Stop has nothing more to wait for, so the
// next attempt is immediate.
func (t *readyTransport) record(host string, reached bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if reached {
		delete(t.hosts, host)

		return
	}

	state, found := t.hosts[host]
	if !found {
		state = &hostState{backoff: t.newBackOff()}
		t.hosts[host] = state
	}

	delay := state.backoff.NextBackOff()
	if delay == backoff.Stop {
		state.next = time.Time{}

		return
	}

	state.next = t.clock.Now().Add(delay)
}
