package dtos

type RegisterDTO struct {
	UserInviteCode string `json:"invite_code_user,omitempty"`
	NormalPin      string `json:"normal_pin"`
	DuressPin      string `json:"duress_pin"`
	Username       string `json:"username,omitempty"`
	InviteCode     string `json:"invite_code"`
}
