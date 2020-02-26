package actions

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
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

	// sort the provider types for ease of testing (avoid map's random order)
	pTypes := []string{}
	for pt, _ := range authConfigs {
		pTypes = append(pTypes, pt)
	}

	sort.Strings(pTypes)

	selectors := []authSelector{}
	for _, pt := range pTypes {
		s := authSelector{
			Name:        pt,
			RedirectURL: fmt.Sprintf(AuthSelectPath, domain.Env.ApiBaseURL, AuthTypeParam, pt),
		}
		selectors = append(selectors, s)

	}
	socialAuthSelectors = selectors
	return selectors
}

func finishAuthRequestForSocialUser(c buffalo.Context, authEmail string) error {
	var user models.User
	if err := user.FindByEmail(authEmail); err != nil {
		return authRequestError(c, http.StatusNotFound, domain.ErrorFindingUserByEmail, err.Error())
	}

	authType := user.SocialAuthProvider.String

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

	authOptions := []authOption{{Name: authType, RedirectURL: redirectURL}}

	c.Session().Set(SocialAuthTypeSessionKey, authType)

	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(authOptions))

}

func finishInviteBasedSocialAuthRequest(c buffalo.Context, extras map[string]interface{}) error {
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

type callbackValues struct {
	authResp auth.Response
	returnTo string
	errCode  string
	errMsg   string
}

// processSocialAuthCallback is a function that holds code common to both
// users logging in based on an invite and those that already have a User
// record.  It gets the appropriate social auth provider,
// calls its AuthCallback function and checks its authResp values
// gets the ReturnTo from the session and clears the session.
func processSocialAuthCallback(c buffalo.Context, authEmail, authType string) callbackValues {
	ap, err := getSocialAuthProvider(authType)
	if err != nil {
		return callbackValues{
			errCode: domain.ErrorLoadingAuthProvider,
			errMsg:  fmt.Sprintf("error loading social auth provider for '%s' ... %v", authType, err),
		}
	}

	authResp := ap.AuthCallback(c)
	if authResp.Error != nil {
		return callbackValues{
			errCode: domain.ErrorAuthProvidersCallback,
			errMsg:  authResp.Error.Error(),
		}
	}

	returnTo := getOrSetReturnTo(c)

	if authResp.AuthUser == nil {
		return callbackValues{
			errCode: domain.ErrorAuthProvidersCallback,
			errMsg:  "nil authResp.AuthUser",
		}
	}

	if authEmail != authResp.AuthUser.Email {
		c.Session().Clear()
		return callbackValues{
			errCode: domain.ErrorAuthEmailMismatch,
			errMsg:  err.Error(),
		}
	}

	c.Session().Clear()

	return callbackValues{
		authResp: authResp,
		returnTo: returnTo,
	}
}

// Finish the auth callback process for an Orgless user that has no invite associated
// with this login.
func socialLoginNonInviteBasedAuthCallback(c buffalo.Context, authEmail, authType, clientID string) error {
	extras := map[string]interface{}{"authEmail": authEmail, "authType": authType}

	var user models.User
	if err := user.FindBySocialAuthProvider(authEmail, authType); err != nil {
		return logErrorAndRedirect(c, domain.ErrorGettingSocialAuthUser,
			fmt.Sprintf("error loading social auth user for '%s' ... %v", authType, err), extras)
	}

	callbackValues := processSocialAuthCallback(c, authEmail, authType)
	if callbackValues.errCode != "" {
		return logErrorAndRedirect(c, callbackValues.errCode, callbackValues.errMsg, extras)
	}

	authUser, err := createOrglessAuthUser(clientID, user)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, callbackValues.returnTo))
}

func socialLoginBasedAuthCallback(c buffalo.Context, authEmail, clientID string) error {

	authType, ok := c.Session().Get(SocialAuthTypeSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.ErrorMissingSessionSocialAuthType,
			SocialAuthTypeSessionKey+" session entry is required to complete login")
	}

	extras := map[string]interface{}{"authEmail": authEmail, "authType": authType}

	// Check for an invite in the Session
	inviteType, inviteObjectUUID := getInviteInfoFromSession(c)

	// If there is no invite associated with this user, then deal with the User record
	if inviteObjectUUID == "" {
		return socialLoginNonInviteBasedAuthCallback(c, authEmail, authType, clientID)
	}

	// There is an invite associated with this process, so deal ith it
	callbackValues := processSocialAuthCallback(c, authEmail, authType)
	if callbackValues.errCode != "" {
		return logErrorAndRedirect(c, callbackValues.errCode, callbackValues.errMsg, extras)
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User

	if err := user.FindOrCreateFromOrglessAuthUser(callbackValues.authResp.AuthUser, authType); err != nil {
		return logErrorAndRedirect(c, domain.ErrorWithAuthUser, err.Error())
	}

	if inviteType != "" {
		dealWithInviteFromCallback(c, inviteType, inviteObjectUUID, user)
	}

	authUser, err := createOrglessAuthUser(clientID, user)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, callbackValues.returnTo))
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
