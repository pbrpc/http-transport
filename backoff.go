//revive:disable:package-comments
package transport

import "github.com/cenkalti/backoff/v7"

// BackOffFactory makes the schedule one host's attempts follow while it cannot be
// reached. Each host gets its own, made when its first attempt fails.
type BackOffFactory func() backoff.BackOff
