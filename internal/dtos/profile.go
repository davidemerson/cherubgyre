package dtos

import (
	"net/http"
)

type ProfileResponseDTO struct {
	Fullname     string `json:"fullname,omitempty"`
	ProfileImage string `json:"profile_image,omitempty"`
	Email        string `json:"email,omitempty"`
}

func (dto *ProfileResponseDTO) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
