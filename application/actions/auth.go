package actions

import (
	"fmt"
	"github.com/silinternational/handcarry-api/auth"

	//	"github.com/silinternational/handcarry-api/auth"
	"net/http"
	"strconv"
	"time"

	"github.com/gobuffalo/envy"

	"github.com/gobuffalo/buffalo/render"

	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

// http param and session key for ReturnTo
const ReturnToParam = "return-to"
const ReturnToSessionKey = "ReturnTo"

// http param and session key for Client ID
const ClientIDParam = "client-id"
const ClientIDSessionKey = "ClientID"

// http param and session key for Auth Email
const AuthEmailParam = "auth-email"
const AuthEmailSessionKey = "AuthEmail"

// http param for organization id
const OrgIDParam = "org-id"

type AuthError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type AuthOrgOption struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	LogoURL string `json:"LogoURL"`
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
	Error          *AuthError       `json:"Error,omitempty"`
	AuthOrgOptions *[]AuthOrgOption `json:"AuthOrgOptions,omitempty"`
	RedirectURL    string           `json:"RedirectURL,omitempty"`
	User           *AuthUser        `json:"User,omitempty"`
}

func getOrSetClientID(c buffalo.Context) (string, error) {
	var clientID string

	clientID = c.Param(ClientIDParam)

	if clientID == "" {
		var ok bool
		clientID, ok = c.Session().Get(ClientIDSessionKey).(string)
		if !ok {
			return "", authError(c, http.StatusBadRequest, "MissingClientID", ClientIDParam+" is required to login")
		}
	} else {
		c.Session().Set(ClientIDSessionKey, clientID)
	}

	return clientID, nil
}

func getOrSetAuthEmail(c buffalo.Context) (string, error) {
	var authEmail string
	var ok bool
	authEmail, ok = c.Session().Get(AuthEmailSessionKey).(string)
	if !ok {
		authEmail = c.Param(AuthEmailParam)
		if authEmail == "" {
			return "", authError(c, http.StatusBadRequest, "MissingAuthEmail", AuthEmailParam+" is required to login")
		}
		c.Session().Set(AuthEmailSessionKey, authEmail)
	}

	return authEmail, nil
}

func getOrSetReturnTo(c buffalo.Context) string {
	returnTo := c.Param(ReturnToParam)

	if returnTo == "" {
		var ok bool
		returnTo, ok = c.Session().Get(ReturnToSessionKey).(string)
		if !ok {
			returnTo = "/#"
		}

		return returnTo
	}

	c.Session().Set(ReturnToSessionKey, returnTo)

	return returnTo
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
	err := userOrgs.FindByAuthEmail(authEmail, orgID)
	if len(userOrgs) == 1 {
		org = userOrgs[0].Organization
	}

	// no user_organization records yet, see if we have an organization for user's email domain
	if len(userOrgs) == 0 {
		err = org.FindByDomain(domain.EmailDomain(authEmail))
		if err != nil {
			extras := map[string]interface{}{"authEmail": authEmail, "code": "UnableToFindOrgByEmail"}
			domain.Error(c, err.Error(), extras)
			return org, userOrgs, authError(c, http.StatusInternalServerError, "UnableToFindOrgByEmail", "unable to find organization by email domain")
		}
		if org.AuthType == "" {
			return org, userOrgs, authError(c, http.StatusNotFound, "OrgNotFound", "unable to find organization by email domain")
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
			LogoURL: uo.Organization.Url.String, // TODO change to a logo url when one is added to organization
		})
	}

	resp := AuthResponse{
		AuthOrgOptions: &orgOpts,
	}
	return c.Render(200, render.JSON(resp))
}

// get auth provider for org to process login
func getAuthResp(authEmail string, org models.Organization, c buffalo.Context) (auth.Response, error) {

	sp, err := org.GetAuthProvider()
	if err != nil {
		extras := map[string]interface{}{"authEmail": authEmail, "code": "UnableToLoadAuthProvider"}
		domain.Error(c, err.Error(), extras)
		return auth.Response{}, authError(c, http.StatusInternalServerError, "UnableToLoadAuthProvider",
			fmt.Sprintf("unable to load auth provider for '%s'", org.Name))
	}

	authResp := sp.Login(c)
	if authResp.Error != nil {
		extras := map[string]interface{}{"authEmail": authEmail, "code": "AuthError"}
		domain.Error(c, authResp.Error.Error(), extras)
		return auth.Response{}, authError(c, http.StatusBadRequest, "AuthError", authResp.Error.Error())
	}

	return authResp, nil
}

func createAuthUser(
	authEmail, clientID string,
	user models.User,
	org models.Organization,
	c buffalo.Context) (AuthUser, error) {
	accessToken, expiresAt, err := user.CreateAccessToken(org, clientID)

	if err != nil {
		extras := map[string]interface{}{"authEmail": authEmail, "code": "CreateAccessTokenFailure"}
		domain.Error(c, err.Error(), extras)
		return AuthUser{}, authError(c, http.StatusBadRequest, "CreateAccessTokenFailure", err.Error())
	}

	var uos []AuthOrgOption
	for _, uo := range user.Organizations {
		uos = append(uos, AuthOrgOption{
			ID:      uo.Uuid.String(),
			Name:    uo.Name,
			LogoURL: uo.Url.String,
		})
	}

	isNew := false
	if time.Since(user.CreatedAt) < time.Duration(time.Second*30) {
		isNew = true
	}

	authUser := AuthUser{
		ID:                   user.Uuid.String(),
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

func AuthLogin(c buffalo.Context) error {
	clientID, err := getOrSetClientID(c)
	if err != nil {
		return err
	}

	authEmail, err := getOrSetAuthEmail(c)
	if err != nil {
		return err
	}

	returnTo := getOrSetReturnTo(c)

	err = c.Session().Save()
	if err != nil {
		domain.Error(c, err.Error(), map[string]interface{}{"authEmail": authEmail})
		return authError(c, http.StatusInternalServerError, "ServerError", "unable to save session")
	}

	// find org for auth config and processing
	org, userOrgs, err := getOrgAndUserOrgs(authEmail, c)
	if err != nil {
		return err
	}

	// User has more than one organization affiliation, return list to choose from
	if len(userOrgs) > 1 {
		if len(userOrgs) > 1 {
			return provideOrgOptions(userOrgs, c)
		}
	}

	// get auth provider for org to process login
	authResp, err := getAuthResp(authEmail, org, c)
	if err != nil {
		return err
	}

	// if redirect url is present it is initial login, not return from auth provider yet
	if authResp.RedirectURL != "" {
		resp := AuthResponse{
			RedirectURL: authResp.RedirectURL,
		}

		return c.Render(200, render.JSON(resp))
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User
	if authResp.AuthUser != nil {
		// login was success, clear session so new login can be initiated if needed
		c.Session().Clear()
		_ = c.Session().Save()

		err := user.FindOrCreateFromAuthUser(org.ID, authResp.AuthUser)
		if err != nil {
			return authError(c, http.StatusBadRequest, "AuthFailure", err.Error())
		}
	}

	authUser, err := createAuthUser(authEmail, clientID, user, org, c)
	if err != nil {
		return err
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser, returnTo))
}

// returnAuthError takes a error code and message and renders AuthResponse to json and returns
func authError(c buffalo.Context, status int, code, message string) error {
	resp := AuthResponse{
		Error: &AuthError{
			Code:    code,
			Message: message,
		},
	}

	return c.Render(status, render.JSON(resp))
}

func AuthDestroy(c buffalo.Context) error {

	bearerToken := domain.GetBearerTokenFromRequest(c.Request())
	if bearerToken == "" {
		domain.Warn(c, "no Bearer token provided", map[string]interface{}{"code": "LogoutError"})
		return authError(c, 400, "LogoutError", "no Bearer token provided")
	}

	var uat models.UserAccessToken
	err := uat.FindByBearerToken(bearerToken)
	if err != nil {
		domain.Error(c, err.Error(), map[string]interface{}{"code": "LogoutError"})
		return authError(c, 500, "LogoutError", err.Error())
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, uat.User.Uuid.String(), uat.User.Nickname, uat.User.Email)

	authPro, err := uat.UserOrganization.Organization.GetAuthProvider()
	if err != nil {
		domain.Error(c, err.Error(), map[string]interface{}{"code": "LogoutError"})
		return authError(c, 500, "LogoutError", err.Error())
	}

	authResp := authPro.Logout(c)
	if authResp.Error != nil {
		domain.Error(c, authResp.Error.Error(), map[string]interface{}{"code": "LogoutError"})
		return authError(c, 500, "LogoutError", authResp.Error.Error())
	}

	var response AuthResponse

	if authResp.RedirectURL != "" {
		var uat models.UserAccessToken
		err = uat.DeleteByBearerToken(bearerToken)
		if err != nil {
			domain.Error(c, err.Error(), map[string]interface{}{"code": "LogoutError"})
			return authError(c, 500, "LogoutError", err.Error())
		}
		c.Session().Clear()
		response.RedirectURL = authResp.RedirectURL
	}

	return c.Render(200, render.JSON(response))
}

func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		bearerToken := domain.GetBearerTokenFromRequest(c.Request())
		if bearerToken == "" {
			return fmt.Errorf("no Bearer token provided")
		}

		var user models.User
		if err := user.FindByAccessToken(bearerToken); err != nil {
			return c.Error(401, fmt.Errorf("invalid bearer token"))
		}
		c.Set("current_user", user)

		// set person on rollbar session
		domain.RollbarSetPerson(c, user.Uuid.String(), user.Nickname, user.Email)
		msg := fmt.Sprintf("user %s authenticated with bearer token from ip %s", user.Email, c.Request().RemoteAddr)
		extras := map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
			"ip":      c.Request().RemoteAddr,
		}
		domain.Info(c, msg, extras)

		return next(c)
	}
}

// getLoginSuccessRedirectURL generates the URL for redirection after a successful login
func getLoginSuccessRedirectURL(authUser AuthUser, returnTo string) string {

	uiUrl := envy.Get("UI_URL", "") + "/#"

	tokenExpiry := time.Unix(authUser.AccessTokenExpiresAt, 0).Format(time.RFC3339)
	params := fmt.Sprintf("?token_type=Bearer&expires_utc=%s&access_token=%s",
		tokenExpiry, authUser.AccessToken)

	if authUser.IsNew {
		uiUrl += "/welcome"
		if len(returnTo) > 0 {
			params += "&" + ReturnToParam + "=" + returnTo
		}
	} else {
		if len(returnTo) > 0 && returnTo[0] != '/' {
			returnTo = "/" + returnTo
		}
		uiUrl += returnTo
	}

	url := fmt.Sprintf("%s%s", uiUrl, params)

	return url
}
