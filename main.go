package main

import (
	"cherubgyre/controllers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("You've reached cherubgyre"))
	}).Methods("GET")
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/health", controllers.Health).Methods("GET")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/profile", controllers.Profile).Methods("GET")
	router.HandleFunc("/invite", controllers.Invite).Methods("GET")
	router.HandleFunc("/follow/{username}", controllers.FollowUser).Methods("POST")
	router.HandleFunc("/unfollow/{username}", controllers.UnfollowUser).Methods("POST")
	router.HandleFunc("/followers/{username}", controllers.GetFollowers).Methods("GET")
	router.HandleFunc("/followers/{username}", controllers.BanFollower).Methods("DELETE")

	log.Print("Attempting app start")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if PORT is not set
	}

	log.Println("Starting server on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
