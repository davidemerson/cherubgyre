package controllers

import (
	"cherubgyre/services"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// adminToken is installed by SetAdminToken at startup. An empty value
// causes every admin request to be rejected.
var adminToken string

// SetAdminToken installs the shared-secret admin token. Called from main().
func SetAdminToken(token string) {
	adminToken = token
}

// requireAdmin returns true and writes a 401 when the request does not
// present the expected X-Admin-Token header. Callers should return if it
// returns false.
func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if adminToken == "" {
		http.Error(w, "admin disabled", http.StatusServiceUnavailable)
		return false
	}
	got := r.Header.Get("X-Admin-Token")
	if subtle.ConstantTimeCompare([]byte(got), []byte(adminToken)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func AdminDeregisterUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

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
	_ = json.NewEncoder(w).Encode(response)
}
