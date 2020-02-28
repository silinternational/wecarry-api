package actions

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

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

// If there is a Key and Secret, then add them to the auth providers' configs
// Maps act as pass-by-reference, so configs gets modified in place
func addConfig(authType, key, secret string, configs map[string]SocialAuthConfig) {
	if key == "" || secret == "" {
		return
	}
	configs[authType] = SocialAuthConfig{Key: key, Secret: secret}
}

func addFacebookConfig(configs map[string]SocialAuthConfig) {
	addConfig(AuthTypeFacebook, domain.Env.FacebookKey, domain.Env.FacebookSecret, configs)
}

func addGoogleConfig(configs map[string]SocialAuthConfig) {
	addConfig(AuthTypeGoogle, domain.Env.GoogleKey, domain.Env.GoogleSecret, configs)
}

func addLinkedInConfig(configs map[string]SocialAuthConfig) {
	addConfig(AuthTypeLinkedIn, domain.Env.LinkedInKey, domain.Env.LinkedInSecret, configs)
}

func addTwitterConfig(configs map[string]SocialAuthConfig) {
	addConfig(AuthTypeTwitter, domain.Env.TwitterKey, domain.Env.TwitterSecret, configs)
}

// Forbidden. Do NOT call this function.  (Only called by init() and a test)
func getSocialAuthConfigs() map[string]SocialAuthConfig {
	configs := map[string]SocialAuthConfig{}

	// Maps act as pass-by-reference, so configs gets modified in place
	addFacebookConfig(configs)
	addGoogleConfig(configs)
	addLinkedInConfig(configs)
	addTwitterConfig(configs)

	return configs
}

func getSocialAuthProvider(authType string) (auth.Provider, error) {
	config, ok := socialAuthConfigs[authType]
	if !ok {
		return nil, errors.New("unknown Social Auth Provider: " + authType + ".")
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

	return nil, errors.New("unmatched Social Auth Provider: " + authType + ".")
}

// Forbidden. Do NOT call this function.  (Only called by init() and a test)
func getSocialAuthSelectors(authConfigs map[string]SocialAuthConfig) []authSelector {
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
	return selectors
}

// For non invite based ... based on a User in the database with a SocialAuthProvider value
func finishAuthRequestForSocialUser(c buffalo.Context, authEmail string) error {
	var user models.User
	if err := user.FindByEmail(authEmail); err != nil {
		authErr := authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   domain.ErrorFindingUserByEmail,
			errorMsg:   err.Error(),
		}
		return authRequestError(c, authErr)
	}

	authType := user.SocialAuthProvider.String
	extras := map[string]interface{}{"authType": authType}

	ap, err := getSocialAuthProvider(authType)
	if err != nil {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorLoadingAuthProvider,
			errorMsg:   fmt.Sprintf("error loading social auth provider, %v", err),
		}
		return authRequestError(c, authErr, extras)
	}

	redirectURL, err := ap.AuthRequest(c)
	if err != nil {
		authErr := authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   domain.ErrorGettingAuthURL,
			errorMsg:   fmt.Sprintf("error getting social auth url, %v", err),
		}
		return authRequestError(c, authErr, extras)
	}

	// UI expects an array, even when there is only one option
	authOptions := []authOption{{Name: authType, RedirectURL: redirectURL}}

	c.Session().Set(SocialAuthTypeSessionKey, authType)

	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(authOptions))

}

// Just get the list of auth/select/... URLS
func finishInviteBasedSocialAuthRequest(c buffalo.Context, extras map[string]interface{}) error {
	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(socialAuthSelectors))
}

// Redirect user to their selected social auth provider
func authSelect(c buffalo.Context) error {
	authType := c.Param(AuthTypeParam)
	if authType == "" {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorMissingAuthType,
			errorMsg:   AuthTypeParam + " is required to login",
		}
		return authRequestError(c, authErr)
	}

	extras := map[string]interface{}{"authType": authType}

	ap, err := getSocialAuthProvider(authType)
	if err != nil {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorLoadingAuthProvider,
			errorMsg:   fmt.Sprintf("error loading social auth provider, %v", err),
		}
		return authRequestError(c, authErr, extras)
	}

	redirectURL, err := ap.AuthRequest(c)
	if err != nil {
		authErr := authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   domain.ErrorGettingAuthURL,
			errorMsg:   fmt.Sprintf("error getting social auth url, %v", err),
		}
		return authRequestError(c, authErr, extras)
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
	if err := user.FindByEmailAndSocialAuthProvider(authEmail, authType); err != nil {
		return logErrorAndRedirect(c, domain.ErrorGettingSocialAuthUser,
			fmt.Sprintf("error loading social auth user for '%s' ... %v", authType, err), extras)
	}

	callbackValues := processSocialAuthCallback(c, authEmail, authType)
	if callbackValues.errCode != "" {
		return logErrorAndRedirect(c, callbackValues.errCode, callbackValues.errMsg, extras)
	}

	authUser, err := newOrglessAuthUser(clientID, user)
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

	// If the invite is for a meeting, ensure there is a MeetingParticipant record
	// that matches the MeetingInvite
	if inviteType != "" {
		dealWithInviteFromCallback(c, inviteType, inviteObjectUUID, user)
	}

	authUser, err := newOrglessAuthUser(clientID, user)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, callbackValues.returnTo))
}

// Hydrates an AuthUser struct based on a User without an Org
func newOrglessAuthUser(clientID string, user models.User) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateOrglessAccessToken(clientID)
	if err != nil {
		return AuthUser{}, err
	}

	return hydrateAuthUser(user, accessToken, expiresAt), nil
}
