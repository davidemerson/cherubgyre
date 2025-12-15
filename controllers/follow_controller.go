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
	response := map[string]string{"message": "Follow request sent successfully"}
	json.NewEncoder(w).Encode(response)
}

func AcceptFollow(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	username := vars["username"]

	err := services.AcceptFollow(token, username)
	if err != nil {
		log.Printf("Error accepting follower: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Follower accepted successfully"}
	json.NewEncoder(w).Encode(response)
}

func DeclineFollow(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	vars := mux.Vars(r)
	username := vars["username"]

	err := services.DeclineFollow(token, username)
	if err != nil {
		log.Printf("Error declining follower: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Follow request declined successfully"}
	json.NewEncoder(w).Encode(response)
}

func GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	requests, err := services.GetFollowRequests(token)
	if err != nil {
		log.Printf("Error getting follow requests: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requests)
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
