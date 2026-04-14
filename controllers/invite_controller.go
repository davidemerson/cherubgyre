package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"log"
	"net/http"
)

func Invite(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	if p.IsDuress {
		response := map[string]interface{}{
			"message":    "Invite code created successfully",
			"inviteCode": services.GetDummyInviteCode(),
		}
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	inviteCode, err := services.CreateInvite(p.Username)
	if err != nil {
		log.Printf("Error creating invite: %v", err)
		if err.Error() == "rate limit exceeded" {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":    "Invite code created successfully",
		"inviteCode": inviteCode,
	}
	_ = json.NewEncoder(w).Encode(response)
}
