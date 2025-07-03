package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
)

const (
	charset       = "abcdefghijklmnopqrstuvwxyz"
	usernameParts = 3
)

func generateRandomString(length int, charSet string) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			return "", err
		}
		result[i] = charSet[num.Int64()]
	}
	return string(result), nil
}

func generateRandomDigits(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(10)) // 0-9
		if err != nil {
			return "", err
		}
		result[i] = byte(num.Int64() + '0')
	}
	return string(result), nil
}

func generateUsername() (string, error) {
	letters, err := generateRandomString(usernameParts, charset)
	if err != nil {
		return "", fmt.Errorf("failed to generate random letters: %w", err)
	}
	digits, err := generateRandomDigits(usernameParts)
	if err != nil {
		return "", fmt.Errorf("failed to generate random digits: %w", err)
	}
	return fmt.Sprintf("%s_%s", letters, digits), nil
}

func RegisterUser(registerDTO dtos.RegisterDTO) (string, dtos.RegisterDTO, error) {
	// Username is no longer expected from the DTO input for validation here
	if registerDTO.NormalPin == "" || registerDTO.DuressPin == "" || registerDTO.InviteCode == "" {
		log.Println("Validation failed: normal_pin, duress_pin, and invite_code are required")
		return "", dtos.RegisterDTO{}, errors.New("normal_pin, duress_pin, and invite_code are required")
	}

	if registerDTO.DuressPin == registerDTO.NormalPin {
		log.Println("Validation failed: duress_pin and normal_pin cannot be the same")
		return "", dtos.RegisterDTO{}, errors.New("duress_pin and normal_pin cannot be the same")
	}

	var username string
	var err error
	var taken bool

	for {
		username, err = generateUsername()
		if err != nil {
			log.Printf("Error generating username: %v", err)
			return "", dtos.RegisterDTO{}, fmt.Errorf("failed to generate username: %w", err)
		}

		taken, err = repositories.IsUsernameTaken(username)
		if err != nil {
			log.Printf("Error checking if username is taken: %v", err)
			return "", dtos.RegisterDTO{}, fmt.Errorf("failed to check username uniqueness: %w", err)
		}
		if !taken {
			break // Unique username found
		}
		log.Printf("Username '%s' is already taken. Generating a new one.", username)
	}

	registerDTO.Username = username // Set the generated username in the DTO

	err = repositories.SaveUser(registerDTO) // Pass the DTO with the username
	if err != nil {
		log.Printf("Error saving user: %v", err)
		return "", dtos.RegisterDTO{}, err
	}

	log.Printf("User saved successfully: %+v", registerDTO)
	return "User registered successfully", registerDTO, nil // Return the DTO with the generated username
}
