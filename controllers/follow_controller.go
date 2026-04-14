package controllers

import (
	"cherubgyre/services"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

func FollowUser(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	target := vars["username"]

	if err := services.FollowUser(p.Username, target); err != nil {
		slog.Error("follow user failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Follow request sent successfully"})
}

func AcceptFollow(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	follower := vars["username"]

	if err := services.AcceptFollow(p.Username, follower); err != nil {
		slog.Error("accept follower failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Follower accepted successfully"})
}

func DeclineFollow(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	follower := vars["username"]

	if err := services.DeclineFollow(p.Username, follower); err != nil {
		slog.Error("decline follower failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Follow request declined successfully"})
}

func GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	// Duress mode: empty list is believable (no one is requesting to
	// follow you right now) and leaks no real follower graph data.
	if p.IsDuress {
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}

	requests, err := services.GetFollowRequests(p.Username)
	if err != nil {
		slog.Error("get follow requests failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(requests)
}

func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	target := vars["username"]

	if err := services.UnfollowUser(p.Username, target); err != nil {
		slog.Error("unfollow failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "User unfollowed successfully"})
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	username := vars["username"]

	if p.IsDuress {
		_ = json.NewEncoder(w).Encode(services.GetDummyFollowers(p.Username))
		return
	}

	followers, err := services.GetFollowers(username)
	if err != nil {
		slog.Error("get followers failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(followers)
}

func BanFollower(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)
	vars := mux.Vars(r)
	target := vars["username"]

	if err := services.BanFollower(p.Username, target); err != nil {
		slog.Error("ban follower failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Follower banned successfully"})
}

func GetFollowing(w http.ResponseWriter, r *http.Request) {
	p := Identity(r)

	if p.IsDuress {
		_ = json.NewEncoder(w).Encode(services.GetDummyFollowing(p.Username))
		return
	}

	following, err := services.GetFollowing(p.Username)
	if err != nil {
		slog.Error("get following failed", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(following)
}
