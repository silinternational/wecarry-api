package actions

import (
	"errors"
	"fmt"
	"net/http"
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

	// http param for expires utc
	ExpiresUTCParam = "expires-utc"

	// http params for the Invite type and code
	InviteCode        = "code"
	InviteType        = "invite-type"
	InviteTypeMeeting = "meeting"

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

	// http param for token type
	TokenTypeParam = "token-type"
)

type AuthOrgOption struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	LogoURL string `json:"LogoURL"`
}

type authInviteResponse struct {
	Type     string `json:"type"`
	Name     string `json:"name"` //  TODO ensure this is XSS safe (i.e. that no javascript gets through - sanitized)
	ImageURL string `json:"imageURL"`
}

type AuthUser struct {
	ID                   string          `json:"ID"`
	Name                 string          `json:"Name"`
	Nickname             string          `json:"Nickname"`
	Email                string          `json:"Email"`
	Organizations        []AuthOrgOption `json:"Organizations"`
	AccessToken          string          `json:"AccessToken"`
	AccessTokenExpiresAt int64           `json:"AccessTokenExpiresAt"`
	IsNew                bool            `json:"IsNew"`
}

type AuthResponse struct {
	Error          *domain.AppError `json:"Error,omitempty"`
	AuthOrgOptions *[]AuthOrgOption `json:"AuthOrgOptions,omitempty"`
	RedirectURL    string           `json:"RedirectURL,omitempty"`
	User           *AuthUser        `json:"User,omitempty"`
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

func getAuthInviteResponse(c buffalo.Context) (authInviteResponse, error) {
	inviteCode := c.Param(InviteCode)

	if inviteCode == "" {
		return authInviteResponse{}, authRequestError(c, http.StatusBadRequest, domain.ErrorInvalidInviteCode,
			"missing Invite Code.")
	}

	extras := map[string]interface{}{"authInviteCode": inviteCode}

	var meeting models.Meeting
	if err := meeting.FindByInviteCode(inviteCode); err != nil {
		return authInviteResponse{}, authRequestError(c, http.StatusNotFound, domain.ErrorInvalidInviteCode,
			"error validating Invite Code: "+err.Error(), extras)
	}

	c.Session().Set(InviteTypeSessionKey, InviteTypeMeeting)
	c.Session().Set(InviteObjectUUIDSessionKey, meeting.UUID.String())

	resp := authInviteResponse{
		Type: InviteTypeMeeting,
		Name: meeting.Name,
	}
	if meeting.ImageFileID.Valid {
		resp.ImageURL = meeting.ImageFile.URL
	}
	return resp, nil
}

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

func getOrgAndUserOrgs(
	authEmail string,
	c buffalo.Context) (models.Organization, models.UserOrganizations, error) {
	var orgID int
	oid := c.Param(OrgIDParam)
	if oid == "" {
		orgID = 0
	} else {
		orgID, _ = strconv.Atoi(oid)
	}

	var org models.Organization
	var userOrgs models.UserOrganizations
	if err := userOrgs.FindByAuthEmail(authEmail, orgID); err != nil {
		return org, userOrgs, err
	}

	if len(userOrgs) == 1 {
		org = userOrgs[0].Organization
	}

	// no user_organization records yet, see if we have an organization for user's email domain
	if len(userOrgs) == 0 {
		if err := org.FindByDomain(domain.EmailDomain(authEmail)); err != nil {
			return org, userOrgs, err
		}
		if org.AuthType == "" {
			return org, userOrgs, errors.New("unable to find organization by email domain")
		}
	}

	return org, userOrgs, nil
}

func provideOrgOptions(userOrgs models.UserOrganizations, c buffalo.Context) error {
	var orgOpts []AuthOrgOption
	for _, uo := range userOrgs {
		orgOpts = append(orgOpts, AuthOrgOption{
			ID:      strconv.Itoa(uo.ID),
			Name:    uo.Organization.Name,
			LogoURL: "",
		})
	}

	resp := AuthResponse{
		AuthOrgOptions: &orgOpts,
	}
	return c.Render(http.StatusOK, render.JSON(resp))
}

func finishInviteAuthRequest(c buffalo.Context, authEmail string, extras map[string]interface{}) error {
	return authRequestError(c, http.StatusInternalServerError, "ErrorNotImplemented",
		"Not ready to do non-Org invite authentication", extras)
}

func meetingAuthRequest(c buffalo.Context, authEmail string, extras map[string]interface{}) error {
	meetingUUID, ok := c.Session().Get(InviteObjectUUIDSessionKey).(string)
	if !ok {
		return authRequestError(c, http.StatusBadRequest, domain.ErrorMissingSessionInviteObjectUUID,
			InviteObjectUUIDSessionKey+" session entry is required to complete login for an invite",
			extras)
	}

	var meeting models.Meeting
	if err := meeting.FindByUUID(meetingUUID); err != nil {
		return authRequestError(c, http.StatusNotFound, domain.ErrorInvalidSessionInviteObjectUUID,
			InviteObjectUUIDSessionKey+" session entry caused an error finding the related meeting: "+err.Error(),
			extras)
	}

	// find org for auth config and processing
	org, userOrgs, err := getOrgAndUserOrgs(authEmail, c)
	if err != nil {
		if domain.IsOtherThanNoRows(err) {
			return authRequestError(c, http.StatusInternalServerError, domain.ErrorFindingOrgUserOrgs,
				fmt.Sprintf("error getting org and userOrgs ... %v", err), extras)
		}
		// Couldn't find org, so use non-org login
		return finishInviteAuthRequest(c, authEmail, extras)
	}

	return finishOrgBasedAuthRequest(c, authEmail, org, userOrgs, extras)
}

func inviteAuthRequest(c buffalo.Context, authEmail, inviteType string) error {

	extras := map[string]interface{}{"authEmail": authEmail, "inviteType": inviteType}
	if inviteType == "" {
		return authRequestError(c, http.StatusBadRequest, domain.ErrorCannotFindOrg,
			"unable to find an organization for this user", extras)
	}

	switch inviteType {
	case InviteTypeMeeting:
		return meetingAuthRequest(c, authEmail, extras)
	default:
		return authRequestError(c, http.StatusBadRequest, domain.ErrorInvalidInviteType,
			"Invite Type '"+inviteType+"' is not valid.", extras)
	}

	return nil
}

func createAuthUser(
	clientID string,
	user models.User,
	org models.Organization) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateAccessToken(org, clientID)

	if err != nil {
		return AuthUser{}, err
	}

	var uos []AuthOrgOption
	for _, uo := range user.Organizations {
		uos = append(uos, AuthOrgOption{
			ID:      uo.UUID.String(),
			Name:    uo.Name,
			LogoURL: uo.Url.String,
		})
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
		Organizations:        uos,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
		IsNew:                isNew,
	}

	return authUser, nil
}

func finishOrgBasedAuthRequest(c buffalo.Context, authEmail string,
	org models.Organization, userOrgs models.UserOrganizations,
	extras map[string]interface{}) error {

	// User has more than one organization affiliation, return list to choose from
	if len(userOrgs) > 1 {
		return provideOrgOptions(userOrgs, c)
	}

	orgID := org.UUID.String()
	c.Session().Set(OrgIDSessionKey, orgID)

	sp, err := org.GetAuthProvider(authEmail)
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("unable to load auth provider for '%s' ... %v", org.Name, err), extras)
	}

	redirectURL, err := sp.AuthRequest(c)
	if err != nil {
		return authRequestError(c, http.StatusInternalServerError, domain.ErrorGettingAuthURL,
			fmt.Sprintf("unable to determine what the authentication url should be for '%s' ... %v", org.Name, err))
	}

	resp := AuthResponse{RedirectURL: redirectURL}

	// Reply with a 200 and leave it to the UI to do the redirect
	return c.Render(http.StatusOK, render.JSON(resp))
}

func authRequest(c buffalo.Context) error {
	// Push the Client ID into the Session
	clientID := c.Param(ClientIDParam)
	if clientID == "" {
		return authRequestError(c, http.StatusBadRequest, domain.ErrorMissingClientID,
			ClientIDParam+" is required to login")
	}

	c.Session().Set(ClientIDSessionKey, clientID)

	// Get the AuthEmail param and push it into the Session
	authEmail := c.Param(AuthEmailParam)
	if authEmail == "" {
		return authRequestError(c, http.StatusBadRequest, domain.ErrorMissingAuthEmail,
			AuthEmailParam+" is required to login")
	}
	c.Session().Set(AuthEmailSessionKey, authEmail)

	getOrSetReturnTo(c)

	// Check for an invite in the Session
	inviteType, ok := c.Session().Get(InviteTypeSessionKey).(string)
	if ok {
		return inviteAuthRequest(c, authEmail, inviteType)
	}

	extras := map[string]interface{}{"authEmail": authEmail}

	// find org for auth config and processing
	org, userOrgs, err := getOrgAndUserOrgs(authEmail, c)
	if err != nil {
		if domain.IsOtherThanNoRows(err) {
			return authRequestError(c, http.StatusInternalServerError, domain.ErrorFindingOrgUserOrgs,
				fmt.Sprintf("error getting org and userOrgs ... %v", err), extras)
		}
		extras["email"] = authEmail
		return authRequestError(c, http.StatusNotFound, domain.ErrorFindingOrgUserOrgs,
			"Could not find org or userOrgs", extras)
	}

	return finishOrgBasedAuthRequest(c, authEmail, org, userOrgs, extras)
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

	if err := participant.CreateForInvite(invite, user.ID); err != nil {
		domain.Error(c, "error creating a MeetingParticipant: "+err.Error())
	}
}

func dealWithInviteFromCallback(c buffalo.Context, inviteType, objectUUID string, user models.User) {

	switch inviteType {
	case InviteTypeMeeting:
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

	orgID, ok := c.Session().Get(OrgIDSessionKey).(string)
	if !ok {
		return logErrorAndRedirect(c, domain.ErrorMissingSessionOrgID,
			OrgIDSessionKey+" session entry is required to complete login")
	}

	org := models.Organization{}
	err := org.FindByUUID(orgID)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorFindingOrgByID,
			fmt.Sprintf("error finding org with UUID %s ... %v", orgID, err.Error()))
	}

	extras := map[string]interface{}{"authEmail": authEmail}

	ap, err := org.GetAuthProvider(authEmail)
	if err != nil {
		return logErrorAndRedirect(c, domain.ErrorLoadingAuthProvider,
			fmt.Sprintf("error loading auth provider for '%s' ... %v", org.Name, err), extras)
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

	if err := verifyEmails(c, authEmail, authResp.AuthUser.Email); err != nil {
		c.Session().Clear()
		extras := map[string]interface{}{"authEmail": authEmail}
		return logErrorAndRedirect(c, domain.ErrorAuthEmailMismatch, err.Error(), extras)
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

	authUser, err := createAuthUser(clientID, user, org)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, returnTo))
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
func authRequestError(c buffalo.Context, httpStatus int, errorCode, message string, extras ...map[string]interface{}) error {
	allExtras := mergeExtras(errorCode, extras...)

	domain.Error(c, message, allExtras)

	authError := domain.AppError{
		Code: httpStatus,
		Key:  errorCode,
	}

	return c.Render(httpStatus, render.JSON(authError))
}

// Make extras variadic, so that it can be omitted from the params
func logErrorAndRedirect(c buffalo.Context, code, message string, extras ...map[string]interface{}) error {
	allExtras := mergeExtras(code, extras...)

	domain.Error(c, message, allExtras)

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
		c.Set("current_user", user)

		// set person on rollbar session
		domain.RollbarSetPerson(c, user.UUID.String(), user.Nickname, user.Email)
		msg := fmt.Sprintf("user %s authenticated with bearer token from ip %s", user.Email, c.Request().RemoteAddr)
		extras := map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      c.Request().RemoteAddr,
		}
		domain.Info(c, msg, extras)

		if err := userAccessToken.Renew(); err != nil {
			domain.Error(c, fmt.Sprintf("error renewing access token, %s", c.Request().RemoteAddr), extras)
		}

		return next(c)
	}
}

// getLoginSuccessRedirectURL generates the URL for redirection after a successful login
func getLoginSuccessRedirectURL(authUser AuthUser, returnTo string) string {
	uiURL := domain.Env.UIURL

	tokenExpiry := time.Unix(authUser.AccessTokenExpiresAt, 0).Format(time.RFC3339)
	params := fmt.Sprintf("?%s=Bearer&%s=%s&%s=%s",
		TokenTypeParam, ExpiresUTCParam, tokenExpiry, AccessTokenParam, authUser.AccessToken)

	// Ensure there is one set of /# between uiURL and the returnTo
	if !strings.HasPrefix(returnTo, `/#`) {
		returnTo = "/#" + returnTo
	}

	// New Users go straight to the welcome page
	if authUser.IsNew {
		uiURL += "/#/welcome"
		if len(returnTo) > 0 {
			params += "&" + ReturnToParam + "=" + returnTo
		}
		return uiURL + params
	}

	return uiURL + returnTo + params
}
