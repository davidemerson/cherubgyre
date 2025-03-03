package dtos

type ClerkEventVerification struct {
	Attempts interface{} `json:"attempts"`
	ExpireAt interface{} `json:"expire_at"`
	Status   string      `json:"status"`
	Strategy string      `json:"strategy"`
}

type ClerkEventEmailAddress struct {
	CreatedAt    int64                  `json:"created_at"`
	EmailAddress string                 `json:"email_address"`
	ID           string                 `json:"id"`
	LinkedTo     []ClerkEventLinkedTo   `json:"linked_to"`
	Object       string                 `json:"object"`
	Reserved     bool                   `json:"reserved"`
	UpdatedAt    int64                  `json:"updated_at"`
	Verification ClerkEventVerification `json:"verification"`
}

type ClerkEventLinkedTo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ClerkEventExternalAccount struct {
	ApprovedScopes string                 `json:"approved_scopes"`
	CreatedAt      int64                  `json:"created_at"`
	EmailAddress   string                 `json:"email_address"`
	FamilyName     string                 `json:"family_name"`
	GivenName      string                 `json:"given_name"`
	GoogleID       string                 `json:"google_id"`
	ID             string                 `json:"id"`
	Label          interface{}            `json:"label"`
	Object         string                 `json:"object"`
	Picture        string                 `json:"picture"`
	PublicMetadata interface{}            `json:"public_metadata"`
	UpdatedAt      int64                  `json:"updated_at"`
	Username       interface{}            `json:"username"`
	Verification   ClerkEventVerification `json:"verification"`
}

type ClerkEventData struct {
	BackupCodeEnabled             bool                        `json:"backup_code_enabled"`
	Banned                        bool                        `json:"banned"`
	CreateOrganizationEnabled     bool                        `json:"create_organization_enabled"`
	CreatedAt                     int64                       `json:"created_at"`
	DeleteSelfEnabled             bool                        `json:"delete_self_enabled"`
	EmailAddresses                []ClerkEventEmailAddress    `json:"email_addresses"`
	ExternalAccounts              []ClerkEventExternalAccount `json:"external_accounts"`
	ExternalID                    interface{}                 `json:"external_id"`
	FirstName                     string                      `json:"first_name"`
	HasImage                      bool                        `json:"has_image"`
	ID                            string                      `json:"id"`
	ImageURL                      string                      `json:"image_url"`
	LastActiveAt                  int64                       `json:"last_active_at"`
	LastName                      string                      `json:"last_name"`
	LastSignInAt                  interface{}                 `json:"last_sign_in_at"`
	Locked                        bool                        `json:"locked"`
	LockoutExpiresInSeconds       interface{}                 `json:"lockout_expires_in_seconds"`
	Object                        string                      `json:"object"`
	Passkeys                      []interface{}               `json:"passkeys"`
	PasswordEnabled               bool                        `json:"password_enabled"`
	PhoneNumbers                  []interface{}               `json:"phone_numbers"`
	PrimaryEmailAddressID         string                      `json:"primary_email_address_id"`
	PrimaryPhoneNumberID          interface{}                 `json:"primary_phone_number_id"`
	PrimaryWeb3WalletID           interface{}                 `json:"primary_web3_wallet_id"`
	PrivateMetadata               interface{}                 `json:"private_metadata"`
	ProfileImageURL               string                      `json:"profile_image_url"`
	PublicMetadata                interface{}                 `json:"public_metadata"`
	SamlAccounts                  []interface{}               `json:"saml_accounts"`
	TotpEnabled                   bool                        `json:"totp_enabled"`
	TwoFactorEnabled              bool                        `json:"two_factor_enabled"`
	UnsafeMetadata                interface{}                 `json:"unsafe_metadata"`
	UpdatedAt                     int64                       `json:"updated_at"`
	Username                      string                      `json:"username"`
	VerificationAttemptsRemaining int                         `json:"verification_attempts_remaining"`
	Web3Wallets                   []interface{}               `json:"web3_wallets"`
}

type ClerkEventDto struct {
	Data   ClerkEventData `json:"data"`
	Object string         `json:"object"`
	Type   string         `json:"type"`
}
