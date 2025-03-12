package services

import (
	"cherubgyre/repositories"
	"errors"
	"log"
	"time"
)

func PostDuress(token, duressType, message string, timestamp time.Time, additionalData map[string]interface{}) error {
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return errors.New("invalid token")
	}

	username, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return err
	}

	err = repositories.SaveDuress(username, duressType, message, timestamp, additionalData)
	if err != nil {
		log.Println("Error saving duress:", err)
		return err
	}

	return nil
}

func CancelDuress(token string) error {
	valid, err := ValidateToken(token)
	if err != nil || !valid {
		log.Println("Invalid token:", token)
		return errors.New("invalid token")
	}

	username, err := GetUsernameFromToken(token)
	if err != nil {
		log.Println("Error getting username from token:", err)
		return err
	}

	err = repositories.DeleteDuress(username)
	if err != nil {
		log.Println("Error deleting duress:", err)
		return err
	}

	return nil
}

func GetDuressMap(token string) (map[string]interface{}, error) {
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

	duressMap, err := repositories.GetDuressMap(username)
	if err != nil {
		log.Println("Error getting duress map:", err)
		return nil, err
	}

	return duressMap, nil
}
