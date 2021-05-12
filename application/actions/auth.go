package actions

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const (
	// http param for access token
	AccessTokenParam = "access-token"

	// http param and session key for Auth Email
	AuthEmailParam      = "auth-email"
	AuthEmailSessionKey = "AuthEmail"

	// http param and session key for Client ID
	ClientIDParam      = "client-id"
	ClientIDSessionKey = "ClientID"

	// http params for the Invite type and code
	InviteCodeParam        = "code"
	InviteTypeMeetingParam = "meeting"

	// session keys for using invites for authentication
	InviteTypeSessionKey       = "InviteType"
	InviteObjectUUIDSessionKey = "InviteObjectID"

	// logout http param for what is normally the bearer token
	LogoutToken = "token"

	// http param for organization id
	OrgIDParam      = "org-id"
	OrgIDSessionKey = "OrgID"

	// http param and session key for ReturnTo
	ReturnToParam      = "return-to"
	ReturnToSessionKey = "ReturnTo"

	// session key for Socual Auth ID
	SocialAuthTypeSessionKey = "SocialAuthType"

	// http param for token type
	TokenTypeParam = "token-type"
)

type authOption struct {
	Name        string `json:"Name"`
	RedirectURL string `json:"RedirectURL"`
}

type authInviteResponse struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	ImageURL string `json:"imageURL"`
}

type AuthUser struct {
	ID                   string `json:"ID"`
	Name                 string `json:"Name"`
	Nickname             string `json:"Nickname"`
	Email                string `json:"Email"`
	AccessToken          string `json:"AccessToken"`
	AccessTokenExpiresAt int64  `json:"AccessTokenExpiresAt"`
	IsNew                bool   `json:"IsNew"`
}

func getOrSetReturnTo(c buffalo.Context) string {
	returnTo := c.Param(ReturnToParam)

	if returnTo == "" {
		var ok bool
		returnTo, ok = c.Session().Get(ReturnToSessionKey).(string)
		if !ok {
			returnTo = domain.DefaultUIPath
		}

		return returnTo
	}

	c.Session().Set(ReturnToSessionKey, returnTo)

	return returnTo
}

// Gets info about an invite object (e.g. a meeting) based on an invite `code`
func getAuthInviteResponse(c buffalo.Context) (authInviteResponse, error) {
	inviteCode := c.Param(InviteCodeParam)

	if inviteCode == "" {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorInvalidInviteCode,
			errorMsg:   "missing Invite Code.",
		}
		return authInviteResponse{}, authRequestError(c, authErr)
	}

	extras := map[string]interface{}{"authInviteCode": inviteCode}

	var meeting models.Meeting
	if err := meeting.FindByInviteCode(inviteCode); err != nil {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   domain.ErrorInvalidInviteCode,
			errorMsg:   "error validating Invite Code: " + err.Error(),
		}
		return authInviteResponse{}, authRequestError(c, authErr, extras)
	}

	c.Session().Set(InviteTypeSessionKey, InviteTypeMeetingParam)
	c.Session().Set(InviteObjectUUIDSessionKey, meeting.UUID.String())

	resp := authInviteResponse{
		Type: InviteTypeMeetingParam,
		Name: meeting.Name,
	}
	if meeting.FileID.Valid {
		f, err := meeting.ImageFile()
		if err != nil {
			domain.ErrLogger.Printf("error loading meeting image file: " + err.Error())
		}
		resp.ImageURL = f.URL
	}
	return resp, nil
}

// Gets info about an invite object (e.g. a meeting) based on an invite `code`
func authInvite(c buffalo.Context) error {
	// push most of the code out to make it easily testable
	resp, err := getAuthInviteResponse(c)
	if err != nil {
		return err
	}

	if resp.Name == "" {
		return nil
	}

	return c.Render(http.StatusOK, render.JSON(resp))
}

func getUserOrgs(c buffalo.Context, authEmail string) (models.UserOrganizations, error) {
	var orgID int
	oid := c.Param(OrgIDParam)

	if oid != "" {
		orgID, _ = strconv.Atoi(oid)
	}

	var userOrgs models.UserOrganizations
	if err := userOrgs.FindByAuthEmail(authEmail, orgID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return userOrgs, err
		}
	}

	return userOrgs, nil
}

// no user_organization records yet, see if we have an organization for user's email domain
func getOrgForNewUser(authEmail string) (models.Organization, error) {
	var org models.Organization
	return org, org.FindByDomain(domain.EmailDomain(authEmail))
}

type authError struct {
	httpStatus int
	errorKey   string
	errorMsg   string
}

func getOrgBasedAuthOption(c buffalo.Context, authEmail string, org models.Organization) (authOption, *authError) {
	c.Session().Set(OrgIDSessionKey, org.UUID.String())

	sp, err := org.GetAuthProvider(authEmail)
	if err != nil {
		return authOption{}, &authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   domain.ErrorLoadingAuthProvider,
			errorMsg:   fmt.Sprintf("unable to load auth provider for '%s' ... %v", org.Name, err),
		}
	}

	redirectURL, err := sp.AuthRequest(c)
	if err != nil {
		return authOption{}, &authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   domain.ErrorGettingAuthURL,
			errorMsg: fmt.Sprintf("unable to determine what the authentication url should be for '%s' ... %v",
				org.Name, err),
		}
	}

	option := authOption{Name: org.Name, RedirectURL: redirectURL}

	return option, nil
}

// Decide whether a meeting invitee should use social login or org-based login
func meetingAuthRequest(c buffalo.Context, authEmail string, extras map[string]interface{}) error {
	meetingUUID, ok := c.Session().Get(InviteObjectUUIDSessionKey).(string)
	if !ok {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorMissingSessionInviteObjectUUID,
			errorMsg:   InviteObjectUUIDSessionKey + " session entry is required to complete login for an invite",
		}
		return authRequestError(c, authErr, extras)
	}

	var meeting models.Meeting
	if err := meeting.FindByUUID(meetingUUID); err != nil {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   domain.ErrorInvalidSessionInviteObjectUUID,
			errorMsg: InviteObjectUUIDSessionKey +
				" session entry caused an error finding the related meeting: " + err.Error(),
		}
		return authRequestError(c, authErr, extras)
	}

	// If the user already has organizational auth options, use them
	userOrgs, err := getUserOrgs(c, authEmail)
	if err != nil {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   domain.ErrorFindingUserOrgs,
			errorMsg:   "error getting UserOrganizations: " + err.Error(),
		}
		return authRequestError(c, authErr, extras)
	}

	if len(userOrgs) > 0 {
		return finishOrgBasedAuthRequest(c, authEmail, userOrgs, extras)
	}

	// Check if user's email has a domain that matches an Organization
	org, err := getOrgForNewUser(authEmail)
	if err != nil && domain.IsOtherThanNoRows(err) {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   domain.ErrorFindingOrgForNewUser,
			errorMsg:   "error getting UserOrganizations: " + err.Error(),
		}
		return authRequestError(c, authErr, extras)
	}

	// If there is a matching Org, use that
	if org.AuthType != "" {
		var userOrg models.UserOrganization
		userOrg.OrganizationID = org.ID
		userOrg.Organization = org

		return finishOrgBasedAuthRequest(c, authEmail, models.UserOrganizations{userOrg}, extras)
	}

	// If no matching Org, then use social login
	return finishInviteBasedSocialAuthRequest(c, extras)
}

// Decide whether an invitee should use social login or org-based login
func inviteAuthRequest(c buffalo.Context, authEmail, inviteType string) error {
	extras := map[string]interface{}{"authEmail": authEmail, "inviteType": inviteType}
	if inviteType == "" {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorCannotFindOrg,
			errorMsg:   "unable to find an organization for this user",
		}
		return authRequestError(c, authErr, extras)
	}

	switch inviteType {
	case InviteTypeMeetingParam:
		return meetingAuthRequest(c, authEmail, extras)
	default:
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorInvalidInviteType,
			errorMsg:   "Invite Type '" + inviteType + "' is not valid.",
		}
		return authRequestError(c, authErr, extras)
	}

	return nil
}

// Hydrates an AuthUser struct based on a user with an Org
func newOrgBasedAuthUser(clientID string, user models.User, org models.Organization) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateAccessToken(org, clientID)
	if err != nil {
		return AuthUser{}, err
	}

	return hydrateAuthUser(user, accessToken, expiresAt), nil
}

func hydrateAuthUser(user models.User, accessToken string, accessTokenExpiresAt int64) AuthUser {
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
		AccessTokenExpiresAt: accessTokenExpiresAt,
		IsNew:                isNew,
	}

	return authUser
}

func finishOrgBasedAuthRequest(c buffalo.Context, authEmail string,
	userOrgs models.UserOrganizations, extras map[string]interface{}) error {

	authOptions := []authOption{}

	for _, uo := range userOrgs {
		option, authErr := getOrgBasedAuthOption(c, authEmail, uo.Organization)
		if authErr != nil {
			return authRequestError(c, *authErr, extras)
		}
		authOptions = append(authOptions, option)
	}

	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(authOptions))
}

func authRequest(c buffalo.Context) error {
	// Push the Client ID into the Session
	clientID := c.Param(ClientIDParam)
	if clientID == "" {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorMissingClientID,
			errorMsg:   ClientIDParam + " is required to login",
		}
		return authRequestError(c, authErr)
	}

	c.Session().Set(ClientIDSessionKey, clientID)

	// Get the AuthEmail param and push it into the Session
	authEmail := c.Param(AuthEmailParam)
	if authEmail == "" {
		authErr := authError{
			httpStatus: http.StatusBadRequest,
			errorKey:   domain.ErrorMissingAuthEmail,
			errorMsg:   AuthEmailParam + " is required to login",
		}
		return authRequestError(c, authErr)
	}
	c.Session().Set(AuthEmailSessionKey, authEmail)

	getOrSetReturnTo(c)

	// If there is an invite in the Session, use the invite Auth Request process
	inviteType, ok := c.Session().Get(InviteTypeSessionKey).(string)
	if ok {
		return inviteAuthRequest(c, authEmail, inviteType)
	}

	extras := map[string]interface{}{"authEmail": authEmail}

	// find org for auth config and processing
	userOrgs, err := getUserOrgs(c, authEmail)
	if err != nil {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   domain.ErrorFindingUserOrgs,
			errorMsg:   "error getting UserOrganizations: " + err.Error(),
		}
		return authRequestError(c, authErr, extras)
	}

	if len(userOrgs) > 0 {
		return finishOrgBasedAuthRequest(c, authEmail, userOrgs, extras)
	}

	// Check if user's email has a domain that matches an Organization
	org, err := getOrgForNewUser(authEmail)
	if err != nil {
		if domain.IsOtherThanNoRows(err) {
			authErr := authError{
				httpStatus: http.StatusNotFound,
				errorKey:   domain.ErrorFindingOrgForNewUser,
				errorMsg:   "error getting UserOrganizations: " + err.Error(),
			}
			return authRequestError(c, authErr, extras)
		}

		return finishAuthRequestForSocialUser(c, authEmail)
	}

	// If there is a matching Org, use that
	if org.AuthType != "" {
		var userOrg models.UserOrganization
		userOrg.OrganizationID = org.ID
		userOrg.Organization = org

		return finishOrgBasedAuthRequest(c, authEmail, models.UserOrganizations{userOrg}, extras)
	}

	// If no matching Org, then error, since this isn't based on an invite
	authErr := authError{
		httpStatus: http.StatusNotFound,
		errorKey:   domain.ErrorOrglessUserNotAllowed,
		errorMsg:   "no Organization found for this authEmail",
	}
	return authRequestError(c, authErr, extras)
}

// If there is a MeetingInvite for this user, then ensure there is also a
// MeetingParticipant for them.
func ensureMeetingParticipant(c buffalo.Context, meetingUUID string, user models.User) {
	var meeting models.Meeting
	if err := meeting.FindByUUID(meetingUUID); err != nil {
		domain.Error(c, "expected to find a Meeting but got "+err.Error())
	}

	// If there is already a MeetingParticipant record for this user, we're done
	var participant models.MeetingParticipant
	if err := participant.FindByMeetingIDAndUserID(meeting.ID, user.ID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			domain.Error(c, "error finding a MeetingParticpant: "+err.Error())
		}
	} else {
		return
	}

	// Try to create a MeetingParticipant record for this user.
	var invite models.MeetingInvite
	if err := invite.FindByMeetingIDAndEmail(meeting.ID, user.Email); err != nil {
		domain.Error(c, "expected to find a MeetingInvite but got "+err.Error())
		return
	}

	if err := participant.CreateFromInvite(invite, user.ID); err != nil {
		domain.Error(c, "error creating a MeetingParticipant: "+err.Error())
	}
}

// Deals with the situation when a user logins as a response to an invite
func dealWithInviteFromCallback(c buffalo.Context, inviteType, objectUUID string, user models.User) {
	switch inviteType {
	case InviteTypeMeetingParam:
		ensureMeetingParticipant(c, objectUUID, user)
	default:
		domain.Error(c, "incorrect meeting invite type in session: "+inviteType)
	}
}

func getInviteInfoFromSession(c buffalo.Context) (string, string) {
	inviteType, ok := c.Session().Get(InviteTypeSessionKey).(string)
	if !ok {
		return "", ""
	}

	objectUUID, ok := c.Session().Get(InviteObjectUUIDSessionKey).(string)
	if !ok {
		domain.Error(c, "got meeting invite type from session but not its UUID")
		return "", ""
	}
	return inviteType, objectUUID
}

func orgBasedAuthCallback(c buffalo.Context, orgUUID, authEmail, clientID string) error {
	org := models.Organization{}
	err := org.FindByUUID(orgUUID)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingOrgByID,
			fmt.Sprintf("error finding org with UUID %s ... %v", orgUUID, err.Error()))
	}

	domain.NewExtra(c, "authEmail", authEmail)

	ap, err := org.GetAuthProvider(authEmail)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("error loading auth provider for '%s' ... %v", org.Name, err))
	}

	authResp := ap.AuthCallback(c)
	if authResp.Error != nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersCallback, authResp.Error.Error())
	}

	returnTo := getOrSetReturnTo(c)

	if authResp.AuthUser == nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersCallback, "nil authResp.AuthUser")
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User

	if err := verifyEmails(c, authEmail, authResp.AuthUser.Email); err != nil {
		c.Session().Clear()
		domain.NewExtra(c, "authEmail", authEmail)
		return logErrorAndRedirect(c, domain.ErrorAuthEmailMismatch, err.Error())
	}

	// Check for an invite in the Session
	inviteType, objectUUID := getInviteInfoFromSession(c)

	// login was success, clear session so new login can be initiated if needed
	c.Session().Clear()

	if err := user.FindOrCreateFromAuthUser(org.ID, authResp.AuthUser); err != nil {
		return logErrorAndRedirect(c, domain.ErrorWithAuthUser, err.Error())
	}

	if inviteType != "" {
		dealWithInviteFromCallback(c, inviteType, objectUUID, user)
	}

	authUser, err := newOrgBasedAuthUser(clientID, user, org)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, returnTo))
}

// authCallback assumes the user has logged in to the IDP or Oauth service and now their browser
// has been redirected back with the final response
func authCallback(c buffalo.Context) error {
	clientID, ok := c.Session().Get(ClientIDSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.ErrorMissingSessionClientID,
			ClientIDSessionKey+" session entry is required to complete login")
	}

	authEmail, ok := c.Session().Get(AuthEmailSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.ErrorMissingSessionAuthEmail,
			AuthEmailSessionKey+" session entry is required to complete login")
	}

	orgUUID, ok := c.Session().Get(OrgIDSessionKey).(string)
	if ok {
		return orgBasedAuthCallback(c, orgUUID, authEmail, clientID)
	}

	return socialLoginBasedAuthCallback(c, authEmail, clientID)
}

func verifyEmails(c buffalo.Context, originalAuthEmail, authRespEmail string) error {
	if originalAuthEmail == authRespEmail {
		return nil
	}

	emailDomain := domain.EmailDomain(originalAuthEmail)
	respDomain := domain.EmailDomain(authRespEmail)

	if emailDomain == respDomain {
		msg := "authentication emails don't match: " + originalAuthEmail + " vs. " + authRespEmail
		domain.Warn(c, msg)
		return nil
	}

	return errors.New("authentication email domains don't match: " + originalAuthEmail +
		" vs. " + authRespEmail)
}

func mergeExtras(code string, extras ...map[string]interface{}) map[string]interface{} {
	allExtras := map[string]interface{}{"code": code}

	for _, e := range extras {
		for k, v := range e {
			allExtras[k] = v
		}
	}
	return allExtras
}

// Make extras variadic, so that it can be omitted from the params
func authRequestError(c buffalo.Context, authErr authError, extras ...map[string]interface{}) error {
	domain.Error(c, authErr.errorMsg)

	appErr := domain.AppError{
		Code: authErr.httpStatus,
		Key:  authErr.errorKey,
	}

	c.Session().Clear()

	return c.Render(authErr.httpStatus, render.JSON(appErr))
}

// Make extras variadic, so that it can be omitted from the params
func logErrorAndRedirect(c buffalo.Context, code, message string) error {
	domain.Error(c, message)

	c.Session().Clear()

	uiUrl := domain.Env.UIURL + "/#/login"
	return c.Redirect(http.StatusFound, uiUrl)
}

// authDestroy uses the bearer token to find the user's access token and
//  calls the appropriate provider's logout function.
func authDestroy(c buffalo.Context) error {
	tokenParam := c.Param(LogoutToken)
	if tokenParam == "" {
		return logErrorAndRedirect(c, domain.ErrorMissingLogoutToken,
			LogoutToken+" is required to logout")
	}

	var uat models.UserAccessToken
	err := uat.FindByBearerToken(tokenParam)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingAccessToken, err.Error())
	}

	org, err := uat.GetOrganization()
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingOrgForAccessToken, err.Error())
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, uat.User.UUID.String(), uat.User.Nickname, uat.User.Email)

	authUser, err := uat.GetUser()
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersLogout, err.Error())
	}

	authPro, err := org.GetAuthProvider(authUser.Email)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorLoadingAuthProvider, err.Error())
	}

	authResp := authPro.Logout(c)
	if authResp.Error != nil {
		return logErrorAndRedirect(c, domain.ErrorAuthProvidersLogout, authResp.Error.Error())
	}

	redirectURL := domain.Env.UIURL

	if authResp.RedirectURL != "" {
		var uat models.UserAccessToken
		err = uat.DeleteByBearerToken(tokenParam)
		if err != nil {
			return logErrorAndRedirect(c, domain.ErrorDeletingAccessToken, err.Error())
		}
		c.Session().Clear()
		redirectURL = authResp.RedirectURL
	}

	return c.Redirect(http.StatusFound, redirectURL)
}

func setCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		bearerToken := domain.GetBearerTokenFromRequest(c.Request())
		if bearerToken == "" {
			return c.Error(http.StatusUnauthorized, errors.New("no Bearer token provided"))
		}

		var userAccessToken models.UserAccessToken
		err := userAccessToken.FindByBearerToken(bearerToken)
		if err != nil {
			if domain.IsOtherThanNoRows(err) {
				domain.Error(c, err.Error())
			}
			return c.Error(http.StatusUnauthorized, errors.New("invalid bearer token"))
		}

		isExpired, err := userAccessToken.DeleteIfExpired()
		if err != nil {
			domain.Error(c, err.Error())
		}

		if isExpired {
			return c.Error(http.StatusUnauthorized, errors.New("expired bearer token"))
		}

		user, err := userAccessToken.GetUser()
		if err != nil {
			return c.Error(http.StatusInternalServerError, fmt.Errorf("error finding user by access token, %s", err.Error()))
		}
		c.Set(domain.ContextKeyCurrentUser, user)

		// set person on rollbar session
		domain.RollbarSetPerson(c, user.UUID.String(), user.Nickname, user.Email)
		msg := fmt.Sprintf("user %s authenticated with bearer token from ip %s", user.Email, c.Request().RemoteAddr)
		domain.NewExtra(c, "user_id", user.ID)
		domain.NewExtra(c, "email", user.Email)
		domain.NewExtra(c, "ip", c.Request().RemoteAddr)
		domain.Info(c, msg)

		return next(c)
	}
}

// getLoginSuccessRedirectURL generates the URL for redirection after a successful login
func getLoginSuccessRedirectURL(authUser AuthUser, returnTo string) string {
	uiURL := domain.Env.UIURL

	params := fmt.Sprintf("?%s=Bearer&%s=%s",
		TokenTypeParam, AccessTokenParam, authUser.AccessToken)

	// New Users go straight to the welcome page
	if authUser.IsNew {
		uiURL += "/#/welcome"
		if len(returnTo) > 0 { // Ensure there is no `/#` at the beginning of the return-to param value
			if strings.HasPrefix(returnTo, `/#/`) {
				returnTo = returnTo[2:]
			}
			params += "&" + ReturnToParam + "=" + url.QueryEscape(returnTo)
		}
		return uiURL + params
	}

	// Ensure there is one set of /# between uiURL and the returnTo
	if !strings.HasPrefix(returnTo, `/#`) {
		returnTo = `/#` + returnTo
	}

	// Avoid two question marks in the params
	if strings.Contains(returnTo, "?") && strings.HasPrefix(params, "?") {
		params = "&" + params[1:]
	}

	return uiURL + returnTo + params
}
