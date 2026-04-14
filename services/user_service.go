package services

import (
	"cherubgyre/repositories"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// BackfillUIDs assigns a UUIDv7 to any legacy user record that predates
// the UID field. Idempotent: users with an existing UID are untouched.
// Called once at startup from main().
func BackfillUIDs() error {
	users, err := repositories.GetAllUsers()
	if err != nil {
		return err
	}
	filled := 0
	for _, u := range users {
		if u.UID != "" {
			continue
		}
		uid, err := uuid.NewV7()
		if err != nil {
			slog.Error("failed to mint uid", slog.String("user", u.Username), slog.Any("err", err))
			continue
		}
		u.UID = uid.String()
		if err := repositories.UpdateUser(u); err != nil {
			slog.Error("failed to persist uid", slog.String("user", u.Username), slog.Any("err", err))
			continue
		}
		filled++
	}
	if filled > 0 {
		slog.Info("uid backfill complete", slog.Int("assigned", filled))
	}
	return nil
}

// DeregisterUser handles the complete removal of a user and notifies friends
func DeregisterUser(username string, reason string) error {
	slog.Info("deregistering user",
		slog.String("user", username), slog.String("reason", reason))

	// 1. Create a "Left" duress signal (mimicking duress). We post this
	// BEFORE the user is deleted so followers can see the final signal.
	// If it fails we still proceed with deletion — the priority is
	// removing the data.
	err := repositories.SaveDuress(
		username,
		"User Left",
		fmt.Sprintf("User %s has left the app (%s)", username, reason),
		time.Now(),
		map[string]any{"reason": reason, "type": "deregistration"},
	)
	if err != nil {
		slog.Error("failed to create exit signal",
			slog.String("user", username), slog.Any("err", err))
	}

	// 2. Delete the user data
	if err := repositories.DeleteUser(username); err != nil {
		slog.Error("failed to delete user",
			slog.String("user", username), slog.Any("err", err))
		return err
	}

	return nil
}

// CheckInactivity iterates through all users and deregisters those
// inactive for > 1 year. Honors ctx cancellation between users so a
// graceful-shutdown signal can interrupt a long sweep cleanly.
func CheckInactivity(ctx context.Context) error {
	slog.Info("inactivity check starting")
	users, err := repositories.GetAllUsers()
	if err != nil {
		return err
	}

	expirationDuration := 365 * 24 * time.Hour

	for _, user := range users {
		select {
		case <-ctx.Done():
			slog.Info("inactivity check interrupted by shutdown")
			return ctx.Err()
		default:
		}

		// Legacy users with a zero LastActive are skipped until they
		// log in once and get a real timestamp.
		if user.LastActive.IsZero() {
			continue
		}

		if time.Since(user.LastActive) > expirationDuration {
			slog.Info("inactive user deregistration",
				slog.String("user", user.Username),
				slog.Time("last_active", user.LastActive))
			if err := DeregisterUser(user.Username, "Inactivity"); err != nil {
				slog.Error("failed to deregister inactive user",
					slog.String("user", user.Username), slog.Any("err", err))
			}
		}
	}

	slog.Info("inactivity check complete")
	return nil
}
