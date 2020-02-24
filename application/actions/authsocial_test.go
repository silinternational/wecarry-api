package actions

import (
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

func (ts *TestSuite) Test_getConfig() {
	fbKey := "testFBKey"
	fbSecret := "testFBSecret"

	fbConfig := SocialAuthConfig{Key: fbKey, Secret: fbSecret}
	fbEnvVars := map[string]string{
		envTypeKey:    fbKey,
		envTypeSecret: fbSecret,
	}

	twKey := "testTwitterKey"
	twSecret := "testTwitterSecret"

	twConfig := SocialAuthConfig{Key: twKey, Secret: twSecret}
	twEnvVars := map[string]string{
		envTypeKey:    twKey,
		envTypeSecret: twSecret,
	}

	tests := []struct {
		name     string
		authType string
		envVars  map[string]string
		want     SocialAuthConfig
	}{
		{name: "Facebook", authType: AuthTypeFacebook, envVars: fbEnvVars, want: fbConfig},
		{name: "Twitter", authType: AuthTypeTwitter, envVars: twEnvVars, want: twConfig},
	}

	for _, tt := range tests {
		ts.T().Run(tt.name, func(t *testing.T) {
			got := getConfig(tt.authType, tt.envVars)
			ts.Equal(tt.want, got, "incorrect config")
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

func (as *ActionSuite) Test_CreateOrglessAuthUser() {

	uf := test.CreateUserFixtures(as.DB, 2)
	user := uf.Users[0]

	resultsAuthUser, err := createOrglessAuthUser("12345678", user)
	as.NoError(err)

	got := resultsAuthUser
	as.Equal(user.FirstName+" "+user.LastName, got.Name, "incorrect name")
	as.Equal(user.Nickname, got.Nickname, "incorrect nickname")
	as.Equal(user.Email, got.Email, "incorrect email")
	as.True(got.IsNew, "incorrect IsNew")
}
