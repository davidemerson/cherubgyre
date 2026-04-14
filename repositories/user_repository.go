// Package repositories persists cherubgyre state to JSON files.
//
// # Lock ordering
//
// Two fileStore mutexes in this package can be held simultaneously:
// userStore and usedInviteStore. SaveUser takes userStore first and
// then calls IsInviteCodeUsed / MarkInviteCodeAsUsed, which acquire
// usedInviteStore. To prevent deadlock, **userStore must always be
// acquired before usedInviteStore**. No function in this package
// currently acquires them in the opposite order; adding one without
// updating this rule will introduce a latent hang.
package repositories

import (
	"cherubgyre/dtos"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// MasterInviteCode is the unlimited-use bootstrap invite. Defaults to a
// well-known value for backwards compatibility with existing test
// fixtures, but can be overridden at startup via SetMasterInviteCode
// (wired from config.MasterInviteCode in main.go).
//
// Deployments that have bootstrapped past the first user should set
// MASTER_INVITE_CODE to an empty string (or any value they never intend
// to honor). When it is empty, the "is this the master code?" check
// short-circuits to false for every input, effectively disabling the
// master-code path entirely.
var MasterInviteCode = "4f88690e-0fbc-47b9-88e3-2d5ee2ac03d2"

// SetMasterInviteCode installs the master-code value from config. Calling
// with "" disables the master-code path. Called from main() at startup.
func SetMasterInviteCode(code string) {
	MasterInviteCode = code
}

// isMaster returns true when the supplied code is non-empty and matches
// the configured master code. An empty MasterInviteCode disables
// matching entirely.
func isMaster(code string) bool {
	return MasterInviteCode != "" && code == MasterInviteCode
}

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
		isMasterCode := isMaster(registerDTO.InviteCode)
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
	slog.Info("user saved", slog.String("user", registerDTO.Username))
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

// ValidateUserCredentials checks the provided PIN against the stored
// Normal and Duress PIN hashes (bcrypt, constant-time).
//
// Returns: 0 = no match, 1 = Normal PIN match, 2 = Duress PIN match.
func ValidateUserCredentials(username, pin string) (int, error) {
	// Bound the input length so an oversized pin cannot pass through to
	// bcrypt.CompareHashAndPassword — cosmetic now that the HTTP body
	// cap is 8 KiB on auth routes, but keeps the repository honest if
	// someone adds a non-HTTP entry point later.
	if len(pin) > 128 {
		return 0, errors.New("invalid credentials")
	}
	user, err := GetUserByID(username)
	if err != nil {
		return 0, err
	}

	if user.NormalPinHash != "" &&
		bcrypt.CompareHashAndPassword([]byte(user.NormalPinHash), []byte(pin)) == nil {
		return 1, nil
	}
	if user.DuressPinHash != "" &&
		bcrypt.CompareHashAndPassword([]byte(user.DuressPinHash), []byte(pin)) == nil {
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
	slog.Debug("user updated", slog.String("user", updatedUser.Username))
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
	if isMaster(inviteCode) {
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
	slog.Info("invite code marked used", slog.String("code", inviteCode))
	return nil
}

// DeleteUser removes the user record from users.json AND purges every
// related entry from followers.json so that a recycled username cannot
// inherit the old user's follower graph. The follower cleanup happens
// first: if it fails, we haven't yet lost the user record, so a retry
// can still succeed.
func DeleteUser(username string) error {
	if err := DeleteUserRelations(username); err != nil {
		return fmt.Errorf("purge follower relations: %w", err)
	}

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
	slog.Info("user deleted", slog.String("user", username))
	return nil
}

func GetAllUsers() ([]dtos.RegisterDTO, error) {
	return loadUsers()
}
