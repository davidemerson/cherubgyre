package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type DuressRequest struct {
	DuressType     string                 `json:"duress_type"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	AdditionalData map[string]interface{} `json:"additional_data"`
	DuressPin      string                 `json:"duress_pin"`
}

func PostDuress(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	var duressRequest DuressRequest
	if err := json.NewDecoder(r.Body).Decode(&duressRequest); err != nil {
		log.Printf("Error decoding duress request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := services.PostDuress(token, duressRequest.DuressType, duressRequest.Message, duressRequest.Timestamp, duressRequest.AdditionalData, duressRequest.DuressPin)
	if err != nil {
		log.Printf("Error posting duress: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Duress posted successfully"}
	json.NewEncoder(w).Encode(response)
}

func CancelDuress(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	err := services.CancelDuress(token)
	if err != nil {
		log.Printf("Error canceling duress: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Duress canceled successfully"}
	json.NewEncoder(w).Encode(response)
}

func GetDuressMap(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	duressMap, err := services.GetDuressMap(token)
	if err != nil {
		log.Printf("Error getting duress map: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(duressMap)
}
