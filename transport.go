package transport

import (
	"net/http"

	"github.com/caarlos0/env/v11"
)

// From is the standard HTTP transport: cleartext HTTP/2 to the
// internal mesh, sent whether or not the remote is active so a peer that
// went silent is noticed on a held connection and not only on the next call.
//
// HTTP/2 is the only protocol enabled: net/http speaks unencrypted HTTP/2 to
// an http:// URL only when HTTP/1 is not also enabled, and gRPC and streaming
// RPCs need HTTP/2.
func From(http2Config *http.HTTP2Config) (*http.Transport, error) {
	configured, err := env.ParseAs[configuration]()
	if err != nil {
		return nil, err
	}

	if http2Config == nil {
		http2Config = &http.HTTP2Config{}
	}

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	http2Config.SendPingTimeout = configured.SendPingTimeout
	http2Config.PingTimeout = configured.PingTimeout

	return &http.Transport{
		Protocols: protocols,
		HTTP2:     http2Config,
	}, nil
}
