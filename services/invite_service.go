package services

import (
	"cherubgyre/repositories"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// ErrRateLimitExceeded is returned by CheckRateLimit and surfaces up
// through CreateInvite, LoginLimiter, and anywhere else the generic
// sliding-window helper is used.
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// CreateInvite mints a new invite UUID on behalf of the already-authenticated
// user. Rate-limited to 5 per 168 hours per the spec.
func CreateInvite(username string) (string, error) {
	user, err := repositories.GetUserByID(username)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	newHistory, err := CheckRateLimit(user.InviteGenerationHistory, now, 5, 168*3600)
	if err != nil {
		return "", err
	}
	user.InviteGenerationHistory = newHistory

	inviteCode := uuid.New().String()
	user.UserInviteCode = inviteCode

	if err := repositories.UpdateUser(user); err != nil {
		return "", err
	}

	slog.Info("invite code created", slog.String("user", username))
	return inviteCode, nil
}

// CheckRateLimit implements a generic sliding-window limiter over a sorted
// list of epoch-second timestamps. Returns a pruned history with `now`
// appended on success, or an error if the window would exceed `limit`.
func CheckRateLimit(history []int64, now int64, limit int, windowSeconds int64) ([]int64, error) {
	validHistory := make([]int64, 0, len(history))
	for _, ts := range history {
		if now-ts < windowSeconds {
			validHistory = append(validHistory, ts)
		}
	}

	if len(validHistory) >= limit {
		return nil, ErrRateLimitExceeded
	}

	validHistory = append(validHistory, now)
	return validHistory, nil
}
