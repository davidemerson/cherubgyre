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
	newHistory, err := CheckRateLimit(user.InviteGenerationHistory, now, 5, 168*3600)
	if err != nil {
		return "", err
	}
	user.InviteGenerationHistory = newHistory

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

func CheckRateLimit(history []int64, now int64, limit int, windowSeconds int64) ([]int64, error) {
	var validHistory []int64
	for _, timestamp := range history {
		if now-timestamp < windowSeconds {
			validHistory = append(validHistory, timestamp)
		}
	}

	if len(validHistory) >= limit {
		return nil, errors.New("rate limit exceeded")
	}

	validHistory = append(validHistory, now)
	return validHistory, nil
}
