package dtos

import (
	"time"
)

// RegisterDTO is both the wire shape for POST /register and the on-disk
// storage record for a user.
//
// NormalPin and DuressPin are incoming-only fields that carry the
// plaintext PIN from a register or change-pin request to the service
// layer. They are never persisted: services.RegisterUser hashes them
// into NormalPinHash / DuressPinHash and clears the plaintext before
// calling repositories.SaveUser, and controllers.Register zeroes them
// again in the response body as defense in depth.
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
