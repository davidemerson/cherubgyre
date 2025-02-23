package dtos

import (
	"net/http"
)

type ResponseDTO struct {
	Message string `json:"message,omitempty"`
}

func (dto *ResponseDTO) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
