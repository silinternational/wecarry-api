package twitter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/pat"
	"github.com/markbates/goth"
	"github.com/mrjones/oauth"
	"github.com/stretchr/testify/assert"

	"github.com/silinternational/wecarry-api/domain"
)

func Test_New(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	provider, err := twitterProvider()
	a.NoError(err)
	a.Equal(provider.ClientKey, domain.Env.TwitterKey)
	a.Equal(provider.Secret, domain.Env.TwitterSecret)
	a.Equal(provider.CallbackURL, "/foo")
}

func Test_Implements_Provider(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	provider, _ := twitterProvider()

	a.Implements((*goth.Provider)(nil), provider)
}

func Test_BeginAuth(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	mockTwitter(func(ts *httptest.Server) {
		provider, err := twitterProvider()
		session, err := provider.BeginAuth("state")
		s := session.(*Session)
		a.NoError(err)
		a.Contains(s.AuthURL, "authorize?oauth_token=TOKEN")
		a.Equal("TOKEN", s.RequestToken.Token)
		a.Equal("SECRET", s.RequestToken.Secret)
	})
	mockTwitter(func(ts *httptest.Server) {
		provider := twitterProviderAuthenticate()
		session, err := provider.BeginAuth("state")
		s := session.(*Session)
		a.NoError(err)
		a.Contains(s.AuthURL, "authenticate?oauth_token=TOKEN")
		a.Equal("TOKEN", s.RequestToken.Token)
		a.Equal("SECRET", s.RequestToken.Secret)
	})
}

func Test_FetchUser(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	mockTwitter(func(ts *httptest.Server) {
		provider, _ := twitterProvider()
		session := Session{AccessToken: &oauth.AccessToken{Token: "TOKEN", Secret: "SECRET"}}

		user, err := provider.FetchUser(&session)
		a.NoError(err)

		a.Equal("Homer", user.Name)
		a.Equal("duffman", user.NickName)
		a.Equal("Duff rules!!", user.Description)
		a.Equal("http://example.com/image.jpg", user.AvatarURL)
		a.Equal("1234", user.UserID)
		a.Equal("Springfield", user.Location)
		a.Equal("TOKEN", user.AccessToken)
		a.Equal("duffman@springfield.com", user.Email)
	})
}

func Test_SessionFromJSON(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	provider, _ := twitterProvider()

	s, err := provider.UnmarshalSession(`{"AuthURL":"http://com/auth_url","AccessToken":{"Token":"1234567890","Secret":"secret!!","AdditionalData":{}},"RequestToken":{"Token":"0987654321","Secret":"!!secret"}}`)
	a.NoError(err)
	session := s.(*Session)
	a.Equal(session.AuthURL, "http://com/auth_url")
	a.Equal(session.AccessToken.Token, "1234567890")
	a.Equal(session.AccessToken.Secret, "secret!!")
	a.Equal(session.RequestToken.Token, "0987654321")
	a.Equal(session.RequestToken.Secret, "!!secret")
}

func twitterProvider() (*Provider, error) {

	domain.Env.TwitterKey = "abc123"
	domain.Env.TwitterSecret = "abc123"
	domain.Env.AuthCallbackURL = "/foo"

	return New([]byte(""))
}

func twitterProviderAuthenticate() *Provider {
	return NewAuthenticate(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "/foo")
}

func mockTwitter(f func(*httptest.Server)) {
	p := pat.New()
	p.Get("/oauth/request_token", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "oauth_token=TOKEN&oauth_token_secret=SECRET")
	})
	p.Get("/1.1/account/verify_credentials.json", func(res http.ResponseWriter, req *http.Request) {
		data := map[string]string{
			"name":              "Homer",
			"screen_name":       "duffman",
			"description":       "Duff rules!!",
			"profile_image_url": "http://example.com/image.jpg",
			"id_str":            "1234",
			"location":          "Springfield",
			"email":             "duffman@springfield.com",
		}
		json.NewEncoder(res).Encode(&data)
	})
	ts := httptest.NewServer(p)
	defer ts.Close()

	originalRequestURL := requestURL
	originalEndpointProfile := endpointProfile

	requestURL = ts.URL + "/oauth/request_token"
	endpointProfile = ts.URL + "/1.1/account/verify_credentials.json"

	f(ts)

	requestURL = originalRequestURL
	endpointProfile = originalEndpointProfile
}

func TestGetFirstLastFromName(t *testing.T) {

	type args struct {
		s []string
	}
	tests := []struct {
		name     string
		userName string
		want     [2]string
	}{
		{
			name:     "one name",
			userName: "OneName",
			want:     [2]string{"OneName", "OneName"},
		},
		{
			name:     "one space",
			userName: "First Last",
			want:     [2]string{"First", "Last"},
		},
		{
			name:     "two spaces",
			userName: "First Middle Last",
			want:     [2]string{"First", "Middle Last"},
		},
		{
			name:     "one undesrscore",
			userName: "First_Last",
			want:     [2]string{"First", "Last"},
		},
		{
			name:     "two underscores",
			userName: "First_Middle_Last",
			want:     [2]string{"First", "Middle_Last"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			gotF, gotL := getFirstLastFromName(tt.userName)

			a.Equal(tt.want, [2]string{gotF, gotL}, "incorrect names")
		})
	}
}
