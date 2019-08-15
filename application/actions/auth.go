package actions

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gobuffalo/buffalo/render"

	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

const SAMLResponseKey = "SAMLResponse"
const IDPMetadataFile = "./samlmetadata/idp-metadata.xml"

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
	clientID, ok := c.Session().Get("ClientID").(string)
	if !ok {
		clientID = c.Param("client_id")
		if clientID == "" {
			return authError(c, http.StatusBadRequest, "MissingClientID", "client_id is required to login")
		}
		c.Session().Set("ClientID", clientID)
	}

	var authEmail string
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
			return authError(c, http.StatusInternalServerError, "UnableToFindOrgByEmail", "unable to find organization by email domain")
		}
		if org.AuthType == "" {
			return authError(c, http.StatusNotFound, "OrgNotFound", "unable to find organization by email domain")
		}
	}

	// get auth provider for org to process login
	sp, err := org.GetAuthProvider()
	if err != nil {
		return authError(c, http.StatusInternalServerError, "UnableToLoadAuthProvider", "unable to load auth provider for organization")
	}

	authResp := sp.Login(c)
	if authResp.Error != nil {
		return authError(c, http.StatusBadRequest, "AuthError", authResp.Error.Error())
	}

	// if redirect url is present it is initial login, not return from auth provider yet
	if authResp.RedirectURL != "" {
		resp := AuthResponse{
			RedirectURL: authResp.RedirectURL,
		}

		return c.Redirect(303, resp.RedirectURL)
	}

	// if we have an authuser, find or create user in local db and finish login
	var user models.User
	if authResp.AuthUser != nil {
		err := user.FindOrCreateFromAuthUser(org.ID, authResp.AuthUser)
		if err != nil {
			return authError(c, http.StatusBadRequest, "AuthFailure", err.Error())
		}
	}

	accessToken, expiresAt, err := user.CreateAccessToken(org.ID, clientID)
	if err != nil {
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

	authUser := &AuthUser{
		ID:                   user.Uuid.String(),
		Name:                 user.FirstName + " " + user.LastName,
		Nickname:             user.Nickname,
		Email:                user.Email,
		Organizations:        uos,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
	}

	authResponse := AuthResponse{
		User: authUser,
	}

	return c.Render(200, render.JSON(authResponse))
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
		return authError(c, 400, "LogoutError", "no Bearer token provided")
	}

	uat, err := models.UserAccessTokenFind(bearerToken)
	if err != nil {
		return authError(c, 500, "LogoutError", err.Error())
	}

	if uat == nil {
		return authError(c, 404, "LogoutError", "access token not found")
	}

	authPro, err := uat.UserOrganization.Organization.GetAuthProvider()
	if err != nil {
		return authError(c, 500, "LogoutError", err.Error())
	}

	authResp := authPro.Logout(c)
	if authResp.Error != nil {
		return authError(c, 500, "LogoutError", authResp.Error.Error())
	}

	var response AuthResponse

	if authResp.RedirectURL != "" {
		err = models.DeleteAccessToken(bearerToken)
		if err != nil {
			return authError(c, 500, "LogoutError", err.Error())
		}
		c.Session().Clear()
		response.RedirectURL = authResp.RedirectURL
		return c.Redirect(302, response.RedirectURL)
	}

	return c.Render(200, render.JSON(response))
}

func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		bearerToken := domain.GetBearerTokenFromRequest(c.Request())
		if bearerToken == "" {
			return errors.WithStack(fmt.Errorf("no Bearer token provided"))
		}

		user, err := models.FindUserByAccessToken(bearerToken)
		if err != nil {
			return errors.WithStack(err)
		}
		c.Set("current_user", user)

		return next(c)
	}
}
