package auth

import (
	"github.com/gobuffalo/buffalo"
)

const TypeSaml = "saml"

// Provider interface to be implemented by any auth providers
// It is expected Login can be called multiple times, whether during initializing a login request or when
// processing a authentication response.
type Provider interface {
	Login(c buffalo.Context) Response
	Logout(c buffalo.Context) Response
}

// User holds common attributes expected from auth providers
type User struct {
	FirstName string
	LastName  string
	Email     string
	UserID    string
}

// Response holds fields for login and logout responses. not all fields will have values
type Response struct {
	RedirectURL string
	AuthUser    *User
	AuthEmail   string
	ClientID    string
	ReturnTo    string
	Error       error
}

type EmptyProvider struct{}

func (e *EmptyProvider) Login(c buffalo.Context) Response {
	return Response{}
}
func (e *EmptyProvider) Logout(c buffalo.Context) Response {
	return Response{}
}
