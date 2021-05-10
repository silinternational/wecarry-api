// Package facebook implements the OAuth2 protocol for authenticating users through Facebook.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package facebook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"golang.org/x/oauth2"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/domain"
)

const (
	authURL         string = "https://www.facebook.com/dialog/oauth"
	tokenURL        string = "https://graph.facebook.com/oauth/access_token"
	endpointProfile string = "https://graph.facebook.com/me?fields="
)

const ProviderName = "facebook"

// New creates a new Facebook provider, and sets up important connection details.
// You should always call `facebook.New` to get a new Provider. Never try to create
// one manually.
func New(config struct{ Key, Secret string }) (*Provider, error) {
	if config.Key == "" || config.Secret == "" {
		err := errors.New("missing required config value for Facebook Auth Provider")
		return &Provider{}, err
	}

	scopes := []string{"public_profile", "email"}

	p := &Provider{
		ClientKey:    config.Key,
		Secret:       config.Secret,
		CallbackURL:  domain.AuthCallbackURL,
		providerName: ProviderName,
	}
	p.config = newConfig(p, scopes)
	p.Fields = "email,first_name,last_name,link,about,id,name,picture,location"
	return p, nil
}

// Provider is the implementation of `goth.Provider` for accessing Facebook.
type Provider struct {
	ClientKey    string
	Secret       string
	CallbackURL  string
	HTTPClient   *http.Client
	Fields       string
	config       *oauth2.Config
	providerName string
}

// AuthCallback deals with the session and the provider to access basic information about the user.
func (p *Provider) AuthCallback(c buffalo.Context) auth.Response {
	res := c.Response()
	req := c.Request()

	resp := auth.Response{}

	defer auth.Logout(res, req)

	msg := auth.CheckSessionStore()
	if msg != "" {
		domain.Logger.Printf("got message from Facebook's CheckSessionStore() in AuthCallback ... %s", msg)
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

// SetCustomFields sets the fields used to return information
// for a user.
//
// A list of available field values can be found at
// https://developers.facebook.com/docs/graph-api/reference/user
func (p *Provider) SetCustomFields(fields []string) *Provider {
	p.Fields = strings.Join(fields, ",")
	return p
}

func (p *Provider) Client() *http.Client {
	return goth.HTTPClientWithFallBack(p.HTTPClient)
}

// Debug is a no-op for the facebook package.
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

// BeginAuth asks Facebook for an authentication end-point.
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	authUrl := p.config.AuthCodeURL(state)
	session := &Session{
		AuthURL: authUrl,
	}
	return session, nil
}

// FetchUser will go to Facebook and access basic information about the user.
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	sess := session.(*Session)
	user := goth.User{
		AccessToken: sess.AccessToken,
		Provider:    p.Name(),
		ExpiresAt:   sess.ExpiresAt,
	}

	if user.AccessToken == "" {
		// data is not yet retrieved since accessToken is still empty
		return user, fmt.Errorf("%s cannot get user information without accessToken", p.providerName)
	}

	// always add appsecretProof to make calls more protected
	// https://github.com/markbates/goth/issues/96
	// https://developers.facebook.com/docs/graph-api/securing-requests
	hash := hmac.New(sha256.New, []byte(p.Secret))
	hash.Write([]byte(sess.AccessToken))
	appsecretProof := hex.EncodeToString(hash.Sum(nil))

	reqUrl := fmt.Sprint(
		endpointProfile,
		p.Fields,
		"&access_token=",
		url.QueryEscape(sess.AccessToken),
		"&appsecret_proof=",
		appsecretProof,
	)
	response, err := p.Client().Get(reqUrl)
	if err != nil {
		return user, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return user, fmt.Errorf("%s responded with a %d trying to fetch user information", p.providerName, response.StatusCode)
	}

	bits, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	err = json.NewDecoder(bytes.NewReader(bits)).Decode(&user.RawData)
	if err != nil {
		return user, err
	}

	err = userFromReader(bytes.NewReader(bits), &user)
	return user, err
}

func userFromReader(reader io.Reader, user *goth.User) error {
	u := struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		About     string `json:"about"`
		Name      string `json:"name"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Link      string `json:"link"`
		Picture   struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
		Location struct {
			Name string `json:"name"`
		} `json:"location"`
	}{}

	err := json.NewDecoder(reader).Decode(&u)
	if err != nil {
		return err
	}

	user.Name = u.Name
	user.FirstName = u.FirstName
	user.LastName = u.LastName
	user.NickName = u.Name
	user.Email = u.Email
	user.Description = u.About
	user.AvatarURL = u.Picture.Data.URL
	user.UserID = u.ID
	user.Location = u.Location.Name

	return err
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: []string{
			"email",
		},
	}

	defaultScopes := map[string]struct{}{
		"email": {},
	}

	for _, scope := range scopes {
		if _, exists := defaultScopes[scope]; !exists {
			c.Scopes = append(c.Scopes, scope)
		}
	}

	return c
}

//RefreshToken refresh token is not provided by facebook
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return nil, errors.New("Refresh token is not provided by facebook")
}

//RefreshTokenAvailable refresh token is not provided by facebook
func (p *Provider) RefreshTokenAvailable() bool {
	return false
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
