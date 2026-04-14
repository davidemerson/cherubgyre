package controllers

import (
	"cherubgyre/dtos"
	"cherubgyre/services"
	"encoding/json"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := services.LoginLimiter.Allow(services.ClientIP(r)); err != nil {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var request dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	response, err := services.Login(request)
	if err != nil {
		// Single opaque error for every failure mode so the endpoint
		// cannot be used to enumerate usernames or distinguish causes.
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Profile returns the authenticated user's profile, or a seeded-random
// dummy profile if the caller is in duress mode.
func Profile(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	if p.IsDuress {
		dummy := services.GetDummyProfile(p.Username)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dummy)
		return
	}

	user, err := services.GetUserProfile(p.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"username": user.Username,
		"avatar":   user.Avatar,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func ChangePin(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	var request struct {
		CurrentPin string `json:"current_pin"`
		NewPin     string `json:"new_pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.ValidatePin(request.NewPin); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.ChangePin(p.Username, request.CurrentPin, request.NewPin); err != nil {
		switch err.Error() {
		case "incorrect current pin":
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case "new pin cannot be the same as your duress pin":
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "PIN changed successfully"})
}

func ChangeDuressPin(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	var request struct {
		CurrentPin string `json:"current_pin"`
		NewPin     string `json:"new_pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.ValidatePin(request.NewPin); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.ChangeDuressPin(p.Username, request.CurrentPin, request.NewPin); err != nil {
		switch err.Error() {
		case "incorrect current duress pin":
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case "new duress pin cannot be the same as your normal pin":
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Duress PIN changed successfully"})
}
