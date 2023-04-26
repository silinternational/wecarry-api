// Package google implements the OAuth2 protocol for authenticating users
// through Google.
package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"golang.org/x/oauth2"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
)

const endpointProfile string = "https://www.googleapis.com/oauth2/v2/userinfo"

const ProviderName = "google"

// New creates a new Google provider, and sets up important connection details.
// You should always call `google.New` to get a new Provider. Never try to create
// one manually.
func New(config struct{ Key, Secret string }, jsonConfig json.RawMessage) (*Provider, error) {
	// If jsonConfig is provided, use it. Otherwise, use the SocialAuthConfig
	if len(jsonConfig) > 10 { // just some small number to see if it probably has valid data
		if err := json.Unmarshal(jsonConfig, &config); err != nil {
			err = errors.New("error unmarshaling google provider config json, " + err.Error())
			return &Provider{}, err
		}
	}

	if config.Key == "" || config.Secret == "" {
		err := errors.New("missing required config value for Google Auth Provider")
		return &Provider{}, err
	}

	scopes := []string{"profile", "email"}

	p := &Provider{
		ClientKey:    config.Key,
		Secret:       config.Secret,
		CallbackURL:  domain.AuthCallbackURL,
		providerName: ProviderName,
	}
	p.config = newConfig(p, scopes)
	return p, nil
}

// Provider is the implementation of `goth.Provider` for accessing Google.
type Provider struct {
	ClientKey    string
	Secret       string
	CallbackURL  string
	HTTPClient   *http.Client
	config       *oauth2.Config
	prompt       oauth2.AuthCodeOption
	providerName string
}

// Logout calls auth.Logout
func (p *Provider) Logout(c buffalo.Context) auth.Response {
	resp := auth.Response{}
	err := auth.Logout(c.Response(), c.Request())
	if err != nil {
		resp.Error = err
	}
	return resp
}

// AuthCallback deals with the session and the provider to access basic information about the user.
func (p *Provider) AuthCallback(c buffalo.Context) auth.Response {
	res := c.Response()
	req := c.Request()

	resp := auth.Response{}

	defer auth.Logout(res, req)

	msg := auth.CheckSessionStore()
	if msg != "" {
		log.Errorf("got message from Google's CheckSessionStore() in AuthCallback ... %s", msg)
	}

	value, err := auth.GetFromSession(ProviderName, req)
	if err != nil {
		resp.Error = err
		return resp
	}

	sess, err := p.UnmarshalSession(value)
	if err != nil {
		resp.Error = err
		return resp
	}

	err = auth.ValidateState(req, sess)
	if err != nil {
		resp.Error = err
		return resp
	}

	user, err := p.FetchUser(sess)
	if err == nil {
		authUser := auth.User{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			UserID:    user.UserID,
			Nickname:  user.NickName,
		}

		resp.AuthUser = &authUser

		// user can be found with existing session data
		return resp
	}

	// get new token and retry fetch
	_, err = sess.Authorize(p, req.URL.Query())
	if err != nil {
		resp.Error = err
		return resp
	}

	err = auth.StoreInSession(ProviderName, sess.Marshal(), req, res)

	if err != nil {
		resp.Error = err
		return resp
	}

	var gu goth.User
	if gu, err = p.FetchUser(sess); err != nil {
		resp.Error = err
		return resp
	}

	authUser := auth.User{
		FirstName: gu.FirstName,
		LastName:  gu.LastName,
		Email:     gu.Email,
		UserID:    gu.UserID,
		Nickname:  gu.NickName,
		PhotoURL:  gu.AvatarURL,
	}

	resp.AuthUser = &authUser
	return resp
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return p.providerName
}

// SetName is to update the name of the provider (needed in case of multiple providers of 1 type)
func (p *Provider) SetName(name string) {
	p.providerName = name
}

// Client returns an HTTP client to be used in all fetch operations.
func (p *Provider) Client() *http.Client {
	return goth.HTTPClientWithFallBack(p.HTTPClient)
}

// Debug is a no-op for the google package.
func (p *Provider) Debug(debug bool) {}

// AuthRequest calls BeginAuth and returns the URL for the authentication end-point
func (p *Provider) AuthRequest(c buffalo.Context) (string, error) {
	req := c.Request()

	sess, err := p.BeginAuth(auth.SetState(req))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	err = auth.StoreInSession(ProviderName, sess.Marshal(), req, c.Response())

	if err != nil {
		return "", err
	}

	return url, err
}

// BeginAuth asks Google for an authentication endpoint.
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	var opts []oauth2.AuthCodeOption
	if p.prompt != nil {
		opts = append(opts, p.prompt)
	}
	url := p.config.AuthCodeURL(state, opts...)
	session := &Session{
		AuthURL: url,
	}
	return session, nil
}

type googleUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Link      string `json:"link"`
	Picture   string `json:"picture"`
}

// FetchUser will go to Google and access basic information about the user.
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	sess := session.(*Session)
	user := goth.User{
		AccessToken:  sess.AccessToken,
		Provider:     p.Name(),
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.ExpiresAt,
	}

	if user.AccessToken == "" {
		// Data is not yet retrieved, since accessToken is still empty.
		return user, fmt.Errorf("%s cannot get user information without accessToken", p.providerName)
	}

	response, err := p.Client().Get(endpointProfile + "?access_token=" + url.QueryEscape(sess.AccessToken))
	if err != nil {
		return user, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			panic("error closing google auth provider response body: " + err.Error())
		}
	}()

	if response.StatusCode != http.StatusOK {
		return user, fmt.Errorf("%s responded with a %d trying to fetch user information", p.providerName, response.StatusCode)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	var u googleUser
	if err := json.Unmarshal(responseBytes, &u); err != nil {
		return user, err
	}

	// Extract the user data we got from Google into our goth.User.
	user.Name = u.Name
	user.FirstName = u.FirstName
	user.LastName = u.LastName
	user.NickName = u.Name
	user.Email = u.Email
	user.AvatarURL = u.Picture
	user.UserID = u.ID
	// Google provides other useful fields such as 'hd'; get them from RawData
	if err := json.Unmarshal(responseBytes, &user.RawData); err != nil {
		return user, err
	}

	return user, nil
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Endpoint:     endpoint,
		Scopes:       []string{},
	}

	if len(scopes) > 0 {
		for _, scope := range scopes {
			c.Scopes = append(c.Scopes, scope)
		}
	} else {
		c.Scopes = []string{"email"}
	}
	return c
}

// RefreshTokenAvailable refresh token is provided by auth provider or not
func (p *Provider) RefreshTokenAvailable() bool {
	return true
}

// RefreshToken get new access token based on the refresh token
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	ts := p.config.TokenSource(goth.ContextForClient(p.Client()), token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, err
}

// SetPrompt sets the prompt values for the google OAuth call. Use this to
// force users to choose and account every time by passing "select_account",
// for example.
// See https://developers.google.com/identity/protocols/OpenIDConnect#authenticationuriparameters
func (p *Provider) SetPrompt(prompt ...string) {
	if len(prompt) == 0 {
		return
	}
	p.prompt = oauth2.SetAuthURLParam("prompt", strings.Join(prompt, " "))
}

// UnmarshalSession will unmarshal a JSON string into a session.
func (p *Provider) UnmarshalSession(data string) (goth.Session, error) {
	sess := &Session{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(sess)
	return sess, err
}
