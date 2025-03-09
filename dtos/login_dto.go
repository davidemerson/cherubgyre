package dtos

type LoginRequest struct {
	Username string `json:"username"`
	PIN      string `json:"pin"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
