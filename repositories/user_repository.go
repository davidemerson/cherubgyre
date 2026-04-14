package repositories

import (
	"cherubgyre/dtos"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// MasterInviteCode is the unlimited-use bootstrap invite. Kept public so
// services and tests can reference it without stringly-typed duplication.
// Should be overridden via a config option in a future pass so it isn't
// compiled into the binary.
const MasterInviteCode = "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2"

var (
	userStore       = newFileStore("users.json")
	usedInviteStore = newFileStore("used_invite_codes.json")
)

func loadUsers() ([]dtos.RegisterDTO, error) {
	var users []dtos.RegisterDTO
	if err := userStore.load(&users); err != nil {
		return nil, err
	}
	return users, nil
}

func loadUsedInviteCodes() ([]string, error) {
	var used []string
	if err := usedInviteStore.load(&used); err != nil {
		return nil, err
	}
	return used, nil
}

// SaveUser persists a new user record, performing the following under a
// single write lock so registration is race-free:
//   - rejects duplicate usernames,
//   - validates the invite code (master code always OK; otherwise must
//     match some existing user's UserInviteCode and be unused),
//   - marks the code as used when appropriate,
//   - appends the new record.
func SaveUser(registerDTO dtos.RegisterDTO) error {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()

	var users []dtos.RegisterDTO
	if err := userStore.loadLocked(&users); err != nil {
		return err
	}

	for _, u := range users {
		if u.Username == registerDTO.Username {
			return errors.New("user already exists")
		}
	}

	if registerDTO.InviteCode != "" {
		isMasterCode := registerDTO.InviteCode == MasterInviteCode
		validCode := isMasterCode
		if !validCode {
			for _, u := range users {
				if u.UserInviteCode == registerDTO.InviteCode {
					validCode = true
					break
				}
			}
		}
		if !validCode {
			return errors.New("invite code is not valid")
		}

		if !isMasterCode {
			used, err := IsInviteCodeUsed(registerDTO.InviteCode)
			if err != nil {
				return err
			}
			if used {
				return errors.New("invite code has already been used")
			}
			if err := MarkInviteCodeAsUsed(registerDTO.InviteCode); err != nil {
				return err
			}
		}
	}

	users = append(users, registerDTO)
	if err := userStore.saveLocked(users); err != nil {
		return err
	}
	log.Printf("User saved: %s", registerDTO.Username)
	return nil
}

func GetUserByID(username string) (dtos.RegisterDTO, error) {
	users, err := loadUsers()
	if err != nil {
		return dtos.RegisterDTO{}, err
	}
	for _, u := range users {
		if u.Username == username {
			return u, nil
		}
	}
	return dtos.RegisterDTO{}, errors.New("user not found")
}

// ValidateUserCredentials checks the provided PIN against the stored Normal
// and Duress PIN hashes (bcrypt, constant-time). If the stored record is a
// legacy plaintext entry from before the hash migration, it falls back to
// a direct comparison so the service stays available during rollout — the
// startup migration in services.MigratePinHashes rewrites those records on
// next boot.
//
// Returns: 0 = no match, 1 = Normal PIN match, 2 = Duress PIN match.
func ValidateUserCredentials(username, pin string) (int, error) {
	user, err := GetUserByID(username)
	if err != nil {
		return 0, err
	}

	if user.NormalPinHash != "" {
		if bcrypt.CompareHashAndPassword([]byte(user.NormalPinHash), []byte(pin)) == nil {
			return 1, nil
		}
	} else if user.NormalPin != "" && user.NormalPin == pin {
		return 1, nil
	}

	if user.DuressPinHash != "" {
		if bcrypt.CompareHashAndPassword([]byte(user.DuressPinHash), []byte(pin)) == nil {
			return 2, nil
		}
	} else if user.DuressPin != "" && user.DuressPin == pin {
		return 2, nil
	}

	return 0, errors.New("invalid credentials")
}

func UpdateUser(updatedUser dtos.RegisterDTO) error {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()

	var users []dtos.RegisterDTO
	if err := userStore.loadLocked(&users); err != nil {
		return err
	}

	found := false
	for i, u := range users {
		if u.Username == updatedUser.Username {
			users[i] = updatedUser
			found = true
			break
		}
	}
	if !found {
		return errors.New("user not found")
	}

	if err := userStore.saveLocked(users); err != nil {
		return err
	}
	log.Printf("User updated: %s", updatedUser.Username)
	return nil
}

func IsUsernameTaken(username string) (bool, error) {
	users, err := loadUsers()
	if err != nil {
		return false, err
	}
	for _, u := range users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

// ValidateInviteCode returns true when the supplied code is either the
// master bootstrap code or is present in some user's UserInviteCode field
// AND has not yet been consumed.
func ValidateInviteCode(inviteCode string) (bool, error) {
	if inviteCode == MasterInviteCode {
		return true, nil
	}

	users, err := loadUsers()
	if err != nil {
		return false, err
	}

	exists := false
	for _, u := range users {
		if u.UserInviteCode == inviteCode {
			exists = true
			break
		}
	}
	if !exists {
		return false, nil
	}

	used, err := IsInviteCodeUsed(inviteCode)
	if err != nil {
		return false, err
	}
	return !used, nil
}

func IsInviteCodeUsed(inviteCode string) (bool, error) {
	used, err := loadUsedInviteCodes()
	if err != nil {
		return false, err
	}
	for _, code := range used {
		if code == inviteCode {
			return true, nil
		}
	}
	return false, nil
}

func MarkInviteCodeAsUsed(inviteCode string) error {
	usedInviteStore.mu.Lock()
	defer usedInviteStore.mu.Unlock()

	var used []string
	if err := usedInviteStore.loadLocked(&used); err != nil {
		return err
	}
	used = append(used, inviteCode)
	if err := usedInviteStore.saveLocked(used); err != nil {
		return err
	}
	log.Printf("Invite code marked as used: %s", inviteCode)
	return nil
}

func DeleteUser(username string) error {
	userStore.mu.Lock()
	defer userStore.mu.Unlock()

	var users []dtos.RegisterDTO
	if err := userStore.loadLocked(&users); err != nil {
		return err
	}

	index := -1
	for i, u := range users {
		if u.Username == username {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New("user not found")
	}

	users = append(users[:index], users[index+1:]...)
	if err := userStore.saveLocked(users); err != nil {
		return err
	}
	log.Printf("User deleted: %s", username)
	return nil
}

func GetAllUsers() ([]dtos.RegisterDTO, error) {
	return loadUsers()
}
