package transport

import (
	"net/http"
)

// New is the standard HTTP transport: cleartext HTTP/2 to the
// internal mesh, with keepalive pings read from GRPC_KEEPALIVE_TIME and
// GRPC_KEEPALIVE_TIMEOUT, sent whether or not an RPC is active so a peer that
// went silent is noticed on a held connection and not only on the next call.
// It is what NewHTTPClient and NewReadyTransport use when given no base, and
// what a caller wraps when it puts its own RoundTripper under them.
//
// HTTP/2 is the only protocol enabled: net/http speaks unencrypted HTTP/2 to
// an http:// URL only when HTTP/1 is not also enabled, and gRPC and the
// streaming RPCs need HTTP/2. Every server on the mesh accepts it.
func New(http2Config *http.HTTP2Config) *http.Transport {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Transport{
		Protocols: protocols,
		HTTP2:     http2Config,
	}
}
