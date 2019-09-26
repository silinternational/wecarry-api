package auth

import (
	"github.com/gobuffalo/buffalo"
)

const TypeSaml = "saml"

// Environment variable keys
const AuthCallbackURLEnv = "AUTH_CALLBACK_URL"

// Provider interface to be implemented by any auth providers
type Provider interface {
	Logout(c buffalo.Context) Response

	AuthRequest(c buffalo.Context) (string, error)

	AuthCallback(c buffalo.Context) Response
}

// User holds common attributes expected from auth providers
type User struct {
	FirstName string
	LastName  string
	Email     string
	UserID    string
	Nickname  string
}

// Response holds fields for login and logout responses. not all fields will have values
type Response struct {
	RedirectURL string
	AuthUser    *User
	Error       error
}

type EmptyProvider struct{}

func (e *EmptyProvider) Logout(c buffalo.Context) Response {
	return Response{}
}

func (e *EmptyProvider) AuthRequest(c buffalo.Context) (string, error) {
	return "", nil
}

func (e *EmptyProvider) AuthCallback(c buffalo.Context) Response {
	return Response{}
}
