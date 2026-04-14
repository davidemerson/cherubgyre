package dtos

import (
	"time"
)

// RegisterDTO is both the wire shape for POST /register and the on-disk
// storage record for a user. NormalPin/DuressPin are only populated on
// incoming requests or in legacy records prior to the hash migration; at
// rest they are always empty and the hash fields are authoritative.
type RegisterDTO struct {
	UID                     string    `json:"uid,omitempty"`
	UserInviteCode          string    `json:"invite_code_user,omitempty"`
	NormalPin               string    `json:"normal_pin,omitempty"`
	DuressPin               string    `json:"duress_pin,omitempty"`
	NormalPinHash           string    `json:"normal_pin_hash,omitempty"`
	DuressPinHash           string    `json:"duress_pin_hash,omitempty"`
	Username                string    `json:"username,omitempty"`
	InviteCode              string    `json:"invite_code,omitempty"`
	Avatar                  string    `json:"avatar,omitempty"`
	InviteGenerationHistory []int64   `json:"invite_generation_history,omitempty"`
	FailedAttempts          int       `json:"failed_attempts"`
	LastActive              time.Time `json:"last_active"`
	LastDuressAt            time.Time `json:"last_duress_at,omitempty"`
}
