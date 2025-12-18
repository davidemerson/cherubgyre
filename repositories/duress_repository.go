package repositories

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Duress struct {
	Username       string                 `json:"username"`
	DuressType     string                 `json:"duress_type"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	AdditionalData map[string]interface{} `json:"additional_data"`
}

func SaveDuress(username, duressType, message string, timestamp time.Time, additionalData map[string]interface{}) error {
	log.Printf("Saving duress for user: %s", username)
	file, err := os.OpenFile("duress.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var duresses []Duress
	if err := json.NewDecoder(file).Decode(&duresses); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding duress data: %v", err)
		return err
	}

	duresses = append(duresses, Duress{
		Username:       username,
		DuressType:     duressType,
		Message:        message,
		Timestamp:      timestamp,
		AdditionalData: additionalData,
	})

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(duresses); err != nil {
		log.Printf("Error encoding duress data: %v", err)
		return err
	}

	log.Printf("Successfully saved duress for user: %s", username)
	return nil
}

func DeleteDuress(username string) error {
	log.Printf("Deleting duress for user: %s", username)
	file, err := os.OpenFile("duress.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var duresses []Duress
	if err := json.NewDecoder(file).Decode(&duresses); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding duress data: %v", err)
		return err
	}

	for i, duress := range duresses {
		if duress.Username == username {
			duresses = append(duresses[:i], duresses[i+1:]...)
			break
		}
	}

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(duresses); err != nil {
		log.Printf("Error encoding duress data: %v", err)
		return err
	}

	log.Printf("Successfully deleted duress for user: %s", username)
	return nil
}

func GetDuressMap(username string) (map[string]interface{}, error) {
	log.Printf("Getting duress map for user: %s", username)
	file, err := os.OpenFile("duress.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	var duresses []Duress
	if err := json.NewDecoder(file).Decode(&duresses); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding duress data: %v", err)
		return nil, err
	}

	duressMap := make(map[string]interface{})
	for _, duress := range duresses {
		if duress.Username == username {
			duressMap[duress.Username] = duress
		}
	}

	log.Printf("Successfully retrieved duress map for user: %s", username)
	return duressMap, nil
}

func GetActiveDuressForUsers(usernames []string) ([]Duress, error) {
	log.Printf("Getting active duress for users: %v", usernames)
	file, err := os.OpenFile("duress.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	var duresses []Duress
	if err := json.NewDecoder(file).Decode(&duresses); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding duress data: %v", err)
		return nil, err
	}

	var activeDuresses []Duress
	for _, duress := range duresses {
		if contains(usernames, duress.Username) {
			activeDuresses = append(activeDuresses, duress)
		}
	}

	log.Printf("Successfully retrieved %d active duress signals", len(activeDuresses))
	return activeDuresses, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
