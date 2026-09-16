package transport

import (
	"net/http"
	"testing"
)

func TestNewTransport(t *testing.T) {
	transport := New(&http.HTTP2Config{})

	// Cleartext HTTP/2 is used for http:// URLs only when HTTP/1 is not also
	// enabled; with both, net/http speaks HTTP/1.1 and gRPC cannot pass.
	if transport.Protocols.HTTP1() || !transport.Protocols.UnencryptedHTTP2() {
		t.Errorf("protocols = %v, want cleartext HTTP/2 alone", transport.Protocols)
	}
}
