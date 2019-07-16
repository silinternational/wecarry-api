package actions

import (
	"fmt"
	"os"

	"strings"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
//	"github.com/silinternational/handcarry-api/google"
	saml2provider "github.com/silinternational/handcarry-api/authproviders/saml2/saml2provider"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/nulls"
	"github.com/markbates/going/defaults"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/pkg/errors"
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

	goth.UseProviders(googleP, &saml2P)

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
		return "", fmt.Errorf("Could not validate %s. %v", SAMLResponseKey, err.Error())
	}

	if response == nil {
		return "", fmt.Errorf("Got nil response validating %s.", SAMLResponseKey)
	}

	assertions := response.Assertions
	if len(assertions) < 1 {
		return "", fmt.Errorf("Did not get any SAML assertions.")
	}

	providerID := saml2provider.GetSAMLAttributeFirstValue(SAMLUserIDKey, assertions[0].AttributeStatement.Attributes)
	if providerID == "" {
		return "", fmt.Errorf("No value found for %s.", SAMLUserIDKey)
	}

	return providerID, nil
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
	u.Name = defaults.String(defaults.String(gothicUser.Name, u.Name), "MISSING")
	u.Provider = nulls.NewString(gothicUser.Provider)
	u.ProviderID = nulls.NewString(providerID)
	u.Email = nulls.NewString(defaults.String(gothicUser.Email, u.Email.String))

	// Create a partial access token,
	// concatenate it to the client_id, hash it,
	// save the hash to User row in database and
	// return the parial access token
	accessTokenPart := models.CreateAccessTokenPart()
	hashedAccessToken := models.GetHashOfClientIDPlusAccessToken(clientID + accessTokenPart)
	u.AccessToken = nulls.NewString(hashedAccessToken)
	u.AccessTokenExpiration = nulls.NewString(models.CreateAccessTokenExpiry())
	u.AuthType = nulls.NewString(gothicUser.Provider)

	if err = tx.Save(u); err != nil {
		return errors.WithStack(err)
	}

	c.Flash().Add("success", "You have been logged in")

	fmt.Printf("\nAuthn User accessTokenPart: %s\n", accessTokenPart)

	// TODO Figure out how to redirect to the ReturnTo url

	returnTo := relayValues["ReturnTo"]
	if returnTo == "" {
		returnTo = "/"
	}

	returnTo += "?ClientID=" + clientID

	return c.Redirect(302, returnTo)
}

func AuthDestroy(c buffalo.Context) error {

	bearerToken := domain.GetBearerTokenFromRequest(c.Request())
	if bearerToken == "" {
		return errors.WithStack(fmt.Errorf("No Bearer token provided."))
	}

	user, err := models.FindUserWithClientIDPlusAccessToken(bearerToken)
	if err != nil {
		return errors.WithStack(err)
	}

	var logoutURL string
	tx := c.Value("tx").(*pop.Connection)
	if err != nil {
		return errors.WithStack(err)
	}

	if user.Provider.String == SAML2Provider {
		logoutURL = os.Getenv("SAML2_LOGOUT_URL")
		user.SetAccessToken("", SAML2Provider)
		if err := tx.Save(&user); err != nil {
			return errors.WithStack(err)
		}
	}

	if logoutURL == "" {
		c.Flash().Add("warning", "To log out completely, close your browser")
		logoutURL = "/"
	}

	c.Session().Clear()
	return c.Redirect(302, logoutURL)
}

func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			if err := tx.Find(u, uid); err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
		}
		return next(c)
	}
}

func OldAuthorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid == nil {
			c.Flash().Add("danger", "You must be authorized to see that page")
			return c.Redirect(302, "/")
		}
		return next(c)
	}
}

// Authorize expects there to be a "Bearer" header in the request.
// It hashes it and tries to find a User with a matching access_token
// that has an access_token_expiry value in the future.
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		bearerToken := domain.GetBearerTokenFromRequest(c.Request())

		originalURL := strings.Split(c.Request().URL.String(), "?")[0]

		clientID, err := domain.GetRequestParam("ClientID", c.Request())
		if err != nil {
			c.Flash().Add("danger", "Error finding 'ClientID' in request params")
			return c.Redirect(302, "/")
		}

		if clientID == "" {
			c.Flash().Add("danger", "Missing 'ClientID' in request params")
			return c.Redirect(302, "/")
		}

		discoURL := fmt.Sprintf("%s?state=ReturnTo=%s-ClientID=%s", AuthDiscoURL, originalURL, clientID)

		if bearerToken == "" {
			// TODO: Add logging
			c.Flash().Add("danger", "You must be authenticated to see that page")
			return c.Redirect(302, discoURL)
		}

		_, err = models.FindUserWithClientIDPlusAccessToken(bearerToken)

		if err != nil {
			fmt.Printf("\nEEEEEE %v\n", err)
			// TODO: Add logging
			c.Flash().Add("danger", "You must be authenticated properly to see that page")
			return c.Redirect(302, discoURL)
		}

		return next(c)
	}
}

func AuthDiscoHandler(c buffalo.Context) error {
	state, err := domain.GetRequestParam("state", c.Request())
	if err != nil {
		c.Flash().Add("danger", "Error finding 'state' in request params")
		return c.Redirect(302, "/")
	}

	c.Set("state", state)

	return c.Render(200, r.HTML("auth_disco.html"))
}
