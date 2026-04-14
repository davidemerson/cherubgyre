package services

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// IPLimiter is a simple in-memory sliding-window rate limiter keyed by
// client IP. Acceptable for single-instance deployments; for a multi-replica
// setup this should be replaced with a shared store (Redis or similar).
type IPLimiter struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	hits   map[string][]time.Time
}

// NewIPLimiter returns a limiter that allows `limit` events per `window`
// per client IP.
func NewIPLimiter(limit int, window time.Duration) *IPLimiter {
	return &IPLimiter{
		window: window,
		limit:  limit,
		hits:   make(map[string][]time.Time),
	}
}

// Allow records an attempt for the given IP and returns nil if it is within
// the limit, or an error if the limit has been exceeded. Old entries are
// pruned as a side effect so the map does not grow unboundedly.
func (l *IPLimiter) Allow(ip string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	history := l.hits[ip]
	kept := history[:0]
	for _, t := range history {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= l.limit {
		l.hits[ip] = kept
		return errors.New("rate limit exceeded")
	}

	l.hits[ip] = append(kept, now)
	return nil
}

// ClientIP returns a best-effort client IP derived from X-Forwarded-For,
// X-Real-IP, or RemoteAddr, in that order. Behind a trusted reverse proxy
// the first header is authoritative; in a direct-to-app deployment the
// forwarding headers are ignored by RemoteAddr.
func ClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if idx := strings.Index(fwd, ","); idx >= 0 {
			return strings.TrimSpace(fwd[:idx])
		}
		return strings.TrimSpace(fwd)
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return strings.TrimSpace(real)
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	return host
}

// LoginLimiter caps login attempts from a single IP. Tuned loosely: a
// legitimate user typing their PIN wrong a few times is fine; an attacker
// scanning the username space gets stopped cold. Limit and window can be
// overridden via LOGIN_RATE_LIMIT and LOGIN_RATE_WINDOW_SECONDS env vars
// so an integration test harness running many rapid requests from a
// single IP can raise the ceiling without surgery to the source.
var LoginLimiter = loginLimiterFromEnv()

func loginLimiterFromEnv() *IPLimiter {
	limit := envInt("LOGIN_RATE_LIMIT", 10)
	window := time.Duration(envInt("LOGIN_RATE_WINDOW_SECONDS", 300)) * time.Second
	return NewIPLimiter(limit, window)
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
