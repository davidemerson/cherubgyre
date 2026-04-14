package controllers

import (
	"cherubgyre/dtos"
	"cherubgyre/services"
	"encoding/json"
	"log/slog"
	"net/http"
)

func ValidateInviteCode(w http.ResponseWriter, r *http.Request) {
	var request dtos.InviteValidationRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		slog.Warn("validate-invite decode failed", slog.Any("err", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := services.ValidateInviteCode(request.InviteCode)
	if err != nil {
		slog.Error("validate invite failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if response.Valid {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	_ = json.NewEncoder(w).Encode(response)
}
