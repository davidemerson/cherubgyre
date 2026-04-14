package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPin produces a bcrypt hash suitable for storing in NormalPinHash or
// DuressPinHash. Exposed so the migration job can reuse the same cost.
func HashPin(pin string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pin), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// pickWord returns a cryptographically random element from the given slice.
func pickWord(words []string) (string, error) {
	if len(words) == 0 {
		return "", errors.New("empty wordlist")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		return "", err
	}
	return words[n.Int64()], nil
}

// generateUsername produces a spec-format username of the shape
// "angel-type-city" (e.g., "cherub-gyre-chicago"). Each component is
// drawn from an embedded wordlist using crypto/rand. The combined space
// is much larger and less pattern-matchable than the legacy
// "{3 letters}_{3 digits}" format and is thus materially harder to
// enumerate, which closes the Launch Lock DoS vector from the plan.
func generateUsername() (string, error) {
	a, err := pickWord(angelWords)
	if err != nil {
		return "", fmt.Errorf("generate username: %w", err)
	}
	t, err := pickWord(typeWords)
	if err != nil {
		return "", fmt.Errorf("generate username: %w", err)
	}
	c, err := pickWord(cityWords)
	if err != nil {
		return "", fmt.Errorf("generate username: %w", err)
	}
	return a + "-" + t + "-" + c, nil
}

func RegisterUser(registerDTO dtos.RegisterDTO) (string, dtos.RegisterDTO, error) {
	if registerDTO.NormalPin == "" || registerDTO.DuressPin == "" || registerDTO.InviteCode == "" {
		return "", dtos.RegisterDTO{}, errors.New("normal_pin, duress_pin, and invite_code are required")
	}
	if err := ValidatePin(registerDTO.NormalPin); err != nil {
		return "", dtos.RegisterDTO{}, err
	}
	if err := ValidatePin(registerDTO.DuressPin); err != nil {
		return "", dtos.RegisterDTO{}, err
	}
	if registerDTO.DuressPin == registerDTO.NormalPin {
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

	registerDTO.Avatar = "https://api.dicebear.com/7.x/identicon/svg?seed=" + username + "&rowColor=000000"
	registerDTO.Username = username

	// Assign a UUIDv7 UID at creation time (spec requirement). Usernames
	// can be recycled when a deregistered slot is freed, but a UID is
	// immutable so relationships can remain stable even after recycling.
	uid, err := uuid.NewV7()
	if err != nil {
		return "", dtos.RegisterDTO{}, fmt.Errorf("failed to allocate uid: %w", err)
	}
	registerDTO.UID = uid.String()

	normalHash, err := HashPin(registerDTO.NormalPin)
	if err != nil {
		return "", dtos.RegisterDTO{}, fmt.Errorf("failed to hash normal pin: %w", err)
	}
	duressHash, err := HashPin(registerDTO.DuressPin)
	if err != nil {
		return "", dtos.RegisterDTO{}, fmt.Errorf("failed to hash duress pin: %w", err)
	}
	registerDTO.NormalPinHash = normalHash
	registerDTO.DuressPinHash = duressHash
	registerDTO.NormalPin = ""
	registerDTO.DuressPin = ""

	if err := repositories.SaveUser(registerDTO); err != nil {
		log.Printf("Error saving user: %v", err)
		return "", dtos.RegisterDTO{}, err
	}

	log.Printf("User saved successfully: %s", registerDTO.Username)
	// Never echo PIN material back to the caller.
	response := registerDTO
	response.NormalPinHash = ""
	response.DuressPinHash = ""
	return "User registered successfully", response, nil
}
