package dtos

// ContextKey is a custom type to avoid key collisions in context
type ContextKey string

const (
	TokenClaimsKey ContextKey = "TokenClaims"
)

/*
To have access to the user claims in the context, we need to use a custom JWT template in clerk with the following claims:

	{
		"id": "{{user.id}}",
		"username": "{{user.username}}",
		"created_at": "{{user.created_at}}",
		"updated_at": "{{user.updated_at}}",
		"email_verified": "{{user.email_verified}}"
	}

Make sure on the frontend, you are generating the JWT token with the above template.
*/
type UserClaims struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	CreatedAt     uint32 `json:"created_at"`
	UpdatedAt     uint32 `json:"updated_at"`
	EmailVerified bool   `json:"email_verified"`
}
