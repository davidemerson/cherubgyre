package services

import (
	"cherubgyre/repositories"
	"errors"
	"log"

	"github.com/google/uuid"
)

func CreateInvite(token string) (string, error) {
	log.Println("CreateInvite called with token:", token)
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return "", errors.New("invalid token")
	}

	username, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return "", err
	}

	user, err := repositories.GetUserByID(username)
	if err != nil {
		log.Println("Error getting user by ID:", err)
		return "", err
	}

	inviteCode := uuid.New().String()
	user.UserInviteCode = inviteCode

	err = repositories.UpdateUser(user)
	if err != nil {
		log.Println("Error updating user:", err)
		return "", err
	}

	log.Println("Invite code created successfully:", inviteCode)
	return inviteCode, nil
}
