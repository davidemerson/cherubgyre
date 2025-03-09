package services

import (
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

func GetFollowers(username string) ([]string, error) {
	followers, err := repositories.GetFollowers(username)
	if err != nil {
		log.Println("Error getting followers:", err)
		return nil, err
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
