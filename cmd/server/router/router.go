package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	clerkv2 "github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/dev3mike/go-api-swagger-boilerplate/cmd/server/dependencies"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/dtos"
	"github.com/dev3mike/go-api-swagger-boilerplate/internal/environments"
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

// InitializeRouter sets up a new router.
func InitializeRouter() *chi.Mux {
	clerkv2.SetKey(environments.Env.ClerkSecretKey)
	return chi.NewRouter()
}

// SetupRoutes configures all the routes for the router.
func SetupRoutes(r *chi.Mux, logger *slog.Logger) *chi.Mux {

	r.Route("/", func(r chi.Router) {
		// Public route
		r.Get("/", dependencies.IndexHandler.GetRootHandler)

		// Protected route
		r.With(injectClerkSession, userProtected).Post("/{username}", dependencies.IndexHandler.PostRootHandler)
	})

	// Clerk Webhooks
	r.Route("/webhooks/", func(r chi.Router) {
		// Make sure to register these endpoints in Clerk's dashboard
		r.Post("/clerk/create", dependencies.ClerkHandler.HandleCreateUserEvent)
		r.Post("/clerk/update", dependencies.ClerkHandler.HandleUpdateUserEvent)
	})

	setupSwagger(r, logger)

	return r
}

// injectClerkSession is middleware to inject Clerk session information.
func injectClerkSession(next http.Handler) http.Handler {
	return clerkhttp.WithHeaderAuthorization(clerkhttp.CustomClaimsConstructor(func(_ context.Context) any {
		return &dtos.UserClaims{}
	}))(next)
}

// userProtected is middleware to ensure the user is authenticated and authorized.
func userProtected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := clerkv2.SessionClaimsFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), dtos.TokenClaimsKey, claims.Custom)
		r = r.WithContext(ctx)

		fmt.Println("User ID: ", claims.Custom.(*dtos.UserClaims).Id)

		next.ServeHTTP(w, r)
	})
}

// setupSwagger configures Swagger if running in the development environment.
func setupSwagger(r *chi.Mux, logger *slog.Logger) {
	if environments.Env.Environment == "development" {
		logger.Info("🚀 Setting up Swagger UI")
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:"+environments.Env.Port+"/swagger/doc.json"), // URL pointing to API definition
		))
	}
}
