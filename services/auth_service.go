package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtKey is set once at startup by SetJWTSecret. Requests fail closed if it
// is not configured so that a misconfigured deployment cannot accidentally
// sign tokens with an empty key.
var jwtKey []byte

// SetJWTSecret installs the signing key used by Login/ValidateToken/etc.
// Call this from main() before starting the HTTP server.
func SetJWTSecret(secret []byte) {
	jwtKey = secret
}

type Claims struct {
	UserID   string `json:"user_id"`
	IsDuress bool   `json:"is_duress"`
	jwt.RegisteredClaims
}

// errInvalidCredentials is the single opaque error returned for any login
// failure (unknown user, wrong PIN, missing pin, etc.) so callers cannot
// distinguish causes and enumerate the username space.
var errInvalidCredentials = errors.New("invalid credentials")

// MinPinLength is the minimum acceptable PIN length, applied uniformly to
// registration, change-pin, and change-duress-pin. Kept as a package-level
// constant so all three sites agree on the same rule.
const MinPinLength = 4

// ValidatePin enforces the shared PIN policy. Today just a length check;
// future rules (no sequential digits, no repeats, etc.) go here.
func ValidatePin(pin string) error {
	if len(pin) < MinPinLength {
		return fmt.Errorf("PIN must be at least %d characters", MinPinLength)
	}
	return nil
}

func Login(request dtos.LoginRequest) (dtos.LoginResponse, error) {
	if len(jwtKey) == 0 {
		return dtos.LoginResponse{}, errors.New("server not configured")
	}

	log.Println("Login attempt for user:", request.Username)

	user, err := repositories.GetUserByID(request.Username)
	if err != nil {
		// Unknown user: return the same opaque error as a wrong PIN so the
		// /login endpoint cannot be used to enumerate existing usernames.
		return dtos.LoginResponse{}, errInvalidCredentials
	}

	pinType, err := repositories.ValidateUserCredentials(request.Username, request.PIN)
	if err != nil || pinType == 0 {
		user.FailedAttempts++
		log.Printf("User %s failed attempt %d/10", user.Username, user.FailedAttempts)

		if user.FailedAttempts >= 10 {
			// Launch Lock: intentional per threat model. Deregister the user
			// completely after 10 failed attempts.
			log.Printf("User %s exceeded Launch Lock limit. Deregistering...", user.Username)
			_ = DeregisterUser(user.Username, "Launch Lock (10 failed PIN attempts)")
			return dtos.LoginResponse{}, errInvalidCredentials
		}

		if err := repositories.UpdateUser(user); err != nil {
			log.Printf("Failed to persist failed-attempt counter: %v", err)
		}
		return dtos.LoginResponse{}, errInvalidCredentials
	}

	user.FailedAttempts = 0
	user.LastActive = time.Now()
	if err := repositories.UpdateUser(user); err != nil {
		log.Printf("Failed to persist successful login state: %v", err)
	}

	switch pinType {
	case 1:
		log.Println("Normal PIN login - preserving any active duress signals")
	case 2:
		log.Println("Duress PIN login - creating silent duress signal")
		if err := repositories.SaveDuress(
			request.Username,
			"Silent Login",
			"Duress initiated via Login Screen",
			time.Now(),
			request.AdditionalData,
		); err != nil {
			log.Printf("Error creating duress signal: %v", err)
		}
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   request.Username,
		IsDuress: pinType == 2,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
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

func parseToken(tokenStr string) (*Claims, error) {
	if len(jwtKey) == 0 {
		return nil, errors.New("server not configured")
	}
	tokenStr = strings.TrimSpace(tokenStr)
	if strings.HasPrefix(strings.ToLower(tokenStr), "bearer ") {
		tokenStr = tokenStr[7:]
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func ValidateToken(tokenStr string) (bool, error) {
	if _, err := parseToken(tokenStr); err != nil {
		return false, err
	}
	return true, nil
}

func GetUsernameFromToken(tokenStr string) (string, error) {
	claims, err := parseToken(tokenStr)
	if err != nil {
		return "", errors.New("invalid token")
	}
	return claims.UserID, nil
}

// GetUserProfile returns user info for a given username.
func GetUserProfile(username string) (dtos.RegisterDTO, error) {
	return repositories.GetUserByID(username)
}

// IsDuressToken reports whether the caller authenticated with a duress PIN.
func IsDuressToken(tokenStr string) bool {
	claims, err := parseToken(tokenStr)
	if err != nil {
		return false
	}
	return claims.IsDuress
}

func ChangePin(username, currentPin, newPin string) error {
	pinType, err := repositories.ValidateUserCredentials(username, currentPin)
	if err != nil || pinType != 1 {
		return errors.New("incorrect current pin")
	}

	if _, err := repositories.ValidateUserCredentials(username, newPin); err == nil {
		return errors.New("new pin cannot be the same as your duress pin")
	}

	user, err := repositories.GetUserByID(username)
	if err != nil {
		return err
	}
	hash, err := HashPin(newPin)
	if err != nil {
		return err
	}
	user.NormalPinHash = hash
	user.NormalPin = ""
	return repositories.UpdateUser(user)
}

func ChangeDuressPin(username, currentPin, newPin string) error {
	pinType, err := repositories.ValidateUserCredentials(username, currentPin)
	if err != nil || pinType != 2 {
		return errors.New("incorrect current duress pin")
	}

	if pt, err := repositories.ValidateUserCredentials(username, newPin); err == nil && pt == 1 {
		return errors.New("new duress pin cannot be the same as your normal pin")
	}

	user, err := repositories.GetUserByID(username)
	if err != nil {
		return err
	}
	hash, err := HashPin(newPin)
	if err != nil {
		return err
	}
	user.DuressPinHash = hash
	user.DuressPin = ""
	return repositories.UpdateUser(user)
}
