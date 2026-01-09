package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("your_secret_key")

type Claims struct {
	UserID   string `json:"user_id"`
	IsDuress bool   `json:"is_duress"`
	jwt.StandardClaims
}

func Login(request dtos.LoginRequest) (dtos.LoginResponse, error) {
	log.Println("Login attempt for user:", request.Username)
	
	// Get user first to check/update status
	user, err := repositories.GetUserByID(request.Username)
	if err != nil {
		// User not found - generic error
		log.Println("User not found:", request.Username)
		return dtos.LoginResponse{}, errors.New("invalid credentials")
	}

	pinType, err := repositories.ValidateUserCredentials(request.Username, request.PIN)
	
	// Handle Failed Login (Launch Lock)
	if err != nil || pinType == 0 {
		log.Println("Invalid credentials for user:", request.Username)
		
		// Increment failed attempts
		user.FailedAttempts++
		log.Printf("User %s failed attempt %d/10", user.Username, user.FailedAttempts)
		
		if user.FailedAttempts >= 10 {
			// Trigger Launch Lock Deregistration
			log.Printf("User %s exceeded Launch Lock limit. Deregistering...", user.Username)
			DeregisterUser(user.Username, "Launch Lock (10 failed PIN attempts)")
			return dtos.LoginResponse{}, errors.New("account has been locked and removed due to excessive failed attempts")
		}
		
		// Save the failed attempt count
		repositories.UpdateUser(user)
		
		remainingAttempts := 10 - user.FailedAttempts
		errorMessage := fmt.Sprintf("Invalid credentials. %d attempts remaining before account deletion.", remainingAttempts)
		return dtos.LoginResponse{}, errors.New(errorMessage)
	}

	// Handle Successful Login
	// Reset failed attempts and update last active
	user.FailedAttempts = 0
	user.LastActive = time.Now()
	repositories.UpdateUser(user)

	// Handle based on PIN type
	switch pinType {
	case 1:
		// Normal PIN - Cancel any active duress signal
		log.Println("Normal PIN login - checking for active duress signals")
		// We use repositories directly here to avoid circular dep if needed, but same package is fine
		err := repositories.DeleteDuress(request.Username)
		if err != nil {
			log.Printf("Note: Error canceling duress (may not exist): %v", err)
			// Don't fail login if duress deletion fails - user may not have active duress
		}
	case 2:
		// Duress PIN - Create silent duress signal
		log.Println("Duress PIN login - creating silent duress signal")
		err := repositories.SaveDuress(
			request.Username,
			"Silent Login",
			"Duress initiated via Login Screen",
			time.Now(),
			request.AdditionalData,
		)
		if err != nil {
			log.Printf("Error creating duress signal: %v", err)
			// Continue with login even if duress creation fails
		}
	}

	// Generate JWT token for both cases
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   request.Username,
		IsDuress: pinType == 2, // Set to true if Duress PIN was used
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("Error signing token:", err)
		return dtos.LoginResponse{}, err
	}

	log.Println("Login successful for user:", request.Username)
	return dtos.LoginResponse{Token: tokenString}, nil
}

func ValidateToken(tokenStr string) (bool, error) {
	log.Println("Validating token")

	// Remove 'bearer ' prefix if it exists
	if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		log.Println("Error parsing token:", err)
		return false, err
	}
	if !token.Valid {
		log.Println("Invalid token")
		return false, errors.New("invalid token")
	}

	log.Println("Token is valid for user:", claims.UserID)
	return true, nil
}

func GetUsernameFromToken(tokenStr string) (string, error) {
	if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

// GetUserProfile returns user info for a given username
func GetUserProfile(username string) (dtos.RegisterDTO, error) {
	return repositories.GetUserByID(username)
}

// IsDuressToken checks if a token is in duress mode
func IsDuressToken(tokenStr string) bool {
	if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	return claims.IsDuress
}

func ChangePin(username, currentPin, newPin string) error {
	user, err := repositories.GetUserByID(username)
	if err != nil {
		return err
	}

	if user.NormalPin != currentPin {
		return errors.New("incorrect current pin")
	}

	if user.DuressPin == newPin {
		return errors.New("new pin cannot be the same as your duress pin")
	}

	user.NormalPin = newPin
	return repositories.UpdateUser(user)
}

func ChangeDuressPin(username, currentPin, newPin string) error {
	user, err := repositories.GetUserByID(username)
	if err != nil {
		return err
	}

	if user.DuressPin != currentPin {
		return errors.New("incorrect current duress pin")
	}

	if user.NormalPin == newPin {
		return errors.New("new duress pin cannot be the same as your normal pin")
	}

	user.DuressPin = newPin
	return repositories.UpdateUser(user)
}
