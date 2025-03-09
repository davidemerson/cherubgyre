package dtos

type RegisterDTO struct {
	InviteCode string `json:"invite_code,omitempty"`
	NormalPin  string `json:"normal_pin"`
	DuressPin  string `json:"duress_pin"`
	Username   string `json:"username"`
}
