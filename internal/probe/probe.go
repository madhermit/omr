// Package probe waits for a TCP port's HTTP layer to respond.
package probe

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	initialBackoff = 100 * time.Millisecond
	maxBackoff     = 500 * time.Millisecond
	httpTimeout    = 2 * time.Second
)

// WaitReady probes localhost:port until the HTTP layer responds, or ctx is done.
//
// "localhost" (not "127.0.0.1") so Go's resolver tries both v4 and v6 addresses
// — necessary because some dev servers (e.g., Nuxt) bind to ::1 only by default.
//
// HTTP HEAD / counts as ready on any response (200, 404, 405, …) so the target
// service doesn't need a dedicated health endpoint.
//
// We don't try to distinguish the about-to-die old process from the new one.
// In dev use, the brief window where the old proc could answer before overmind
// kills it is small enough that the worst-case is a one-refresh 502, which is
// far better than the polling races a "wait for the port to go down" gate
// introduces (those can hang for the full timeout).
func WaitReady(ctx context.Context, port int) error {
	addr := fmt.Sprintf("localhost:%d", port)
	backoff := initialBackoff

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if httpResponds(ctx, addr) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func httpResponds(ctx context.Context, addr string) bool {
	reqCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodHead, "http://"+addr+"/", nil)
	if err != nil {
		return false
	}
	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
