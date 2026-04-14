package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type DuressRequest struct {
	DuressType     string                 `json:"duress_type"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	AdditionalData map[string]interface{} `json:"additional_data"`
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
		switch err.Error() {
		case "invalid credentials":
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case "duress rate limit exceeded":
			http.Error(w, err.Error(), http.StatusTooManyRequests)
		default:
			log.Printf("Error posting duress: %v", err)
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
		if err.Error() == "invalid credentials" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Printf("Error canceling duress: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Duress canceled successfully"})
}

func GetDuressMap(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	duressMap, err := services.GetDuressMap(p.Username)
	if err != nil {
		log.Printf("Error getting duress map: %v", err)
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
		_ = json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	duresses, err := services.GetFollowingDuress(p.Username)
	if err != nil {
		log.Printf("Error getting following duress: %v", err)
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
		log.Printf("Error dismissing notification (unfollowing %s): %v", target, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Notification dismissed"})
}
