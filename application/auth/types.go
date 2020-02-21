package auth

import (
	"errors"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/auth/facebook"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/auth/linkedin"
	"github.com/silinternational/wecarry-api/auth/twitter"
	"github.com/silinternational/wecarry-api/domain"
)

const AuthTypeAzureAD = "azureadv2"
const AuthTypeFacebook = "facebook"
const AuthTypeGoogle = "google"
const AuthTypeLinkedIn = "linkedin"
const AuthTypeSaml = "saml"
const AuthTypeTwitter = "twitter"

const envTypeKey = "key"
const envTypeSecret = "secret"

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
	PhotoURL  string
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

type SocialAuthConfig struct {
	Key    string
	Secret string
}

// Get the necessary env vars to create the associated SocialAuthConfig
func getConfig(authType string, envVars map[string]string) SocialAuthConfig {
	combinedLen := 0
	for _, value := range envVars {
		combinedLen += len(value)
	}

	if combinedLen == 0 {
		return SocialAuthConfig{}
	}

	for key, value := range envVars {
		if value == "" {
			domain.ErrLogger.Printf("missing %s %s env variable.", authType, key)
			return SocialAuthConfig{}
		}
	}

	config := SocialAuthConfig{}

	config.Key = envVars[envTypeKey]
	config.Secret = envVars[envTypeSecret]
	return config
}

func addFacebookConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeFacebook
	envTypes := map[string]string{
		envTypeKey:    domain.Env.FacebookKey,
		envTypeSecret: domain.Env.FacebookSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

func addGoogleConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeGoogle
	envTypes := map[string]string{
		envTypeKey:    domain.Env.GoogleKey,
		envTypeSecret: domain.Env.GoogleSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

func addLinkedInConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeLinkedIn
	envTypes := map[string]string{
		envTypeKey:    domain.Env.LinkedInKey,
		envTypeSecret: domain.Env.LinkedInSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

func addTwitterConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeTwitter
	envTypes := map[string]string{
		envTypeKey:    domain.Env.TwitterKey,
		envTypeSecret: domain.Env.TwitterSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

// GetSocialAuthConfigs returns a map of the enabled Social Auth Provider configs
func GetSocialAuthConfigs() map[string]SocialAuthConfig {

	configs := map[string]SocialAuthConfig{}

	// Maps act as pass-by-reference, so configs gets modified in place
	addFacebookConfig(configs)
	addGoogleConfig(configs)
	addLinkedInConfig(configs)
	addTwitterConfig(configs)

	return configs
}

func GetSocialAuthProvider(authType string) (Provider, error) {
	_, ok := socialAuthConfigs[authType]
	if !ok {
		return nil, errors.New("unknown Social Auth Provider: " + authType)
	}

	switch authType {
	case AuthTypeFacebook:
		return facebook.New([]byte(""))
	case AuthTypeGoogle:
		return google.New([]byte(""))
	case AuthTypeLinkedIn:
		return linkedin.New([]byte(""))
	case AuthTypeTwitter:
		return twitter.New([]byte(""))
	}

	return nil, errors.New("unmatched Social Auth Provider: " + authType)
}
