package services

import (
	"cherubgyre/repositories"
	"errors"
	"log"
	"time"

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

	// Rate limiting logic
	now := time.Now().Unix()
	var validHistory []int64
	for _, timestamp := range user.InviteGenerationHistory {
		if now-timestamp < 3600 { // 3600 seconds = 1 hour
			validHistory = append(validHistory, timestamp)
		}
	}

	if len(validHistory) >= 3 {
		return "", errors.New("rate limit exceeded: max 3 invites per hour")
	}

	validHistory = append(validHistory, now)
	user.InviteGenerationHistory = validHistory

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
