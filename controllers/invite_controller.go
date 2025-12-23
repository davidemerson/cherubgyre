package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"net/http"
)

func Invite(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Check if token is in duress mode
	if services.IsDuressToken(token) {
		// Return dummy invite code
		response := map[string]interface{}{
			"message":    "Invite code created successfully",
			"inviteCode": "DUMMY-INVITE-CODE-000",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	inviteCode, err := services.CreateInvite(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":    "Invite code created successfully",
		"inviteCode": inviteCode,
	}
	json.NewEncoder(w).Encode(response)
}
