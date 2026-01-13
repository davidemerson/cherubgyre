package main

import (
	"cherubgyre/controllers"
	"cherubgyre/services"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You've reached cherubgyre"))
	}).Methods("GET")
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/validate-invite", controllers.ValidateInviteCode).Methods("POST")
	router.HandleFunc("/health", controllers.Health).Methods("GET")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/profile", controllers.Profile).Methods("GET")
	router.HandleFunc("/user/change-pin", controllers.ChangePin).Methods("POST")
	router.HandleFunc("/user/change-duress-pin", controllers.ChangeDuressPin).Methods("POST")
	router.HandleFunc("/invite", controllers.Invite).Methods("GET")
	
	// Follow Request Routes
	router.HandleFunc("/follow/requests", controllers.GetFollowRequests).Methods("GET")
	router.HandleFunc("/follow/accept/{username}", controllers.AcceptFollow).Methods("POST")
	router.HandleFunc("/follow/decline/{username}", controllers.DeclineFollow).Methods("POST")
	router.HandleFunc("/follow/{username}", controllers.FollowUser).Methods("POST")
	
	router.HandleFunc("/unfollow/{username}", controllers.UnfollowUser).Methods("POST")
	router.HandleFunc("/followers/{username}", controllers.GetFollowers).Methods("GET")
	router.HandleFunc("/following", controllers.GetFollowing).Methods("GET")
	router.HandleFunc("/followers/{username}", controllers.BanFollower).Methods("DELETE")
	router.HandleFunc("/duress", controllers.PostDuress).Methods("POST")
	router.HandleFunc("/duress/cancel", controllers.CancelDuress).Methods("POST")
	router.HandleFunc("/duress/cancel", controllers.CancelDuress).Methods("POST")
	router.HandleFunc("/users/map", controllers.GetDuressMap).Methods("GET")
	router.HandleFunc("/duress/following", controllers.GetFollowingDuress).Methods("GET")
	router.HandleFunc("/duress/verify", controllers.VerifyAccess).Methods("POST")
	router.HandleFunc("/duress/dismiss/{username}", controllers.DismissDuressNotification).Methods("POST")

	// Admin Routes
	router.HandleFunc("/admin/users/{username}", controllers.AdminDeregisterUser).Methods("DELETE")

	// Start Background Jobs
	go func() {
		log.Println("Starting background inactivity checker...")
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run once on startup
		if err := services.CheckInactivity(); err != nil {
			log.Printf("Error running initial inactivity check: %v", err)
		}

		for range ticker.C {
			if err := services.CheckInactivity(); err != nil {
				log.Printf("Error running scheduled inactivity check: %v", err)
			}
		}
	}()

	log.Print("Attempting app start")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT is not set
	}

	log.Println("Starting server on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
