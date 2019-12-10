// +build !go1.9

package google

import (
	"golang.org/x/oauth2"
)

// endpoint is Google's OAuth 2.0 endpoint.
var endpoint = oauth2.Endpoint{
	AuthURL:  "https://accounts.google.com/o/oauth2/auth",
	TokenURL: "https://accounts.google.com/o/oauth2/token",
}
