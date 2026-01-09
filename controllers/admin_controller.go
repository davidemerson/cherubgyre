package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func AdminDeregisterUser(w http.ResponseWriter, r *http.Request) {
	// Ideally, add admin authentication here.
	// For now, we assume the route is protected or internal.
	
	vars := mux.Vars(r)
	username := vars["username"]

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	log.Printf("Admin Request: Deregister user %s", username)

	err := services.DeregisterUser(username, "Admin Manual Deregistration")
	if err != nil {
		log.Printf("Error deregistering user: %v", err)
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "User successfully deregistered"}
	json.NewEncoder(w).Encode(response)
}
