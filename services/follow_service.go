package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"errors"
	"log"
)

func FollowUser(token, username string) error {
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return errors.New("invalid token")
	}

	followerUsername, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return err
	}

	err = repositories.AddFollower(followerUsername, username)
	if err != nil {
		log.Println("Error adding follower:", err)
		return err
	}

	return nil
}

func UnfollowUser(token, username string) error {
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return errors.New("invalid token")
	}

	followerUsername, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return err
	}

	err = repositories.RemoveFollower(followerUsername, username)
	if err != nil {
		log.Println("Error removing follower:", err)
		return err
	}

	return nil
}

func GetFollowers(username string) ([]dtos.UserResponseDTO, error) {
	followerUsernames, err := repositories.GetFollowers(username)
	if err != nil {
		log.Println("Error getting followers:", err)
		return nil, err
	}

	var followers []dtos.UserResponseDTO
	for _, followerUsername := range followerUsernames {
		user, err := repositories.GetUserByID(followerUsername)
		if err != nil {
			log.Printf("Error getting user details for follower %s: %v", followerUsername, err)
			continue
		}
		followers = append(followers, dtos.UserResponseDTO{
			Username: user.Username,
			Avatar:   user.Avatar,
		})
	}

	return followers, nil
}

func BanFollower(token, username string) error {
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return errors.New("invalid token")
	}

	banningUsername, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return err
	}

	err = repositories.BanFollower(username, banningUsername)
	if err != nil {
		log.Println("Error banning follower:", err)
		return err
	}

	return nil
}

func GetFollowing(token string) ([]dtos.UserResponseDTO, error) {
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return nil, errors.New("invalid token")
	}

	username, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return nil, err
	}

	followingUsernames, err := repositories.GetFollowing(username)
	if err != nil {
		log.Println("Error getting following list:", err)
		return nil, err
	}

	var following []dtos.UserResponseDTO
	for _, followingUsername := range followingUsernames {
		user, err := repositories.GetUserByID(followingUsername)
		if err != nil {
			log.Printf("Error getting user details for following %s: %v", followingUsername, err)
			continue
		}
		following = append(following, dtos.UserResponseDTO{
			Username: user.Username,
			Avatar:   user.Avatar,
		})
	}

	return following, nil
}
