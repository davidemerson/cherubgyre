package dtos

type LoginRequest struct {
	Username       string                 `json:"username"`
	PIN            string                 `json:"pin"`
	AdditionalData map[string]any `json:"additional_data"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
