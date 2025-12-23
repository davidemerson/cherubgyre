package repositories

import (
	"cherubgyre/dtos"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

func SaveUser(registerDTO dtos.RegisterDTO) error {
	file, err := os.OpenFile("users.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	if err := json.NewDecoder(file).Decode(&users); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding user data: %v", err)
		return err
	}

	// Check if username already exists
	for _, user := range users {
		if user.Username == registerDTO.Username {
			return errors.New("user already exists")
		}
	}

	// Check if invite_code is valid
	if registerDTO.InviteCode != "" {
		validInviteCode := false
		isMasterCode := false

		// Check hardcoded invite code first
		if registerDTO.InviteCode == "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2" {
			validInviteCode = true
			isMasterCode = true
		} else {
			// Check against existing users' invite codes
			for _, user := range users {
				if user.UserInviteCode == registerDTO.InviteCode {
					validInviteCode = true
					break
				}
			}
		}

		if !validInviteCode {
			return errors.New("invite code is not valid")
		}

		// Check if code has been used (skip for master code)
		if !isMasterCode {
			used, err := IsInviteCodeUsed(registerDTO.InviteCode)
			if err != nil {
				log.Printf("Error checking if invite code is used: %v", err)
				return err
			}
			if used {
				return errors.New("invite code has already been used")
			}
		}

		// Mark the code as used (skip for master code)
		if !isMasterCode {
			err := MarkInviteCodeAsUsed(registerDTO.InviteCode)
			if err != nil {
				log.Printf("Error marking invite code as used: %v", err)
				return err
			}
		}
	}

	users = append(users, registerDTO)

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(users); err != nil {
		log.Printf("Error encoding user data: %v", err)
		return err
	}

	log.Printf("User data written to file: %+v", registerDTO)
	return nil
}

func GetUserByID(username string) (dtos.RegisterDTO, error) {
	file, err := os.OpenFile("users.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return dtos.RegisterDTO{}, err
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	if err := json.NewDecoder(file).Decode(&users); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding user data: %v", err)
		return dtos.RegisterDTO{}, err
	}

	for _, user := range users {
		if user.Username == username {
			return user, nil
		}
	}

	return dtos.RegisterDTO{}, errors.New("user not found")
}

// ValidateUserCredentials checks if the provided PIN matches either the Normal or Duress PIN
// Returns: 0 = no match, 1 = Normal PIN match, 2 = Duress PIN match
func ValidateUserCredentials(username, pin string) (int, error) {
	user, err := GetUserByID(username)
	if err != nil {
		return 0, err
	}

	if user.NormalPin == pin {
		return 1, nil // Normal PIN match
	}

	if user.DuressPin == pin {
		return 2, nil // Duress PIN match
	}

	return 0, errors.New("invalid credentials")
}

func UpdateUser(updatedUser dtos.RegisterDTO) error {
	file, err := os.OpenFile("users.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	if err := json.NewDecoder(file).Decode(&users); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding user data: %v", err)
		return err
	}

	for i, user := range users {
		if user.Username == updatedUser.Username {
			users[i] = updatedUser
			break
		}
	}

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(users); err != nil {
		log.Printf("Error encoding user data: %v", err)
		return err
	}

	log.Printf("User data updated in file: %+v", updatedUser)
	return nil
}

func IsUsernameTaken(username string) (bool, error) {
	file, err := os.OpenFile("users.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file for username check: %v", err)
		return false, err // Return false and error if file cannot be opened
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	// Use json.NewDecoder.Decode. If the file is empty or not valid JSON, it might return EOF or other errors.
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	if err != nil {
		// If it's an EOF error and the file was just created (or empty), it means no users exist yet.
		// So, the username cannot be taken.
		fileInfo, statErr := file.Stat()
		if statErr == nil && fileInfo.Size() == 0 && errors.Is(err, io.EOF) {
			return false, nil // Empty file, username not taken
		}
		// For other decoding errors, or if it's EOF on a non-empty file (which is unusual for a list),
		// treat as an error, but log it. If it's just EOF on an empty array `[]`, that's fine.
		if errors.Is(err, io.EOF) && len(users) == 0 { // Handles empty JSON array `[]` or just empty file.
			return false, nil
		}
		// If there's any other error, or if it's EOF but users were partially decoded (which shouldn't happen with a list)
		log.Printf("Error decoding user data for username check: %v", err)
		return false, err // Return false and error for other decode issues
	}

	for _, user := range users {
		if user.Username == username {
			return true, nil // Username is taken
		}
	}

	return false, nil // Username is not taken
}

func ValidateInviteCode(inviteCode string) (bool, error) {
	// Check hardcoded invite code first (master code is always valid and unlimited)
	if inviteCode == "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2" {
		return true, nil
	}

	file, err := os.OpenFile("users.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file for invite code validation: %v", err)
		return false, err
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&users)
	if err != nil {
		fileInfo, statErr := file.Stat()
		if statErr == nil && fileInfo.Size() == 0 && errors.Is(err, io.EOF) {
			return false, nil // Empty file, no valid invite codes
		}
		if errors.Is(err, io.EOF) && len(users) == 0 {
			return false, nil
		}
		log.Printf("Error decoding user data for invite code validation: %v", err)
		return false, err
	}

	codeExists := false
	for _, user := range users {
		log.Printf("User invite code: %s", user.UserInviteCode)
		log.Printf("user name: %s", user.Username)
		if user.UserInviteCode == inviteCode {
			codeExists = true
			break
		}
	}

	if !codeExists {
		return false, nil
	}

	// Check if the code has been used
	used, err := IsInviteCodeUsed(inviteCode)
	if err != nil {
		log.Printf("Error checking if invite code is used: %v", err)
		return false, err
	}

	if used {
		return false, nil // Code exists but has been used
	}

	return true, nil
}

// IsInviteCodeUsed checks if an invite code has already been used
func IsInviteCodeUsed(inviteCode string) (bool, error) {
	file, err := os.OpenFile("used_invite_codes.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening used invite codes file: %v", err)
		return false, err
	}
	defer file.Close()

	var usedCodes []string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&usedCodes)
	if err != nil {
		fileInfo, statErr := file.Stat()
		if statErr == nil && fileInfo.Size() == 0 && errors.Is(err, io.EOF) {
			return false, nil // Empty file, no used codes
		}
		if errors.Is(err, io.EOF) && len(usedCodes) == 0 {
			return false, nil
		}
		log.Printf("Error decoding used invite codes: %v", err)
		return false, err
	}

	for _, code := range usedCodes {
		if code == inviteCode {
			return true, nil
		}
	}

	return false, nil
}

// MarkInviteCodeAsUsed adds an invite code to the used list
func MarkInviteCodeAsUsed(inviteCode string) error {
	file, err := os.OpenFile("used_invite_codes.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening used invite codes file: %v", err)
		return err
	}
	defer file.Close()

	var usedCodes []string
	if err := json.NewDecoder(file).Decode(&usedCodes); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding used invite codes: %v", err)
		return err
	}

	// Add the code to the list
	usedCodes = append(usedCodes, inviteCode)

	// Write back to file
	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(usedCodes); err != nil {
		log.Printf("Error encoding used invite codes: %v", err)
		return err
	}

	log.Printf("Invite code marked as used: %s", inviteCode)
	return nil
}

