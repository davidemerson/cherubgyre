package main

import (
	"cherubgyre/controllers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := mux.NewRouter()
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/health", controllers.Health).Methods("GET")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/profile", controllers.Profile).Methods("GET")
	router.HandleFunc("/invite", controllers.Invite).Methods("GET")
	router.HandleFunc("/follow/{username}", controllers.FollowUser).Methods("POST")
	router.HandleFunc("/unfollow/{username}", controllers.UnfollowUser).Methods("POST")
	router.HandleFunc("/followers/{username}", controllers.GetFollowers).Methods("GET")
	router.HandleFunc("/followers/{username}", controllers.BanFollower).Methods("DELETE")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: No .env file found, relying on Heroku environment variables")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT is not set
	}

	log.Println("Starting server on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
