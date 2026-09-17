package transport

import (
	"net/http"
)

// New is the standard HTTP transport: cleartext HTTP/2 to the
// internal mesh, sent whether or not the remote is active so a peer that
// went silent is noticed on a held connection and not only on the next call.
//
// HTTP/2 is the only protocol enabled: net/http speaks unencrypted HTTP/2 to
// an http:// URL only when HTTP/1 is not also enabled, and gRPC and streaming
// RPCs need HTTP/2.
func New(http2Config *http.HTTP2Config) *http.Transport {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Transport{
		Protocols: protocols,
		HTTP2:     http2Config,
	}
}
