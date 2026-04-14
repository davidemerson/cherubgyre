package controllers

import (
	"cherubgyre/services"
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// sanitizeForLog strips newlines and other control characters from a
// caller-supplied string before it is written to a log line. Without
// this, an attacker could put `\n` (URL-encoded as `%0A`) into a URL
// path segment and inject forged log records.
func sanitizeForLog(s string) string {
	const maxLen = 128
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

// MaxBodyBytes wraps a handler so any read on the request body past n
// bytes fails with an http.MaxBytesError. Callers' json.Decode invocations
// then naturally return an error on oversized input; the helper below
// translates that to 413.
func MaxBodyBytes(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

// MaxBodyBytesFunc is the http.HandlerFunc variant of MaxBodyBytes, for
// per-route wrapping in main.go where we already use auth() wrappers.
func MaxBodyBytesFunc(n int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, n)
		next(w, r)
	}
}

// SecurityHeaders sets conservative defense-in-depth response headers on
// every response, regardless of handler. Applied globally in main.go so
// even the static `/` and `/health` endpoints carry them. Values are
// intentionally strict: this API never serves HTML, never wants to be
// framed, never wants to be cached, and should always be TLS if a client
// honors HSTS.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// authCtxKey is unexported so callers must go through the helpers to read
// identity off a request context.
type authCtxKey struct{}

// requestIDCtxKey carries the per-request ID installed by RequestID.
type requestIDCtxKey struct{}

// RequestIDFromContext returns the request ID installed by RequestID
// middleware, or an empty string if the middleware was not applied.
// Exported so handlers and services can include it in slog fields.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// RequestID is middleware that ensures every request carries a stable
// identifier for the duration of its processing. If the client sends
// an X-Request-ID header we trust it (trimmed, capped at 128 chars,
// sanitized to printable ASCII); otherwise we mint a fresh UUIDv4.
// The value is attached to the request context and echoed back as a
// response header so clients can correlate with their own logs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = uuid.NewString()
		} else {
			id = sanitizeForLog(id)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDCtxKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authPrincipal is what RequireAuth attaches to the request context after
// it has validated the bearer token.
type authPrincipal struct {
	Username string
	IsDuress bool
	Token    string // raw token string, preserved for services that still take it
}

// RequireAuth validates the bearer token and forwards the request with an
// authPrincipal attached. Handlers can then call Identity(r) instead of
// re-parsing the Authorization header themselves. This is the single
// source of truth for "is this request authenticated?" — before this
// middleware existed, each controller did its own header parse with
// subtly different behavior.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		valid, err := services.ValidateToken(token)
		if err != nil || !valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		username, err := services.GetUsernameFromToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		p := authPrincipal{
			Username: username,
			IsDuress: services.IsDuressToken(token),
			Token:    token,
		}
		ctx := context.WithValue(r.Context(), authCtxKey{}, p)
		next(w, r.WithContext(ctx))
	}
}

// Identity returns the authenticated principal for a request. Panics if
// the handler was mounted without RequireAuth — that's a programming
// error, not a runtime condition, so failing loudly is correct.
func Identity(r *http.Request) authPrincipal {
	v, ok := r.Context().Value(authCtxKey{}).(authPrincipal)
	if !ok {
		panic("controllers.Identity: handler not protected by RequireAuth")
	}
	return v
}
