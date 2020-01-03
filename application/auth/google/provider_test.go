package google_test

import (
	"fmt"
	"testing"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	provider := googleProvider()
	a.Equal(provider.ClientKey, domain.Env.GoogleKey)
	a.Equal(provider.Secret, domain.Env.GoogleSecret)
	a.Equal(provider.CallbackURL, "/foo")
}

func Test_BeginAuth(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	provider := googleProvider()
	session, err := provider.BeginAuth("test_state")
	s := session.(*google.Session)
	a.NoError(err)
	a.Contains(s.AuthURL, "accounts.google.com/o/oauth2/auth")
	a.Contains(s.AuthURL, fmt.Sprintf("client_id=%s", domain.Env.GoogleKey))
	a.Contains(s.AuthURL, "state=test_state")
	a.Contains(s.AuthURL, "scope=email")
}

func Test_BeginAuthWithPrompt(t *testing.T) {
	// This exists because there was a panic caused by the oauth2 package when
	// the AuthCodeOption passed was nil. This test uses it, Test_BeginAuth does
	// not, to ensure both cases are covered.
	t.Parallel()
	a := assert.New(t)

	provider := googleProvider()
	provider.SetPrompt("test", "prompts")
	session, err := provider.BeginAuth("test_state")
	s := session.(*google.Session)
	a.NoError(err)
	a.Contains(s.AuthURL, "accounts.google.com/o/oauth2/auth")
	a.Contains(s.AuthURL, fmt.Sprintf("client_id=%s", domain.Env.GoogleKey))
	a.Contains(s.AuthURL, "state=test_state")
	a.Contains(s.AuthURL, "scope=email")
	a.Contains(s.AuthURL, "prompt=test+prompts")
}

func Test_Implements_Provider(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	a.Implements((*goth.Provider)(nil), googleProvider())
}

func Test_SessionFromJSON(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	provider := googleProvider()

	s, err := provider.UnmarshalSession(`{"AuthURL":"https://accounts.google.com/o/oauth2/auth","AccessToken":"1234567890"}`)
	a.NoError(err)
	session := s.(*google.Session)
	a.Equal(session.AuthURL, "https://accounts.google.com/o/oauth2/auth")
	a.Equal(session.AccessToken, "1234567890")
}

func googleProvider() *google.Provider {
	return google.New(domain.Env.GoogleKey, domain.Env.GoogleSecret, "/foo")
}
