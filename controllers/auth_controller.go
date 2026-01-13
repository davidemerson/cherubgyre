package controllers

import (
	"cherubgyre/dtos"
	"cherubgyre/services"
	"encoding/json"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var request dtos.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	response, err := services.Login(request)
	if err != nil {
		// Pass the error message directly to the client
		// This allows sending "X attempts remaining" or "Account locked" messages
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(response)
}

func Profile(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	valid, err := services.ValidateToken(token)
	if err != nil || !valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if token is in duress mode
	if services.IsDuressToken(token) {
		// Return dummy profile data
		response := map[string]interface{}{
			"username": "gst_001",
			"avatar":   "https://api.dicebear.com/7.x/identicon/svg?seed=gst_001&rowColor=000000",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	username, err := services.GetUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	user, err := services.GetUserProfile(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"username": user.Username,
		"avatar":   user.Avatar,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ChangePin(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	username, err := services.GetUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var request struct {
		CurrentPin string `json:"current_pin"`
		NewPin     string `json:"new_pin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.NewPin == "" || len(request.NewPin) < 4 { // Basic validation
		http.Error(w, "New PIN must be at least 4 characters", http.StatusBadRequest)
		return
	}

	err = services.ChangePin(username, request.CurrentPin, request.NewPin)
	if err != nil {
		if err.Error() == "incorrect current pin" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else if err.Error() == "new pin cannot be the same as your duress pin" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "PIN changed successfully"}
	json.NewEncoder(w).Encode(response)
}

func ChangeDuressPin(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	username, err := services.GetUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var request struct {
		CurrentPin string `json:"current_pin"`
		NewPin     string `json:"new_pin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.NewPin == "" || len(request.NewPin) < 4 { // Basic validation
		http.Error(w, "New Duress PIN must be at least 4 characters", http.StatusBadRequest)
		return
	}

	err = services.ChangeDuressPin(username, request.CurrentPin, request.NewPin)
	if err != nil {
		if err.Error() == "incorrect current duress pin" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else if err.Error() == "new duress pin cannot be the same as your normal pin" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Duress PIN changed successfully"}
	json.NewEncoder(w).Encode(response)
}
