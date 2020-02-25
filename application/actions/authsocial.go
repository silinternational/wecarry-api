package actions

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/auth/facebook"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/auth/linkedin"
	"github.com/silinternational/wecarry-api/auth/twitter"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const (
	// http param for social auth provider
	AuthTypeParam = "auth-type"

	// Auth Type Identifiers
	AuthTypeFacebook = "facebook"
	AuthTypeGoogle   = "google"
	AuthTypeLinkedIn = "linkedin"
	AuthTypeTwitter  = "twitter"

	envSocialAuthKey    = "key"
	envSocialAuthSecret = "secret"

	AuthSelectPath = "%s/auth/select/?%s=%s"
)

// SocialAuthConfig holds the Key and Secret for a social auth provider
type SocialAuthConfig struct{ Key, Secret string }

// Don't Modify outside of this file.
var socialAuthConfigs = map[string]SocialAuthConfig{}

// Don't Modify outside of this file.
var socialAuthSelectors = []authSelector{}

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

	config.Key = envVars[envSocialAuthKey]
	config.Secret = envVars[envSocialAuthSecret]
	return config
}

func addFacebookConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeFacebook
	envTypes := map[string]string{
		envSocialAuthKey:    domain.Env.FacebookKey,
		envSocialAuthSecret: domain.Env.FacebookSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

func addGoogleConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeGoogle
	envTypes := map[string]string{
		envSocialAuthKey:    domain.Env.GoogleKey,
		envSocialAuthSecret: domain.Env.GoogleSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

func addLinkedInConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeLinkedIn
	envTypes := map[string]string{
		envSocialAuthKey:    domain.Env.LinkedInKey,
		envSocialAuthSecret: domain.Env.LinkedInSecret,
	}
	config := getConfig(authType, envTypes)
	if config.Key != "" {
		configs[authType] = config
	}
}

func addTwitterConfig(configs map[string]SocialAuthConfig) {
	authType := AuthTypeTwitter
	envTypes := map[string]string{
		envSocialAuthKey:    domain.Env.TwitterKey,
		envSocialAuthSecret: domain.Env.TwitterSecret,
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

func getSocialAuthSelectors() []authSelector {
	if len(socialAuthSelectors) > 0 {
		return socialAuthSelectors
	}
	authConfigs := getSocialAuthConfigs()
	selectors := []authSelector{}
	for pType, _ := range authConfigs {
		s := authSelector{
			Name:        pType,
			RedirectURL: fmt.Sprintf(AuthSelectPath, domain.Env.ApiBaseURL, AuthTypeParam, pType),
		}
		selectors = append(selectors, s)

	}
	socialAuthSelectors = selectors
	return selectors
}

func finishSocialAuthRequest(c buffalo.Context, extras map[string]interface{}) error {
	selectors := getSocialAuthSelectors()

	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(selectors))
}

func authSelect(c buffalo.Context) error {
	authType := c.Param(AuthTypeParam)
	if authType == "" {
		return authRequestError(c, http.StatusBadRequest, domain.ErrorMissingAuthType,
			AuthTypeParam+" is required to login")
	}

	ap, err := getSocialAuthProvider(authType)
	if err != nil {
		return authRequestError(c, http.StatusBadRequest, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("error loading social auth provider for '%s' ... %v", authType, err))
	}

	redirectURL, err := ap.AuthRequest(c)
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorGettingAuthURL,
			fmt.Sprintf("error getting social auth url for '%s' ... %v", authType, err))
	}

	c.Session().Set(SocialAuthTypeSessionKey, authType)

	return c.Redirect(http.StatusFound, redirectURL)
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
