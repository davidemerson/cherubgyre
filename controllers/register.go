package controllers

import (
	"cherubgyre/dtos"
	"cherubgyre/services"
	"encoding/json"
	"log/slog"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var registerDTO dtos.RegisterDTO
	err := json.NewDecoder(r.Body).Decode(&registerDTO)
	if err != nil {
		slog.Warn("register decode failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, user, err := services.RegisterUser(registerDTO)
	if err != nil {
		slog.Error("register user failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("user registered via api", slog.String("user", user.Username))
	// Never echo PIN material back to the caller, even though the service
	// already clears the hash fields.
	user.NormalPin = ""
	user.DuressPin = ""
	user.NormalPinHash = ""
	user.DuressPinHash = ""
	w.WriteHeader(http.StatusCreated)
	response := map[string]any{
		"message": message,
		"user":    user,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(response)
}
