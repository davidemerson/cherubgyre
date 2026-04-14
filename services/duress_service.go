package services

import (
	"cherubgyre/repositories"
	"errors"
	"log"
	"time"
)

// PostDuress creates a user-initiated duress signal. The caller must have
// re-entered their duress PIN; we re-validate here as defense in depth.
// Rate-limited to one signal per hour per user (per spec).
func PostDuress(username, duressType, message string, timestamp time.Time, additionalData map[string]interface{}, duressPin string) error {
	user, err := repositories.GetUserByID(username)
	if err != nil {
		return err
	}

	pinType, err := repositories.ValidateUserCredentials(username, duressPin)
	if err != nil || pinType != 2 {
		return errors.New("invalid credentials")
	}

	if !user.LastDuressAt.IsZero() && time.Since(user.LastDuressAt) < time.Hour {
		return errors.New("duress rate limit exceeded")
	}

	if err := repositories.SaveDuress(username, duressType, message, timestamp, additionalData); err != nil {
		log.Println("Error saving duress:", err)
		return err
	}

	user.LastDuressAt = time.Now()
	if err := repositories.UpdateUser(user); err != nil {
		log.Printf("Failed to persist LastDuressAt for %s: %v", username, err)
	}
	return nil
}

// CancelDuress clears the caller's active duress signal. Per spec, this
// requires re-entering the normal PIN so a coercer holding the duress-mode
// session cannot cancel the alert they triggered.
func CancelDuress(username, normalPin string) error {
	pinType, err := repositories.ValidateUserCredentials(username, normalPin)
	if err != nil || pinType != 1 {
		return errors.New("invalid credentials")
	}
	return repositories.DeleteDuress(username)
}

func GetDuressMap(username string) (map[string]interface{}, error) {
	return repositories.GetDuressMap(username)
}

func GetFollowingDuress(username string) ([]repositories.Duress, error) {
	following, err := repositories.GetFollowing(username)
	if err != nil {
		return nil, err
	}
	return repositories.GetActiveDuressForUsers(following)
}

// VerifyDuressPin confirms that `pin` is the caller's duress PIN. Used by
// the /duress/verify endpoint before granting special access to
// duress-related views in the app UI.
func VerifyDuressPin(username, pin string) error {
	pinType, err := repositories.ValidateUserCredentials(username, pin)
	if err != nil {
		return errors.New("invalid credentials")
	}
	if pinType != 2 {
		return errors.New("invalid duress pin")
	}
	return nil
}
