package main

import (
	"cherubgyre/config"
	"cherubgyre/controllers"
	"cherubgyre/services"
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// Request body size caps. 64 KiB is generous for any JSON payload this
// API accepts; register/duress/login get a tighter 8 KiB. Anything
// larger returns 413 at the middleware layer before touching a
// decoder, preventing memory-exhaustion DoS.
const (
	defaultMaxBodyBytes = 64 * 1024
	authMaxBodyBytes    = 8 * 1024
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// Install a JSON slog handler as the process-wide default logger.
	// Every package that imports log/slog and calls slog.Info(...) etc.
	// picks this up automatically — no package-level "logger" field
	// needed. Output goes to stdout so `docker logs` / `kubectl logs`
	// see it without extra plumbing.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	services.SetJWTSecret(cfg.JWTSecret)
	controllers.SetAdminToken(cfg.AdminToken)

	if err := services.BackfillUIDs(); err != nil {
		slog.Error("UID backfill failed", slog.Any("err", err))
	}

	router := mux.NewRouter()

	// Public routes — no auth required.
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("You've reached cherubgyre"))
	}).Methods("GET")
	router.HandleFunc("/health", controllers.Health).Methods("GET")
	router.HandleFunc("/ready", controllers.Ready).Methods("GET")

	// Public mutating routes get the tighter body cap — these see
	// unauthenticated traffic and are the biggest DoS surface.
	tight := func(h http.HandlerFunc) http.HandlerFunc {
		return controllers.MaxBodyBytesFunc(authMaxBodyBytes, h)
	}
	router.HandleFunc("/register", tight(controllers.Register)).Methods("POST")
	router.HandleFunc("/validate-invite", tight(controllers.ValidateInviteCode)).Methods("POST")
	router.HandleFunc("/login", tight(controllers.Login)).Methods("POST")

	// Authenticated routes — every handler below can assume a valid
	// principal via controllers.Identity(r).
	auth := func(h http.HandlerFunc) http.HandlerFunc { return controllers.RequireAuth(h) }

	router.HandleFunc("/profile", auth(controllers.Profile)).Methods("GET")
	router.HandleFunc("/user/change-pin", auth(controllers.ChangePin)).Methods("POST")
	router.HandleFunc("/user/change-duress-pin", auth(controllers.ChangeDuressPin)).Methods("POST")
	router.HandleFunc("/invite", auth(controllers.Invite)).Methods("GET")

	// Follow graph.
	router.HandleFunc("/follow/requests", auth(controllers.GetFollowRequests)).Methods("GET")
	router.HandleFunc("/follow/accept/{username}", auth(controllers.AcceptFollow)).Methods("POST")
	router.HandleFunc("/follow/decline/{username}", auth(controllers.DeclineFollow)).Methods("POST")
	router.HandleFunc("/follow/{username}", auth(controllers.FollowUser)).Methods("POST")
	router.HandleFunc("/unfollow/{username}", auth(controllers.UnfollowUser)).Methods("POST")
	router.HandleFunc("/followers/{username}", auth(controllers.GetFollowers)).Methods("GET")
	router.HandleFunc("/following", auth(controllers.GetFollowing)).Methods("GET")
	router.HandleFunc("/followers/{username}", auth(controllers.BanFollower)).Methods("DELETE")

	// Duress.
	router.HandleFunc("/duress", auth(controllers.PostDuress)).Methods("POST")
	router.HandleFunc("/duress/cancel", auth(controllers.CancelDuress)).Methods("POST")
	router.HandleFunc("/users/map", auth(controllers.GetDuressMap)).Methods("GET")
	router.HandleFunc("/duress/following", auth(controllers.GetFollowingDuress)).Methods("GET")
	router.HandleFunc("/duress/verify", auth(controllers.VerifyAccess)).Methods("POST")
	router.HandleFunc("/duress/dismiss/{username}", auth(controllers.DismissDuressNotification)).Methods("POST")

	// Admin — token-auth'd via requireAdmin inside the handler.
	router.HandleFunc("/admin/users/{username}", controllers.AdminDeregisterUser).Methods("DELETE")

	// Root context for the process. Canceled on SIGINT/SIGTERM so
	// background goroutines can unwind cleanly.
	rootCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	// Background inactivity sweep. Skipped when RUN_INACTIVITY_JOB=false
	// so a multi-replica deployment can disable it everywhere except a
	// single worker. Default: enabled.
	if os.Getenv("RUN_INACTIVITY_JOB") != "false" {
		go func() {
			slog.Info("background inactivity checker started")
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			if err := services.CheckInactivity(rootCtx); err != nil {
				slog.Error("initial inactivity check failed", slog.Any("err", err))
			}
			for {
				select {
				case <-rootCtx.Done():
					slog.Info("inactivity checker stopping due to shutdown")
					return
				case <-ticker.C:
					if err := services.CheckInactivity(rootCtx); err != nil {
						slog.Error("scheduled inactivity check failed", slog.Any("err", err))
					}
				}
			}
		}()
	}

	// Apply global middleware. Order matters:
	//   1. Body-size cap — enforced first so an unauthenticated
	//      attacker can't chew memory trying to blow past it.
	//   2. Request ID — injected early so every log line and
	//      downstream handler can reference the same ID.
	//   3. Security headers — set on every response including 413s.
	//   4. Router — applies RequireAuth per authed route.
	handler := controllers.MaxBodyBytes(defaultMaxBodyBytes)(router)
	handler = controllers.RequestID(handler)
	handler = controllers.SecurityHeaders(handler)

	// Wrap the listener in an explicit http.Server with sane timeouts
	// so a slow or malicious client cannot tie up goroutines indefinitely
	// by opening a connection and never sending headers.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Run ListenAndServe in a goroutine so we can react to signals in
	// the main goroutine.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server starting", slog.String("addr", ":"+cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	// Wait for either a server crash or a signal. On signal, give
	// in-flight requests up to 30 seconds to drain before forcing the
	// listener closed.
	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	case <-rootCtx.Done():
		slog.Info("shutdown signal received, draining connections")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown error", slog.Any("err", err))
		} else {
			slog.Info("server stopped cleanly")
		}
	}
}
