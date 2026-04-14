package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func Invite(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	if p.IsDuress {
		response := map[string]any{
			"message":    "Invite code created successfully",
			"inviteCode": services.GetDummyInviteCode(),
		}
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	inviteCode, err := services.CreateInvite(p.Username)
	if err != nil {
		slog.Error("create invite failed", slog.Any("err", err))
		if errors.Is(err, services.ErrRateLimitExceeded) {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"message":    "Invite code created successfully",
		"inviteCode": inviteCode,
	}
	_ = json.NewEncoder(w).Encode(response)
}
