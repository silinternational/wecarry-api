package actions

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/auth/facebook"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/auth/linkedin"
	"github.com/silinternational/wecarry-api/auth/twitter"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const (
	// Auth Type Identifiers
	AuthTypeFacebook = "facebook"
	AuthTypeGoogle   = "google"
	AuthTypeLinkedIn = "linkedin"
	AuthTypeTwitter  = "twitter"

	envTypeKey    = "key"
	envTypeSecret = "secret"
)

// SocialAuthConfig holds the Key and Secret for a social auth provider
type SocialAuthConfig struct{ Key, Secret string }

// Don't Modify outside of the code in this file that sets it.
var socialAuthConfigs = map[string]SocialAuthConfig{}

// Get the necessary env vars to create the associated SocialAuthConfig
func getConfig(authType string, envVars map[string]string) SocialAuthConfig {
	config := SocialAuthConfig{}
	combinedLen := 0
	for _, value := range envVars {
		combinedLen += len(value)
	}

	if combinedLen == 0 {
		return config
	}

	for key, value := range envVars {
		if value == "" {
			domain.ErrLogger.Printf("missing %s %s env variable.", authType, key)
			return config
		}
	}

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

// getSocialAuthConfigs returns a map of the enabled Social Auth Provider configs
func getSocialAuthConfigs() map[string]SocialAuthConfig {

	configs := socialAuthConfigs
	// Don't keep adding the social auth providers after we've already done it once.
	if len(configs) > 0 {
		return configs
	}

	// Maps act as pass-by-reference, so configs gets modified in place
	addFacebookConfig(configs)
	addGoogleConfig(configs)
	addLinkedInConfig(configs)
	addTwitterConfig(configs)

	return configs
}

func getSocialAuthProvider(authType string) (auth.Provider, error) {
	authConfigs := getSocialAuthConfigs()
	config, ok := authConfigs[authType]
	if !ok {
		return nil, errors.New("unknown Social Auth Provider: " + authType)
	}

	switch authType {
	case AuthTypeFacebook:
		return facebook.New(config)
	case AuthTypeGoogle:
		return google.New(config, []byte{})
	case AuthTypeLinkedIn:
		return linkedin.New(config)
	case AuthTypeTwitter:
		return twitter.New(config)
	}

	return nil, errors.New("unmatched Social Auth Provider: " + authType)
}

func socialLoginBasedAuthCallback(c buffalo.Context, authEmail, clientID string) error {

	authType, ok := c.Session().Get(SocialAuthTypeSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.ErrorMissingSessionSocialAuthType,
			SocialAuthTypeSessionKey+" session entry is required to complete login")
	}

	extras := map[string]interface{}{"authEmail": authEmail}

	ap, err := getSocialAuthProvider(authType)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("error loading social auth provider for '%s' ... %v", authType, err), extras)
	}

	authResp := ap.AuthCallback(c)
	if authResp.Error != nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersCallback, authResp.Error.Error(), extras)
	}

	returnTo := getOrSetReturnTo(c)

	if authResp.AuthUser == nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersCallback, "nil authResp.AuthUser", extras)
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User

	if authEmail != authResp.AuthUser.Email {
		c.Session().Clear()
		extras := map[string]interface{}{"authEmail": authEmail}
		return logErrorAndRedirect(c, domain.ErrorAuthEmailMismatch, err.Error(), extras)
	}

	// Check for an invite in the Session
	inviteType, objectUUID := getInviteInfoFromSession(c)

	// login was success, clear session so new login can be initiated if needed
	c.Session().Clear()

	if err := user.FindOrCreateFromOrglessAuthUser(authResp.AuthUser); err != nil {
		return logErrorAndRedirect(c, domain.ErrorWithAuthUser, err.Error())
	}

	if inviteType != "" {
		dealWithInviteFromCallback(c, inviteType, objectUUID, user)
	}

	authUser, err := createOrglessAuthUser(clientID, user)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, returnTo))
}

func createOrglessAuthUser(clientID string, user models.User) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateOrglessAccessToken(clientID)

	if err != nil {
		return AuthUser{}, err
	}

	isNew := false
	if time.Since(user.CreatedAt) < time.Duration(time.Second*30) {
		isNew = true
	}

	authUser := AuthUser{
		ID:                   user.UUID.String(),
		Name:                 user.FirstName + " " + user.LastName,
		Nickname:             user.Nickname,
		Email:                user.Email,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		IsNew:                isNew,
	}

	return authUser, nil
}
