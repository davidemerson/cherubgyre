package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"errors"
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
	pinType, err := repositories.ValidateUserCredentials(request.Username, request.PIN)
	if err != nil {
		log.Println("Error validating user credentials:", err)
		return dtos.LoginResponse{}, errors.New("invalid credentials")
	}
	
	if pinType == 0 {
		log.Println("Invalid credentials for user:", request.Username)
		return dtos.LoginResponse{}, errors.New("invalid credentials")
	}

	// Handle based on PIN type
	switch pinType {
	case 1:
		// Normal PIN - Cancel any active duress signal
		log.Println("Normal PIN login - checking for active duress signals")
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
			map[string]interface{}{},
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
