package controllers

import (
	"cherubgyre/services"
	"context"
	"net/http"
)

// authCtxKey is unexported so callers must go through the helpers to read
// identity off a request context.
type authCtxKey struct{}

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
