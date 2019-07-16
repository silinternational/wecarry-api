package actions

import (
	"fmt"
	"os"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	// 	"github.com/silinternational/handcarry-api/google"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/pkg/errors"
	saml2provider "github.com/silinternational/handcarry-api/authproviders/saml2"
)

const SessionNameUserID = "current_user_id"
const SAML2Provider = "saml2"
const SAMLResponseKey = "SAMLResponse"
const SAMLUserIDKey = "UserID"

const AuthDiscoURL = "/auth/disco"

func init() {
	gothic.Store = App().SessionStore

	//googleP := google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:3000/auth/google/callback")

	saml2P := saml2provider.Provider{CallbackURL: saml2provider.CallbackURL}
	saml2P.SetName(SAML2Provider)
	saml2P.IDPURL = os.Getenv("SAML2_SINGLE_SIGN_ON_URL")

	goth.UseProviders(&saml2P)

	//goth.UseProviders(&saml2P)
	//github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/github/callback")),
	// cloudfoundry.New(os.Getenv("UAA_URL"), os.Getenv("UAA_CLIENT"),
	// 	os.Getenv("UAA_CLIENT_SECRET"),
	// 	fmt.Sprintf("%s%s", App().Host, "/auth/cloudfoundry/callback"),
	// 	"openid",
	// ),
}

func getClientIDAndSAMLResponse(relayState string, c buffalo.Context) (string, string, error) {
	relayValues := domain.GetSubPartKeyValues(relayState, "?", "=")

	clientID := relayValues["ClientID"]
	fmt.Printf("\nIIIIIIIIII ClientID %v <<  relayState: %v\n", clientID, relayState)

	reqData, err := domain.GetRequestData(c.Request())
	if err != nil {
		return "", "", err
	}

	samlResponse := domain.GetFirstStringFromSlice(reqData[SAMLResponseKey])
	if samlResponse == "" {
		return "", "", fmt.Errorf("%s not found in request", SAMLResponseKey)
	}

	return clientID, samlResponse, nil
}

func getProviderIDFromSAMLResponse(samlResponse string) (string, error) {
	sprov := saml2provider.Provider{}
	sp, err := sprov.GetSAMLServiceProvider()
	if err != nil {
		return "", err
	}

	response, err := sp.ValidateEncodedResponse(samlResponse)

	if err != nil {
		return "", fmt.Errorf("could not validate %s. %v", SAMLResponseKey, err.Error())
	}

	if response == nil {
		return "", fmt.Errorf("got nil response validating %s", SAMLResponseKey)
	}

	assertions := response.Assertions
	if len(assertions) < 1 {
		return "", fmt.Errorf("did not get any SAML assertions")
	}

	providerID := saml2provider.GetSAMLAttributeFirstValue(SAMLUserIDKey, assertions[0].AttributeStatement.Attributes)
	if providerID == "" {
		return "", fmt.Errorf("no value found for %s", SAMLUserIDKey)
	}

	return providerID, nil
}

func AuthLogin(c buffalo.Context) error {

	returnTo := c.Param("ReturnTo")
	if returnTo == "" {
		returnTo = "/"
	}

	clientID := c.Param("client_id")
	if clientID == "" {
		return fmt.Errorf("client_id is required to login")
	}

	state := fmt.Sprintf("ReturnTo=%s-ClientID=%s", returnTo, clientID)

	sprov := saml2provider.Provider{}
	session, err := sprov.BeginAuth(state)
	if err != nil {
		return err
	}

	authURL, err := session.GetAuthURL()
	if err != nil {
		return err
	}

	return c.Redirect(302, authURL)
}

func AuthCallback(c buffalo.Context) error {
	println("\n With buffalo dev, use localhost:3000?ClientID=123456.  Otherwise the session won't be recognized. \n")

	relayState, err := domain.GetRequestParam("RelayState", c.Request())
	if err != nil {
		return errors.WithStack(err)
	}

	relayValues := domain.GetSubPartKeyValues(relayState, "?", "=")

	clientID, samlResponse, err := getClientIDAndSAMLResponse(relayState, c)
	if err != nil {
		return c.Error(401, err)
	}

	providerID, err := getProviderIDFromSAMLResponse(samlResponse)
	if err != nil {
		return c.Error(401, err)
	}

	gothicUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}
	fmt.Printf("\nPPPPP %v\n", providerID)

	// Get an existing User with the current provider and provider userID
	tx := c.Value("tx").(*pop.Connection)
	q := tx.Where("provider = ? and provider_id = ?", gothicUser.Provider, providerID)
	exists, err := q.Exists("users")
	if err != nil {
		return errors.WithStack(err)
	}
	u := &models.User{}
	if exists {
		if err = q.First(u); err != nil {
			return errors.WithStack(err)
		}
	}

	// Create or update the User with latest data from the provider
	// u.FirstName = defaults.String(defaults.String(gothicUser.Name, u.FirstName), "MISSING")
	// u.AuthOrgID = nulls.NewString(gothicUser.Provider)
	// u.ProviderID = nulls.NewString(providerID)
	// u.Email = nulls.NewString(defaults.String(gothicUser.Email, u.Email.String))

	// Create a partial access token,
	// concatenate it to the client_id, hash it,
	// save the hash to User row in database and
	// return the parial access token
	// accessTokenPart := models.CreateAccessTokenPart()
	// hashedAccessToken := models.GetHashOfClientIDPlusAccessToken(clientID + accessTokenPart)
	// u.AccessToken = nulls.NewString(hashedAccessToken)
	// u.AccessTokenExpiration = nulls.NewString(models.CreateAccessTokenExpiry())
	// u.AuthType = nulls.NewString(gothicUser.Provider)

	accessToken, expiresAt, err := u.CreateAccessToken(tx, clientID)
	if err != nil {
		return errors.WithStack(err)
	}

	if err = tx.Save(u); err != nil {
		return errors.WithStack(err)
	}

	returnTo := relayValues["ReturnTo"]
	if returnTo == "" {
		returnTo = "/"
	}

	returnTo += fmt.Sprintf("?access_token=%s&expires=%v", accessToken, expiresAt)

	return c.Redirect(302, returnTo)
}

func AuthDestroy(c buffalo.Context) error {

	bearerToken := domain.GetBearerTokenFromRequest(c.Request())
	if bearerToken == "" {
		return errors.WithStack(fmt.Errorf("no Bearer token provided"))
	}

	user, err := models.FindUserByAccessToken(bearerToken)
	if err != nil {
		return errors.WithStack(err)
	}

	var logoutURL string

	if user.AuthOrg.AuthType == SAML2Provider {
		// TODO get logout url from user.AuthOrg.AuthConfig
		logoutURL = os.Getenv("SAML2_LOGOUT_URL")
		err := models.DeleteAccessToken(bearerToken)
		if err != nil {
			return err
		}
	}

	if logoutURL == "" {
		logoutURL = "/"
	}

	c.Session().Clear()
	return c.Redirect(302, logoutURL)
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
