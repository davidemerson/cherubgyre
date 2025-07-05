package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"errors"
	"log"
)

func ValidateInviteCode(inviteCode string) (dtos.InviteValidationResponse, error) {
	if inviteCode == "" {
		return dtos.InviteValidationResponse{
			Valid:   false,
			Message: "Invite code is required",
		}, errors.New("invite code is required")
	}

	valid, err := repositories.ValidateInviteCode(inviteCode)
	if err != nil {
		log.Printf("Error validating invite code: %v", err)
		return dtos.InviteValidationResponse{
			Valid:   false,
			Message: "Error validating invite code",
		}, err
	}

	if valid {
		return dtos.InviteValidationResponse{
			Valid:   true,
			Message: "Invite code is valid",
		}, nil
	}

	return dtos.InviteValidationResponse{
		Valid:   false,
		Message: "Invalid invite code",
	}, nil
}
