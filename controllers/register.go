package controllers

import (
	"cherubgyre/dtos"
	"cherubgyre/services"
	"encoding/json"
	"log"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var registerDTO dtos.RegisterDTO
	err := json.NewDecoder(r.Body).Decode(&registerDTO)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, user, err := services.RegisterUser(registerDTO)
	if err != nil {
		log.Printf("Error registering user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("User registered successfully: %+v", user)
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"message": message,
		"user":    user,
	}
	json.NewEncoder(w).Encode(response)
}
