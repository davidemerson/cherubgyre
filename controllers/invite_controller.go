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
