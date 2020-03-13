package actions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/auth/facebook"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
)

// TestSuite establishes a test suite for domain tests
type TestSuite struct {
	suite.Suite
}

// Test_TestSuite runs the test suite
func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (ts *TestSuite) Test_addConfig() {
	fbKey := "testFBKey"
	fbSecret := "testFBSecret"

	fbConfig := SocialAuthConfig{Key: fbKey, Secret: fbSecret}

	twKey := "testTwitterKey"
	twSecret := "testTwitterSecret"

	twConfig := SocialAuthConfig{Key: twKey, Secret: twSecret}

	tests := []struct {
		name     string
		authType string
		key      string
		secret   string
		want     map[string]SocialAuthConfig
	}{
		{name: "Facebook", authType: AuthTypeFacebook, key: fbKey, secret: fbSecret,
			want: map[string]SocialAuthConfig{AuthTypeFacebook: fbConfig}},
		{name: "Twitter", authType: AuthTypeTwitter, key: twKey, secret: twSecret,
			want: map[string]SocialAuthConfig{AuthTypeTwitter: twConfig}},
		{name: "Twitter", authType: AuthTypeLinkedIn, key: "", secret: twSecret,
			want: map[string]SocialAuthConfig{}},
	}

	for _, tc := range tests {
		ts.T().Run(tc.name, func(t *testing.T) {
			configs := map[string]SocialAuthConfig{}
			addConfig(tc.authType, tc.key, tc.secret, configs)
			ts.Equal(tc.want, configs, "incorrect SocialAuthConfigs")
		})
	}
}

func (ts *TestSuite) Test_GetSocialAuthConfigs() {
	fbKey := "testFBKey"
	fbSecret := "testFBSecret"

	fbConfig := SocialAuthConfig{Key: fbKey, Secret: fbSecret}
	domain.Env.FacebookKey = fbKey
	domain.Env.FacebookSecret = fbSecret

	domain.Env.GoogleKey = ""
	domain.Env.GoogleSecret = "testGoogleSecret"

	domain.Env.LinkedInKey = "testLinkedInKey"
	domain.Env.LinkedInSecret = ""

	twKey := "testTwitterKey"
	twSecret := "testTwitterSecret"

	twConfig := SocialAuthConfig{Key: twKey, Secret: twSecret}
	domain.Env.TwitterKey = twKey
	domain.Env.TwitterSecret = twSecret

	got := getSocialAuthConfigs()

	want := map[string]SocialAuthConfig{
		AuthTypeFacebook: fbConfig,
		AuthTypeTwitter:  twConfig,
		// Others won't be included because of missing values
	}
	ts.Equal(want, got, "incorrect configs")
}

func (ts *TestSuite) Test_GetSocialAuthProvider() {
	domain.Env.FacebookKey = "testFBKey"
	domain.Env.FacebookSecret = "testFBSecret"

	got, err := getSocialAuthProvider(AuthTypeFacebook)
	ts.NoError(err, "unexpected error getting Facebook provider")
	ts.IsType(&facebook.Provider{}, got, "auth provider not expected facebook type")
}

func (ts *TestSuite) Test_getSocialAuthSelectors() {
	fbKey := "testFBKey"
	fbSecret := "testFBSecret"

	domain.Env.FacebookKey = fbKey
	domain.Env.FacebookSecret = fbSecret

	domain.Env.GoogleKey = ""
	domain.Env.GoogleSecret = "testGoogleSecret"

	domain.Env.LinkedInKey = "testLinkedInKey"
	domain.Env.LinkedInSecret = ""

	twKey := "testTwitterKey"
	twSecret := "testTwitterSecret"

	domain.Env.TwitterKey = twKey
	domain.Env.TwitterSecret = twSecret

	domain.Env.ApiBaseURL = "http://wecarry.local:3000"
	configs := getSocialAuthConfigs()
	got := getSocialAuthSelectors(configs)

	want := []authOption{
		{
			Name:        AuthTypeFacebook,
			RedirectURL: fmt.Sprintf("%s/auth/select/?%s=%s", domain.Env.ApiBaseURL, AuthTypeParam, AuthTypeFacebook),
		},
		{
			Name:        AuthTypeTwitter,
			RedirectURL: fmt.Sprintf("%s/auth/select/?%s=%s", domain.Env.ApiBaseURL, AuthTypeParam, AuthTypeTwitter),
		},
		// Others won't be included because of missing values
	}
	ts.Equal(want, got, "incorrect auth selectors")
}

func (as *ActionSuite) Test_newOrglessAuthUser() {

	uf := test.CreateUserFixtures(as.DB, 2)
	user := uf.Users[0]

	resultsAuthUser, err := newOrglessAuthUser("12345678", user)
	as.NoError(err)

	got := resultsAuthUser
	as.Equal(user.FirstName+" "+user.LastName, got.Name, "incorrect name")
	as.Equal(user.Nickname, got.Nickname, "incorrect nickname")
	as.Equal(user.Email, got.Email, "incorrect email")
	as.True(got.IsNew, "incorrect IsNew")
}
