package repositories

import (
	"log/slog"
	"time"
)

type Duress struct {
	Username       string                 `json:"username"`
	DuressType     string                 `json:"duress_type"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	AdditionalData map[string]any `json:"additional_data"`
}

var duressStore = newFileStore("duress.json")

func loadDuresses() ([]Duress, error) {
	var duresses []Duress
	if err := duressStore.load(&duresses); err != nil {
		return nil, err
	}
	return duresses, nil
}

// SaveDuress enforces the per-user invariant that there is at most one
// active duress signal. Any pre-existing signal for the same user is
// dropped before the new one is appended.
func SaveDuress(username, duressType, message string, timestamp time.Time, additionalData map[string]any) error {
	slog.Debug("saving duress signal", slog.String("user", username), slog.String("type", duressType))
	duressStore.mu.Lock()
	defer duressStore.mu.Unlock()

	var duresses []Duress
	if err := duressStore.loadLocked(&duresses); err != nil {
		return err
	}

	filtered := make([]Duress, 0, len(duresses))
	for _, d := range duresses {
		if d.Username != username {
			filtered = append(filtered, d)
		}
	}
	filtered = append(filtered, Duress{
		Username:       username,
		DuressType:     duressType,
		Message:        message,
		Timestamp:      timestamp,
		AdditionalData: additionalData,
	})

	return duressStore.saveLocked(filtered)
}

func DeleteDuress(username string) error {
	slog.Debug("deleting duress signal", slog.String("user", username))
	duressStore.mu.Lock()
	defer duressStore.mu.Unlock()

	var duresses []Duress
	if err := duressStore.loadLocked(&duresses); err != nil {
		return err
	}

	kept := make([]Duress, 0, len(duresses))
	for _, d := range duresses {
		if d.Username != username {
			kept = append(kept, d)
		}
	}

	return duressStore.saveLocked(kept)
}

// GetMyDuress returns the caller's own active duress signal (if any).
// Replaces the legacy GetDuressMap which returned a 1-entry map.
func GetMyDuress(username string) (*Duress, error) {
	duresses, err := loadDuresses()
	if err != nil {
		return nil, err
	}
	for i := range duresses {
		if duresses[i].Username == username {
			d := duresses[i]
			return &d, nil
		}
	}
	return nil, nil
}

// GetDuressMap remains for backwards compatibility with the existing
// /users/map route. Wraps GetMyDuress in the legacy shape.
func GetDuressMap(username string) (map[string]any, error) {
	d, err := GetMyDuress(username)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return map[string]any{}, nil
	}
	return map[string]any{d.Username: *d}, nil
}

func GetActiveDuressForUsers(usernames []string) ([]Duress, error) {
	duresses, err := loadDuresses()
	if err != nil {
		return nil, err
	}

	want := make(map[string]struct{}, len(usernames))
	for _, u := range usernames {
		want[u] = struct{}{}
	}

	var active []Duress
	for _, d := range duresses {
		if _, ok := want[d.Username]; ok {
			active = append(active, d)
		}
	}
	return active, nil
}
