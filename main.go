package main

import (
	"cherubgyre/controllers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
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
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
