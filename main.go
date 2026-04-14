package main

import (
	"cherubgyre/config"
	"cherubgyre/controllers"
	"cherubgyre/services"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}
	services.SetJWTSecret(cfg.JWTSecret)
	controllers.SetAdminToken(cfg.AdminToken)

	if err := services.MigratePinHashes(); err != nil {
		log.Printf("PIN hash migration error: %v", err)
	}
	if err := services.BackfillUIDs(); err != nil {
		log.Printf("UID backfill error: %v", err)
	}

	router := mux.NewRouter()

	// Public routes — no auth required.
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("You've reached cherubgyre"))
	}).Methods("GET")
	router.HandleFunc("/health", controllers.Health).Methods("GET")
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/validate-invite", controllers.ValidateInviteCode).Methods("POST")
	router.HandleFunc("/login", controllers.Login).Methods("POST")

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

	// Background inactivity sweep. Skipped when RUN_INACTIVITY_JOB=false
	// so a multi-replica deployment can disable it everywhere except a
	// single worker. Default: enabled.
	if os.Getenv("RUN_INACTIVITY_JOB") != "false" {
		go func() {
			log.Println("Starting background inactivity checker...")
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			if err := services.CheckInactivity(); err != nil {
				log.Printf("Error running initial inactivity check: %v", err)
			}
			for range ticker.C {
				if err := services.CheckInactivity(); err != nil {
					log.Printf("Error running scheduled inactivity check: %v", err)
				}
			}
		}()
	}

	// Wrap the listener in an explicit http.Server with sane timeouts
	// so a slow or malicious client cannot tie up goroutines indefinitely
	// by opening a connection and never sending headers.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Println("Starting server on :" + cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
