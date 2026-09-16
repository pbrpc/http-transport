# http-transport

`http-transport` builds cleartext HTTP/2 transports and provides per-host
readiness pacing for HTTP requests.

## Installation

```bash
go get github.com/pbrpc/http-transport
```

## Transport

`New` creates an `http.Transport` using cleartext HTTP/2 and the supplied
`http.HTTP2Config`:

```go
wire := transport.New(&http.HTTP2Config{
	SendPingTimeout: 2 * time.Minute,
	PingTimeout:     20 * time.Second,
})
```

The caller constructs `http.HTTP2Config` from the configuration sources it
owns.

## Readiness

`WithReadiness` wraps an `http.RoundTripper`. After a request fails before
receiving an HTTP response, later requests to that host follow its backoff
schedule. Any HTTP response marks the host reachable and clears that state.

```go
wire := transport.New(http2Config)
ready := transport.WithReadiness(wire, nil, nil)
client := &http.Client{Transport: ready}
```

Readiness state is maintained independently per host. Waiting respects the
request context, and every request is sent at most once.

A nil wrapped transport selects `http.DefaultTransport`. A nil clock selects
the system clock. A nil backoff factory selects the default exponential
schedule. Callers can supply each dependency explicitly when they need custom
behavior.
