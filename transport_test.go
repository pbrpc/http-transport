package transport

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
)

func TestTransport(t *testing.T) {
	t.Run("resolves the correct configuration", func(t *testing.T) {
		t.Setenv("HTTP2_SEND_PING_TIMEOUT", "3m")
		t.Setenv("HTTP2_PING_TIMEOUT", "30s")

		configured, err := env.ParseAs[configuration]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if configured.SendPingTimeout != 3*time.Minute {
			t.Errorf("send ping timeout = %v, want 3m", configured.SendPingTimeout)
		}
		if configured.PingTimeout != 30*time.Second {
			t.Errorf("ping timeout = %v, want 30s", configured.PingTimeout)
		}
	})

	t.Run("defaults configuration when given none",
		func(t *testing.T) {
			transport, err := From(nil)
			if err != nil {
				t.Fatal("error = nil, want the configuration error")
			}

			def := &http.HTTP2Config{
				SendPingTimeout: 2 * time.Minute,
				PingTimeout:     20 * time.Second,
			}
			if !reflect.DeepEqual(transport.HTTP2, def) {
				t.Fatalf("HTTP2 = %+v, want %+v", transport.HTTP2, def)
			}
		})

	t.Run("creates new valid transport", func(t *testing.T) {
		transport, err := From(&http.HTTP2Config{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Cleartext HTTP/2 is used for http:// URLs only when HTTP/1 is not also
		// enabled; with both, net/http speaks HTTP/1.1 and gRPC cannot pass.
		if transport.Protocols.HTTP1() || !transport.Protocols.UnencryptedHTTP2() {
			t.Errorf("protocols = %v, want cleartext HTTP/2 alone", transport.Protocols)
		}
	})

	t.Run("fails to create transport when improperly configured",
		func(t *testing.T) {
			t.Setenv("HTTP2_SEND_PING_TIMEOUT", "not-a-duration")

			if _, err := From(nil); err == nil {
				t.Fatal("error = nil, want the configuration error")
			}
		})
}
