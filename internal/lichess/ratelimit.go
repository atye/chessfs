package lichess

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// RateLimitRoundTripper wraps a standard RoundTripper to handle 429 errors.
type RateLimitRoundTripper struct {
	Proxied http.RoundTripper
	Log     *slog.Logger
	mu      sync.Mutex
}

func (l *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	resp, err := l.Proxied.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		resp.Body.Close()

		l.Log.Info("Lichess Rate limit hit (429)! Enforcing a 60-second cooldown...")
		time.Sleep(60 * time.Second)

		return l.Proxied.RoundTrip(req)
	}

	return resp, nil
}
