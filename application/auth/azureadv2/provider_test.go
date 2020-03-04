package azureadv2_test

import (
	"testing"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/assert"

	"github.com/silinternational/wecarry-api/auth/azureadv2"
	"github.com/silinternational/wecarry-api/domain"
)

const (
	applicationID = "6731de76-14a6-49ae-97bc-6eba6914391e"
	tenantID      = "edf3cc03-7edf-4299-871a-940bc318789c"
	secret        = "foo"
)

func Test_New(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	provider := azureadProvider()

	a.Equal(provider.Name(), "azureadv2")
	a.Equal(provider.ClientKey, applicationID)
	a.Equal(provider.Secret, secret)
	a.Equal(provider.CallbackURL, domain.AuthCallbackURL)
}

func Test_Implements_Provider(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	p := azureadProvider()
	a.Implements((*goth.Provider)(nil), p)
}

func Test_BeginAuth(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	provider := azureadProvider()
	session, err := provider.BeginAuth("test_state")
	a.NoError(err)
	s := session.(*azureadv2.Session)
	a.Contains(s.AuthURL, "login.microsoftonline.com/"+tenantID+"/oauth2/v2.0/authorize")
	a.Contains(s.AuthURL, "scope=openid+profile+email")
}

func Test_SessionFromJSON(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	provider := azureadProvider()
	session, err := provider.UnmarshalSession(`{"au":"http://foo","at":"1234567890"}`)
	a.NoError(err)

	s := session.(*azureadv2.Session)
	a.Equal(s.AuthURL, "http://foo")
	a.Equal(s.AccessToken, "1234567890")
}

func azureadProvider() *azureadv2.Provider {
	authConfig :=
		`{
    "TenantID": "` + tenantID + `",
    "ClientSecret": "` + secret + `",
    "ApplicationID": "` + applicationID + `"
}`

	p, _ := azureadv2.New([]byte(authConfig))
	return p
}
