package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"log"
)

// The follow service now takes already-authenticated usernames from the
// controllers (via the RequireAuth middleware) instead of raw tokens.
// This collapses the layer of per-function ValidateToken + GetUsernameFromToken
// calls that were being done inconsistently at the service layer.

func FollowUser(followerUsername, targetUsername string) error {
	if err := repositories.AddFollower(followerUsername, targetUsername, "pending"); err != nil {
		log.Println("Error adding follow request:", err)
		return err
	}
	return nil
}

func AcceptFollow(currentUsername, requesterUsername string) error {
	if err := repositories.AcceptFollower(requesterUsername, currentUsername); err != nil {
		log.Println("Error accepting follower:", err)
		return err
	}
	return nil
}

func DeclineFollow(currentUsername, requesterUsername string) error {
	if err := repositories.RemoveFollower(requesterUsername, currentUsername); err != nil {
		log.Println("Error declining follower:", err)
		return err
	}
	return nil
}

func UnfollowUser(followerUsername, targetUsername string) error {
	if err := repositories.RemoveFollower(followerUsername, targetUsername); err != nil {
		log.Println("Error removing follower:", err)
		return err
	}
	return nil
}

func GetFollowRequests(username string) ([]dtos.UserResponseDTO, error) {
	requestUsernames, err := repositories.GetFollowRequests(username)
	if err != nil {
		log.Println("Error getting follow requests:", err)
		return nil, err
	}
	return lookupUsers(requestUsernames), nil
}

func GetFollowers(username string) ([]dtos.UserResponseDTO, error) {
	followerUsernames, err := repositories.GetFollowers(username)
	if err != nil {
		log.Println("Error getting followers:", err)
		return nil, err
	}
	return lookupUsers(followerUsernames), nil
}

func BanFollower(currentUsername, targetUsername string) error {
	if err := repositories.BanFollower(targetUsername, currentUsername); err != nil {
		log.Println("Error banning follower:", err)
		return err
	}
	return nil
}

func GetFollowing(username string) ([]dtos.UserResponseDTO, error) {
	followingUsernames, err := repositories.GetFollowing(username)
	if err != nil {
		log.Println("Error getting following list:", err)
		return nil, err
	}
	return lookupUsers(followingUsernames), nil
}

// lookupUsers enriches a list of usernames with avatar info. Missing
// users are skipped rather than errored — a deleted account in someone's
// follower list shouldn't break the whole response.
func lookupUsers(usernames []string) []dtos.UserResponseDTO {
	out := make([]dtos.UserResponseDTO, 0, len(usernames))
	for _, name := range usernames {
		user, err := repositories.GetUserByID(name)
		if err != nil {
			continue
		}
		out = append(out, dtos.UserResponseDTO{
			Username: user.Username,
			Avatar:   user.Avatar,
		})
	}
	return out
}
