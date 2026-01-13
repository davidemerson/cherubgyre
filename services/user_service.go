package services

import (
	"cherubgyre/repositories"
	"fmt"
	"log"
	"time"
)

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

// CheckInactivity iterates through all users and deregisers those inactive for > 1 year
func CheckInactivity() error {
	log.Println("Starting inactivity check...")
	users, err := repositories.GetAllUsers()
	if err != nil {
		return err
	}

	expirationDuration := 365 * 24 * time.Hour 
	// expirationDuration := 2 * time.Minute // Debug mode

	for _, user := range users {
		// If LastActive is zero (legacy users), maybe default to now? 
		// Or if we want to be strict, we can't delete them yet. 
		// Let's assume zero time means 'active now' for legacy migration purposes, 
		// OR we ignore them until they login once.
		if user.LastActive.IsZero() {
			continue 
		}

		if time.Since(user.LastActive) > expirationDuration {
			log.Printf("User %s is inactive (Last active: %v). Deregistering...", user.Username, user.LastActive)
			DeregisterUser(user.Username, "Inactivity")
		}
	}
	
	log.Println("Inactivity check complete.")
	return nil
}
