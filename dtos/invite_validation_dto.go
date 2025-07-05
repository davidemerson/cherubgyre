package dtos

type InviteValidationRequest struct {
	InviteCode string `json:"invite_code"`
}

type InviteValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}
