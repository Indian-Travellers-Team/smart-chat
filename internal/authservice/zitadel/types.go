package zitadel

// ValidateTokenRequest accepts an optional token from request body.
// When empty, Authorization bearer token is used as fallback.
type ValidateTokenRequest struct {
	Token *string `json:"token"`
}

type ValidateTokenUser struct {
	ID    *string `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Role  *string `json:"role"`
}

type ValidateTokenResponse struct {
	User  *ValidateTokenUser `json:"user"`
	Error *string            `json:"error"`
}

type ZitadelConfig struct {
	AuthServiceBaseURL string
	ValidateTokenPath  string
}
