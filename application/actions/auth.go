package actions

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gobuffalo/envy"

	"github.com/gobuffalo/buffalo/render"

	"github.com/gobuffalo/buffalo"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

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
}

type AuthResponse struct {
	Error          *AuthError       `json:"Error,omitempty"`
	AuthOrgOptions *[]AuthOrgOption `json:"AuthOrgOptions,omitempty"`
	RedirectURL    string           `json:"RedirectURL,omitempty"`
	User           *AuthUser        `json:"User,omitempty"`
}

func AuthLogin(c buffalo.Context) error {
	var clientID string

	clientID = c.Param("client_id")
	if clientID == "" {
		var ok bool
		clientID, ok = c.Session().Get("ClientID").(string)
		if !ok {
			return authError(c, http.StatusBadRequest, "MissingClientID", "client_id is required to login")
		}
	} else {
		c.Session().Set("ClientID", clientID)
	}

	var authEmail string
	var ok bool
	authEmail, ok = c.Session().Get("AuthEmail").(string)
	if !ok {
		authEmail = c.Param("authEmail")
		if authEmail == "" {
			return authError(c, http.StatusBadRequest, "MissingAuthEmail", "authEmail is required to login")
		}
		c.Session().Set("AuthEmail", authEmail)
	}

	returnTo := c.Param("ReturnTo")
	if returnTo == "" {
		returnTo = "/"
	}
	c.Session().Set("ReturnTo", returnTo)

	err := c.Session().Save()
	if err != nil {
		domain.Error(c, err.Error(), map[string]interface{}{"authEmail": authEmail})
		return authError(c, http.StatusInternalServerError, "ServerError", "unable to save session")
	}

	// find org for auth config and processing
	var orgID int
	oid := c.Param("org_id")
	if oid == "" {
		orgID = 0
	} else {
		orgID, _ = strconv.Atoi(oid)
	}

	var org models.Organization
	userOrgs, err := models.UserOrganizationFindByAuthEmail(authEmail, orgID)
	if len(userOrgs) == 1 {
		org = userOrgs[0].Organization
	}

	// no user_organization records yet, see if we have an organization for user's email domain
	if len(userOrgs) == 0 {
		org, err = models.OrganizationFindByDomain(domain.EmailDomain(authEmail))
		if err != nil {
			extras := map[string]interface{}{"authEmail": authEmail, "code": "UnableToFindOrgByEmail"}
			domain.Error(c, err.Error(), extras)
			return authError(c, http.StatusInternalServerError, "UnableToFindOrgByEmail", "unable to find organization by email domain")
		}
		if org.AuthType == "" {
			return authError(c, http.StatusNotFound, "OrgNotFound", "unable to find organization by email domain")
		}
	}

	// User has more than one organization affiliation, return list to choose from
	if len(userOrgs) > 1 {
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
	sp, err := org.GetAuthProvider()
	if err != nil {
		extras := map[string]interface{}{"authEmail": authEmail, "code": "UnableToLoadAuthProvider"}
		domain.Error(c, err.Error(), extras)
		return authError(c, http.StatusInternalServerError, "UnableToLoadAuthProvider",
			fmt.Sprintf("unable to load auth provider for '%s'", org.Name))
	}

	authResp := sp.Login(c)
	if authResp.Error != nil {
		extras := map[string]interface{}{"authEmail": authEmail, "code": "AuthError"}
		domain.Error(c, authResp.Error.Error(), extras)
		return authError(c, http.StatusBadRequest, "AuthError", authResp.Error.Error())
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

	accessToken, expiresAt, err := user.CreateAccessToken(org, clientID)
	if err != nil {
		extras := map[string]interface{}{"authEmail": authEmail, "code": "CreateAccessTokenFailure"}
		domain.Error(c, err.Error(), extras)
		return authError(c, http.StatusBadRequest, "CreateAccessTokenFailure", err.Error())
	}

	var uos []AuthOrgOption
	for _, uo := range user.Organizations {
		uos = append(uos, AuthOrgOption{
			ID:      uo.Uuid.String(),
			Name:    uo.Name,
			LogoURL: uo.Url.String,
		})
	}

	authUser := AuthUser{
		ID:                   user.Uuid.String(),
		Name:                 user.FirstName + " " + user.LastName,
		Nickname:             user.Nickname,
		Email:                user.Email,
		Organizations:        uos,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
	}

	// set person on rollbar session
	domain.RollbarSetPerson(c, authUser.ID, authUser.Nickname, authUser.Email)

	return c.Redirect(302, getLoginSuccessRedirectURL(authUser))
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

	uat, err := models.UserAccessTokenFind(bearerToken)
	if err != nil {
		domain.Error(c, err.Error(), map[string]interface{}{"code": "LogoutError"})
		return authError(c, 500, "LogoutError", err.Error())
	}

	if uat == nil {
		domain.Warn(c, "access token not found", map[string]interface{}{"code": "LogoutError"})
		return authError(c, 404, "LogoutError", "access token not found")
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
		err = models.DeleteAccessToken(bearerToken)
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

		user, err := models.FindUserByAccessToken(bearerToken)
		if err != nil {
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
func getLoginSuccessRedirectURL(authUser AuthUser) string {
	uiUrl := envy.Get("UI_URL", "/")

	tokenExpiry := time.Unix(authUser.AccessTokenExpiresAt, 0).Format(time.RFC3339)
	url := fmt.Sprintf("%s?token_type=Bearer&expires_utc=%s&access_token=%s",
		uiUrl, tokenExpiry, authUser.AccessToken)

	return url
}
