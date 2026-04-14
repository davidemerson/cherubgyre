package services

import (
	"cherubgyre/repositories"
	"context"
	"fmt"
	"log"
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
			log.Printf("failed to mint uid for %s: %v", u.Username, err)
			continue
		}
		u.UID = uid.String()
		if err := repositories.UpdateUser(u); err != nil {
			log.Printf("failed to persist uid for %s: %v", u.Username, err)
			continue
		}
		filled++
	}
	if filled > 0 {
		log.Printf("UID backfill: assigned %d UID(s)", filled)
	}
	return nil
}

// DeregisterUser handles the complete removal of a user and notifies friends
func DeregisterUser(username string, reason string) error {
	log.Printf("Deregistering user %s due to: %s", username, reason)

	// 1. Create a "Left" duress signal (mimicking duress)
	// We need to manually construct this since we might not have a token
	// This relies on the repository layer directly
	err := repositories.SaveDuress(
		username,
		"User Left",
		fmt.Sprintf("User %s has left the app (%s)", username, reason),
		time.Now(),
		map[string]interface{}{"reason": reason, "type": "deregistration"},
	)
	if err != nil {
		log.Printf("Error creating exit signal for %s: %v", username, err)
		// Continue with deletion even if signal fails? 
		// Ideally yes, we want them gone.
	}

	// 2. Delete the user data
	if err := repositories.DeleteUser(username); err != nil {
		log.Printf("Error deleting user %s: %v", username, err)
		return err
	}

	return nil
}

// MigratePinHashes rewrites any user record that still has plaintext
// NormalPin/DuressPin fields, hashing them with bcrypt and clearing the
// plaintext. Idempotent: records that already have hashes are skipped.
// Called once at startup so that legacy users.json files from before the
// bcrypt rollout are brought into the new format with zero operator work.
func MigratePinHashes() error {
	users, err := repositories.GetAllUsers()
	if err != nil {
		return err
	}
	migrated := 0
	for _, user := range users {
		changed := false
		if user.NormalPinHash == "" && user.NormalPin != "" {
			hash, err := HashPin(user.NormalPin)
			if err != nil {
				log.Printf("Failed to hash normal pin for %s: %v", user.Username, err)
				continue
			}
			user.NormalPinHash = hash
			user.NormalPin = ""
			changed = true
		}
		if user.DuressPinHash == "" && user.DuressPin != "" {
			hash, err := HashPin(user.DuressPin)
			if err != nil {
				log.Printf("Failed to hash duress pin for %s: %v", user.Username, err)
				continue
			}
			user.DuressPinHash = hash
			user.DuressPin = ""
			changed = true
		}
		if changed {
			if err := repositories.UpdateUser(user); err != nil {
				log.Printf("Failed to persist migrated pins for %s: %v", user.Username, err)
				continue
			}
			migrated++
		}
	}
	if migrated > 0 {
		log.Printf("PIN migration: hashed %d legacy user record(s)", migrated)
	}
	return nil
}

// CheckInactivity iterates through all users and deregisters those
// inactive for > 1 year. Honors ctx cancellation between users so a
// graceful-shutdown signal can interrupt a long sweep cleanly.
func CheckInactivity(ctx context.Context) error {
	log.Println("Starting inactivity check...")
	users, err := repositories.GetAllUsers()
	if err != nil {
		return err
	}

	expirationDuration := 365 * 24 * time.Hour

	for _, user := range users {
		select {
		case <-ctx.Done():
			log.Println("Inactivity check interrupted by shutdown")
			return ctx.Err()
		default:
		}

		// Legacy users with a zero LastActive are skipped until they
		// log in once and get a real timestamp.
		if user.LastActive.IsZero() {
			continue
		}

		if time.Since(user.LastActive) > expirationDuration {
			log.Printf("User %s is inactive (Last active: %v). Deregistering...", user.Username, user.LastActive)
			if err := DeregisterUser(user.Username, "Inactivity"); err != nil {
				log.Printf("Failed to deregister inactive user %s: %v", user.Username, err)
			}
		}
	}

	log.Println("Inactivity check complete.")
	return nil
}
