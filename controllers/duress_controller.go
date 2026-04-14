package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type DuressRequest struct {
	DuressType     string                 `json:"duress_type"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	AdditionalData map[string]any `json:"additional_data"`
	DuressPin      string                 `json:"duress_pin"`
}

func PostDuress(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	var req DuressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := services.PostDuress(p.Username, req.DuressType, req.Message, req.Timestamp, req.AdditionalData, req.DuressPin)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.Is(err, services.ErrDuressRateLimited):
			http.Error(w, err.Error(), http.StatusTooManyRequests)
		default:
			slog.Error("post duress failed", slog.Any("err", err))
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Duress posted successfully"})
}

// CancelDuress requires the normal PIN in the request body. Per the
// threat model, a coercer who already holds a duress-mode session must
// not be able to cancel the silent alert they caused.
func CancelDuress(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	var req struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Pin == "" {
		http.Error(w, "pin is required", http.StatusBadRequest)
		return
	}

	if err := services.CancelDuress(p.Username, req.Pin); err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		slog.Error("cancel duress failed", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Duress canceled successfully"})
}

func GetDuressMap(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	duressMap, err := services.GetDuressMap(p.Username)
	if err != nil {
		slog.Error("get duress map failed", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(duressMap)
}

func GetFollowingDuress(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	// Duress mode hides real duress signals entirely — a coercer looking
	// at this screen must not see who else is in trouble.
	if p.IsDuress {
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}

	duresses, err := services.GetFollowingDuress(p.Username)
	if err != nil {
		slog.Error("get following duress failed", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(duresses)
}

func VerifyAccess(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	var req struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := services.VerifyDuressPin(p.Username, req.Pin); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Access granted"})
}

// DismissDuressNotification removes the follow relationship for a user
// who has left the service. Semantically equivalent to unfollow since the
// target account no longer exists — the client-side notification goes
// away once the relationship is gone.
func DismissDuressNotification(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	target := vars["username"]

	if err := services.UnfollowUser(p.Username, target); err != nil {
		slog.Error("dismiss duress notification failed",
			slog.String("target", sanitizeForLog(target)), slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Notification dismissed"})
}
