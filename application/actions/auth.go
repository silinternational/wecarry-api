package actions

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v6"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
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
			errorKey:   api.ErrorInvalidInviteCode,
			errorMsg:   "missing Invite Code.",
		}
		return authInviteResponse{}, authRequestError(c, authErr)
	}

	extras := map[string]interface{}{"authInviteCode": inviteCode}

	var meeting models.Meeting
	tx := models.Tx(c)
	if err := meeting.FindByInviteCode(tx, inviteCode); err != nil {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   api.ErrorInvalidInviteCode,
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
		f, err := meeting.ImageFile(tx)
		if err != nil {
			log.WithContext(c).Errorf("error loading meeting image file: " + err.Error())
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
	if err := userOrgs.FindByAuthEmail(models.Tx(c), authEmail, orgID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return userOrgs, err
		}
	}

	return userOrgs, nil
}

// no user_organization records yet, see if we have an organization for user's email domain
func getOrgForNewUser(tx *pop.Connection, authEmail string) (models.Organization, error) {
	var org models.Organization
	return org, org.FindByDomain(tx, domain.EmailDomain(authEmail))
}

type authError struct {
	httpStatus int
	errorKey   api.ErrorKey
	errorMsg   string
}

func getOrgBasedAuthOption(c buffalo.Context, authEmail string, org models.Organization) (authOption, *authError) {
	c.Session().Set(OrgIDSessionKey, org.UUID.String())

	sp, err := org.GetAuthProvider(models.Tx(c), authEmail)
	if err != nil {
		return authOption{}, &authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   api.ErrorLoadingAuthProvider,
			errorMsg:   fmt.Sprintf("unable to load auth provider for '%s' ... %v", org.Name, err),
		}
	}

	redirectURL, err := sp.AuthRequest(c)
	if err != nil {
		return authOption{}, &authError{
			httpStatus: http.StatusInternalServerError,
			errorKey:   api.ErrorGettingAuthURL,
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
			errorKey:   api.ErrorMissingSessionInviteObjectUUID,
			errorMsg:   InviteObjectUUIDSessionKey + " session entry is required to complete login for an invite",
		}
		return authRequestError(c, authErr, extras)
	}

	var meeting models.Meeting
	tx := models.Tx(c)
	if err := meeting.FindByUUID(tx, meetingUUID); err != nil {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   api.ErrorInvalidSessionInviteObjectUUID,
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
			errorKey:   api.ErrorFindingUserOrgs,
			errorMsg:   "error getting UserOrganizations: " + err.Error(),
		}
		return authRequestError(c, authErr, extras)
	}

	if len(userOrgs) > 0 {
		return finishOrgBasedAuthRequest(c, authEmail, userOrgs, extras)
	}

	// Check if user's email has a domain that matches an Organization
	org, err := getOrgForNewUser(tx, authEmail)
	if err != nil && domain.IsOtherThanNoRows(err) {
		authErr := authError{
			httpStatus: http.StatusNotFound,
			errorKey:   api.ErrorFindingOrgForNewUser,
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
			errorKey:   api.ErrorCannotFindOrg,
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
			errorKey:   api.ErrorInvalidInviteType,
			errorMsg:   "Invite Type '" + inviteType + "' is not valid.",
		}
		return authRequestError(c, authErr, extras)
	}

	return nil
}

// Hydrates an AuthUser struct based on a user with an Org
func newOrgBasedAuthUser(ctx context.Context, clientID string, user models.User, org models.Organization) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateAccessToken(models.Tx(ctx), org, clientID)
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
	userOrgs models.UserOrganizations, extras map[string]interface{},
) error {
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
			errorKey:   api.ErrorMissingClientID,
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
			errorKey:   api.ErrorMissingAuthEmail,
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
			errorKey:   api.ErrorFindingUserOrgs,
			errorMsg:   "error getting UserOrganizations: " + err.Error(),
		}
		return authRequestError(c, authErr, extras)
	}

	if len(userOrgs) > 0 {
		return finishOrgBasedAuthRequest(c, authEmail, userOrgs, extras)
	}

	// Check if user's email has a domain that matches an Organization
	org, err := getOrgForNewUser(models.Tx(c), authEmail)
	if err != nil {
		if domain.IsOtherThanNoRows(err) {
			authErr := authError{
				httpStatus: http.StatusNotFound,
				errorKey:   api.ErrorFindingOrgForNewUser,
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
		errorKey:   api.ErrorOrglessUserNotAllowed,
		errorMsg:   "no Organization found for this authEmail",
	}
	return authRequestError(c, authErr, extras)
}

// If there is a MeetingInvite for this user, then ensure there is also a
// MeetingParticipant for them.
func ensureMeetingParticipant(c buffalo.Context, meetingUUID string, user models.User) {
	var meeting models.Meeting
	tx := models.Tx(c)
	if err := meeting.FindByUUID(tx, meetingUUID); err != nil {
		log.WithContext(c).Errorf("expected to find a Meeting but got %s", err)
	}

	// If there is already a MeetingParticipant record for this user, we're done
	var participant models.MeetingParticipant
	if err := participant.FindByMeetingIDAndUserID(tx, meeting.ID, user.ID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			log.WithContext(c).Errorf("error finding a MeetingParticpant: %s", err)
		}
	} else {
		return
	}

	// Try to create a MeetingParticipant record for this user.
	var invite models.MeetingInvite
	if err := invite.FindByMeetingIDAndEmail(tx, meeting.ID, user.Email); err != nil {
		log.WithContext(c).Errorf("expected to find a MeetingInvite but got %s", err)
		return
	}

	if err := participant.CreateFromInvite(tx, invite, user); err != nil {
		log.WithContext(c).Errorf("error creating a MeetingParticipant: %s", err)
	}
}

// Deals with the situation when a user logins as a response to an invite
func dealWithInviteFromCallback(c buffalo.Context, inviteType, objectUUID string, user models.User) {
	switch inviteType {
	case InviteTypeMeetingParam:
		ensureMeetingParticipant(c, objectUUID, user)
	default:
		log.WithContext(c).Error("incorrect meeting invite type in session: " + inviteType)
	}
}

func getInviteInfoFromSession(c buffalo.Context) (string, string) {
	inviteType, ok := c.Session().Get(InviteTypeSessionKey).(string)
	if !ok {
		return "", ""
	}

	objectUUID, ok := c.Session().Get(InviteObjectUUIDSessionKey).(string)
	if !ok {
		log.WithContext(c).Error("got meeting invite type from session but not its UUID")
		return "", ""
	}
	return inviteType, objectUUID
}

func orgBasedAuthCallback(c buffalo.Context, orgUUID, authEmail, clientID string) error {
	org := models.Organization{}
	tx := models.Tx(c)
	err := org.FindByUUID(tx, orgUUID)
	if err != nil {
		return logErrorAndRedirect(c, api.ErrorFindingOrgByID,
			fmt.Sprintf("error finding org with UUID %s ... %v", orgUUID, err.Error()))
	}

	domain.NewExtra(c, "authEmail", authEmail)

	ap, err := org.GetAuthProvider(tx, authEmail)
	if err != nil {
		return logErrorAndRedirect(c, api.ErrorLoadingAuthProvider,
			fmt.Sprintf("error loading auth provider for '%s' ... %v", org.Name, err))
	}

	authResp := ap.AuthCallback(c)
	if authResp.Error != nil {
		return logErrorAndRedirect(c, api.ErrorAuthProvidersCallback, authResp.Error.Error())
	}

	returnTo := getOrSetReturnTo(c)

	if authResp.AuthUser == nil {
		return logErrorAndRedirect(c, api.ErrorAuthProvidersCallback, "nil authResp.AuthUser")
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User

	if err := verifyEmails(c, authEmail, authResp.AuthUser.Email); err != nil {
		c.Session().Clear()
		domain.NewExtra(c, "authEmail", authEmail)
		appError := api.NewAppError(err, api.ErrorAuthEmailMismatch, api.CategoryUser)
		appError.HttpStatus = 302 // Get this redirected to the UI to display an error message
		return reportError(c, appError)
	}

	// Check for an invite in the Session
	inviteType, objectUUID := getInviteInfoFromSession(c)

	// login was success, clear session so new login can be initiated if needed
	c.Session().Clear()

	if err := user.FindOrCreateFromAuthUser(tx, org.ID, authResp.AuthUser); err != nil {
		return logErrorAndRedirect(c, api.ErrorWithAuthUser, err.Error())
	}

	if inviteType != "" {
		dealWithInviteFromCallback(c, inviteType, objectUUID, user)
	}

	authUser, err := newOrgBasedAuthUser(c, clientID, user, org)
	if err != nil {
		return err
	}

	log.SetUser(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, returnTo))
}

// authCallback assumes the user has logged in to the IDP or Oauth service and now their browser
// has been redirected back with the final response
func authCallback(c buffalo.Context) error {
	clientID, ok := c.Session().Get(ClientIDSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, api.ErrorMissingSessionClientID,
			ClientIDSessionKey+" session entry is required to complete login")
	}

	authEmail, ok := c.Session().Get(AuthEmailSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, api.ErrorMissingSessionAuthEmail,
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
		log.WithContext(c).Warning(msg)
		return nil
	}

	return errors.New("authentication email domains don't match: " + originalAuthEmail +
		" vs. " + authRespEmail)
}

// Make extras variadic, so that it can be omitted from the params
func authRequestError(c buffalo.Context, authErr authError, extras ...map[string]interface{}) error {
	log.WithContext(c).Error(authErr.errorMsg)

	appErr := api.AppError{
		Code: authErr.httpStatus,
		Key:  authErr.errorKey,
	}

	c.Session().Clear()

	return c.Render(authErr.httpStatus, render.JSON(appErr))
}

// Make extras variadic, so that it can be omitted from the params
func logErrorAndRedirect(c buffalo.Context, code api.ErrorKey, message string) error {
	log.WithContext(c).Error(message)

	c.Session().Clear()

	uiUrl := domain.Env.UIURL + "/login"
	return c.Redirect(http.StatusFound, uiUrl)
}

// authDestroy uses the bearer token to find the user's access token and
// calls the appropriate provider's logout function.
func authDestroy(c buffalo.Context) error {
	tokenParam := c.Param(LogoutToken)
	if tokenParam == "" {
		return logErrorAndRedirect(c, api.ErrorMissingLogoutToken,
			LogoutToken+" is required to logout")
	}

	var uat models.UserAccessToken
	tx := models.Tx(c)
	err := uat.FindByBearerToken(tx, tokenParam)
	if err != nil {
		return logErrorAndRedirect(c, api.ErrorFindingAccessToken, err.Error())
	}

	org, err := uat.GetOrganization(tx)
	if err != nil {
		return logErrorAndRedirect(c, api.ErrorFindingOrgForAccessToken, err.Error())
	}

	authUser, err := uat.GetUser(tx)
	if err != nil {
		return logErrorAndRedirect(c, api.ErrorAuthProvidersLogout, err.Error())
	}

	log.SetUser(c, authUser.UUID.String(), authUser.Nickname, authUser.Email)

	authPro, err := org.GetAuthProvider(tx, authUser.Email)
	if err != nil {
		return logErrorAndRedirect(c, api.ErrorLoadingAuthProvider, err.Error())
	}

	authResp := authPro.Logout(c)
	if authResp.Error != nil {
		return logErrorAndRedirect(c, api.ErrorAuthProvidersLogout, authResp.Error.Error())
	}

	redirectURL := domain.Env.UIURL

	if authResp.RedirectURL != "" {
		var uat models.UserAccessToken
		err = uat.DeleteByBearerToken(tx, tokenParam)
		if err != nil {
			return logErrorAndRedirect(c, api.ErrorDeletingAccessToken, err.Error())
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
		err := userAccessToken.FindByBearerToken(models.DB, bearerToken)
		if err != nil {
			if domain.IsOtherThanNoRows(err) {
				log.WithContext(c).Error(err)
			}
			return c.Error(http.StatusUnauthorized, errors.New("invalid bearer token"))
		}

		isExpired, err := userAccessToken.DeleteIfExpired(models.DB)
		if err != nil {
			log.WithContext(c).Error(err)
		}

		if isExpired {
			return c.Error(http.StatusUnauthorized, errors.New("expired bearer token"))
		}

		user, err := userAccessToken.GetUser(models.DB)
		if err != nil {
			return c.Error(http.StatusInternalServerError, fmt.Errorf("error finding user by access token, %s", err.Error()))
		}
		c.Set(domain.ContextKeyCurrentUser, user)

		log.SetUser(c, user.UUID.String(), user.Nickname, user.Email)
		msg := fmt.Sprintf("user %s authenticated with bearer token from ip %s", user.Email, c.Request().RemoteAddr)
		domain.NewExtra(c, "user_id", user.ID)
		domain.NewExtra(c, "email", user.Email)
		domain.NewExtra(c, "ip", c.Request().RemoteAddr)
		log.WithContext(c).Info(msg)

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
		uiURL += "/welcome"
		if len(returnTo) > 0 {
			params += "&" + ReturnToParam + "=" + url.QueryEscape(returnTo)
		}
		return uiURL + params
	}

	// Avoid two question marks in the params
	if strings.Contains(returnTo, "?") && strings.HasPrefix(params, "?") {
		params = "&" + params[1:]
	}

	return uiURL + returnTo + params
}

// getClientIPAddress gets the client IP address from CF-Connecting-IP or RemoteAddr
func getClientIPAddress(c buffalo.Context) (net.IP, error) {
	req := c.Request()

	// https://developers.cloudflare.com/fundamentals/get-started/reference/http-request-headers/#cf-connecting-ip
	if cf := req.Header.Get("CF-Connecting-IP"); cf != "" {
		return net.ParseIP(cf), nil
	}

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port, %w", req.RemoteAddr, err)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not a valid IP address, %w", req.RemoteAddr, err)
	}

	return userIP, nil
}
