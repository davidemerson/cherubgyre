package services

import (
	"cherubgyre/repositories"
	"log/slog"
	"time"
)

// PostDuress creates a user-initiated duress signal. The caller must have
// re-entered their duress PIN; we re-validate here as defense in depth.
// Rate-limited to one signal per hour per user (per spec).
func PostDuress(username, duressType, message string, timestamp time.Time, additionalData map[string]any, duressPin string) error {
	user, err := repositories.GetUserByID(username)
	if err != nil {
		return ErrInvalidCredentials
	}

	pinType, err := repositories.ValidateUserCredentials(username, duressPin)
	if err != nil || pinType != 2 {
		return ErrInvalidCredentials
	}

	if !user.LastDuressAt.IsZero() && time.Since(user.LastDuressAt) < time.Hour {
		return ErrDuressRateLimited
	}

	if err := repositories.SaveDuress(username, duressType, message, timestamp, additionalData); err != nil {
		slog.Error("save duress failed", slog.String("user", username), slog.Any("err", err))
		return err
	}

	user.LastDuressAt = time.Now()
	if err := repositories.UpdateUser(user); err != nil {
		slog.Error("failed to persist last_duress_at",
			slog.String("user", username), slog.Any("err", err))
	}
	return nil
}

// CancelDuress clears the caller's active duress signal. Per spec, this
// requires re-entering the normal PIN so a coercer holding the duress-mode
// session cannot cancel the alert they triggered.
func CancelDuress(username, normalPin string) error {
	pinType, err := repositories.ValidateUserCredentials(username, normalPin)
	if err != nil || pinType != 1 {
		return ErrInvalidCredentials
	}
	return repositories.DeleteDuress(username)
}

func GetDuressMap(username string) (map[string]any, error) {
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
		return ErrInvalidCredentials
	}
	if pinType != 2 {
		return ErrInvalidDuressPin
	}
	return nil
}
