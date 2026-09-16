//revive:disable:package-comments
package transport

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/pbrpc/connect-testing/mocks/roundtripper"
)

var errUnreachable = errors.New("connection refused")

func newRequest(ctx context.Context, t *testing.T, host string) *http.Request {
	t.Helper()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+host+"/", nil)
	if err != nil {
		t.Fatalf("http.NewRequestWithContext() error = %v", err)
	}

	return request
}

func attempt(
	ctx context.Context,
	t *testing.T,
	transport http.RoundTripper,
	host string,
) <-chan error {
	t.Helper()
	result := make(chan error, 1)

	go func() {
		_, err := transport.RoundTrip(newRequest(ctx, t, host))
		result <- err
	}()

	return result
}

func TestReadyTransport(t *testing.T) {
	const h1, h2 = "h1:50051", "h2:50051"

	t.Run("sends at once while the host is reachable", func(t *testing.T) {
		wire := roundtripper.Record(
			roundtripper.Respond(http.StatusServiceUnavailable, nil, ""),
		)
		clock := newClockStub()
		transport := WithReadiness(wire, clock, fixedSchedule(time.Second))

		response, err := transport.RoundTrip(newRequest(t.Context(), t, h1))
		if err != nil {
			t.Fatalf("RoundTrip() error = %v, want nil", err)
		}
		if response.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
		}
		if len(clock.waited) != 0 {
			t.Errorf("waited %v, want no wait", clock.waited)
		}
	})

	t.Run("waits after failures and forgets a reachable host", func(t *testing.T) {
		down := map[string]bool{h1: true}
		wire := roundtripper.Record(func(request *http.Request) (*http.Response, error) {
			if down[request.URL.Host] {
				return roundtripper.Fail(errUnreachable)(request)
			}

			return roundtripper.Respond(
				http.StatusServiceUnavailable,
				nil,
				"",
			)(request)
		})
		clock := newClockStub()
		transport := WithReadiness(wire, clock, fixedSchedule(time.Second, 2*time.Second))

		if _, err := transport.RoundTrip(newRequest(t.Context(), t, h1)); !errors.Is(err, errUnreachable) {
			t.Fatalf("RoundTrip() error = %v, want %v", err, errUnreachable)
		}

		result := attempt(t.Context(), t, transport, h1)
		clock.fire <- clock.now
		if err := <-result; !errors.Is(err, errUnreachable) {
			t.Fatalf("RoundTrip() error = %v, want %v", err, errUnreachable)
		}

		delete(down, h1)
		result = attempt(t.Context(), t, transport, h1)
		clock.fire <- clock.now
		if err := <-result; err != nil {
			t.Fatalf("RoundTrip() error = %v, want nil", err)
		}

		want := []time.Duration{time.Second, 2 * time.Second}
		if len(clock.waited) != len(want) || clock.waited[0] != want[0] || clock.waited[1] != want[1] {
			t.Errorf("waited %v, want %v", clock.waited, want)
		}

		if _, err := transport.RoundTrip(newRequest(t.Context(), t, h1)); err != nil {
			t.Fatalf("RoundTrip() error = %v, want nil", err)
		}
		if len(clock.waited) != len(want) {
			t.Errorf("waited %v, want no further wait", clock.waited)
		}
		if got := len(wire.Sent()); got != 4 {
			t.Errorf("wire calls = %d, want 4", got)
		}
	})

	t.Run("keeps hosts independent", func(t *testing.T) {
		wire := roundtripper.Record(func(request *http.Request) (*http.Response, error) {
			if request.URL.Host == h1 {
				return roundtripper.Fail(errUnreachable)(request)
			}

			return roundtripper.Respond(http.StatusNoContent, nil, "")(request)
		})
		clock := newClockStub()
		transport := WithReadiness(wire, clock, fixedSchedule(time.Second))

		if _, err := transport.RoundTrip(newRequest(t.Context(), t, h1)); !errors.Is(err, errUnreachable) {
			t.Fatalf("RoundTrip() error = %v, want %v", err, errUnreachable)
		}
		if _, err := transport.RoundTrip(newRequest(t.Context(), t, h2)); err != nil {
			t.Fatalf("RoundTrip() error = %v, want nil", err)
		}
		if len(clock.waited) != 0 {
			t.Errorf("waited %v, want no wait for another host", clock.waited)
		}
	})

	t.Run("attempts immediately after the schedule stops", func(t *testing.T) {
		wire := roundtripper.Record(roundtripper.Fail(errUnreachable))
		clock := newClockStub()
		transport := WithReadiness(wire, clock, fixedSchedule(backoff.Stop))

		for range 2 {
			if _, err := transport.RoundTrip(newRequest(t.Context(), t, h1)); !errors.Is(err, errUnreachable) {
				t.Fatalf("RoundTrip() error = %v, want %v", err, errUnreachable)
			}
		}
		if len(clock.waited) != 0 {
			t.Errorf("waited %v, want no wait", clock.waited)
		}
	})

	t.Run("stops waiting when the request context ends", func(t *testing.T) {
		wire := roundtripper.Record(roundtripper.Fail(errUnreachable))
		clock := newClockStub()
		transport := WithReadiness(wire, clock, fixedSchedule(time.Second))

		if _, err := transport.RoundTrip(newRequest(t.Context(), t, h1)); !errors.Is(err, errUnreachable) {
			t.Fatalf("RoundTrip() error = %v, want %v", err, errUnreachable)
		}

		ctx, cancel := context.WithCancel(t.Context())
		result := attempt(ctx, t, transport, h1)

		select {
		case <-clock.asked:
		case <-t.Context().Done():
			t.Fatal("transport never waited")
		}
		cancel()

		if err := <-result; !errors.Is(err, context.Canceled) {
			t.Fatalf("RoundTrip() error = %v, want %v", err, context.Canceled)
		}
		if got := len(wire.Sent()); got != 1 {
			t.Errorf("wire calls = %d, want 1", got)
		}
	})
}
