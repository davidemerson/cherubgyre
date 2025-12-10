package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func FollowUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	username := vars["username"]

	err := services.FollowUser(token, username)
	if err != nil {
		log.Printf("Error following user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "User followed successfully"}
	json.NewEncoder(w).Encode(response)
}

func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	username := vars["username"]

	err := services.UnfollowUser(token, username)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "User unfollowed successfully"}
	json.NewEncoder(w).Encode(response)
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	followers, err := services.GetFollowers(username)
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(followers)
}

func BanFollower(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	username := vars["username"]

	err := services.BanFollower(token, username)
	if err != nil {
		log.Printf("Error banning follower: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Follower banned successfully"}
	json.NewEncoder(w).Encode(response)
}

func GetFollowing(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	following, err := services.GetFollowing(token)
	if err != nil {
		log.Printf("Error getting following list: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(following)
}
