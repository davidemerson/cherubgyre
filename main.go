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
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
