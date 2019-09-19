package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/silinternational/handcarry-api/auth"
	"github.com/silinternational/handcarry-api/models"
	"testing"
)

func (as *ActionSuite) TestGetLoginSuccessRedirectURL() {
	t := as.T()

	uiURL := envy.Get("UI_URL", "")

	tests := []struct {
		name          string
		authUser      AuthUser
		returnTo      string
		wantBeginning string
		wantEnd       string
	}{
		{
			name:          "New No ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: true, AccessToken: "new"},
			returnTo:      "",
			wantBeginning: uiURL + "/#/welcome?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=new",
		},
		{
			name:          "New With ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: true, AccessToken: "new"},
			returnTo:      "/posts",
			wantBeginning: uiURL + "/#/welcome?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=new&ReturnTo=/posts",
		},
		{
			name:          "Not New ReturnTo Without a Slash",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old1"},
			returnTo:      "posts",
			wantBeginning: uiURL + "/#/posts?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=old1",
		},
		{
			name:          "Not New With a Good ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old2"},
			returnTo:      "/posts",
			wantBeginning: uiURL + "/#/posts?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=old2",
		},
		{
			name:          "Not New With No ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old3"},
			returnTo:      "",
			wantBeginning: uiURL + "/#?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=old3",
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {
			allResults := getLoginSuccessRedirectURL(test.authUser, test.returnTo)

			expected := test.wantBeginning
			beginningResults := allResults[0:len(expected)]

			if beginningResults != expected {
				t.Errorf("Bad results at beginning for test \"%s\". \nExpected %s\n  but got %s",
					test.name, expected, allResults)
				return
			}

			expected = test.wantEnd
			endResults := allResults[len(allResults)-len(expected) : len(allResults)]
			if endResults != expected {
				t.Errorf("Bad results at end for test \"%s\". \nExpected %s\n  but got %s",
					test.name, expected, allResults)
			}
		})
	}
}

type bufTestCtx struct {
	buffalo.DefaultContext
	params map[string]string
	sess   *buffalo.Session
}

func (b *bufTestCtx) setParam(key, value string) {
	b.params[key] = value
}

func (b *bufTestCtx) Param(key string) string {
	return b.params[key]
}

func (b *bufTestCtx) Session() *buffalo.Session {
	return b.sess
}

func (b *bufTestCtx) Render(status int, r render.Renderer) error {
	return fmt.Errorf("%v", status)
}

func (as *ActionSuite) TestGetOrSetClientID() {
	t := as.T()

	tests := []struct {
		name         string
		param        string
		sessionValue string
		returnTo     string
		wantErr      bool
		want         string
		wantSession  string
	}{
		{
			name:         "No Param No Session",
			param:        "",
			sessionValue: "",
			wantErr:      true,
		},
		{
			name:         "No Param But With Session",
			param:        "",
			sessionValue: "2222",
			wantErr:      false,
			want:         "2222",
			wantSession:  "2222",
		},
		{
			name:         "With Param But Not With Session",
			param:        "3333",
			sessionValue: "",
			wantErr:      false,
			want:         "3333",
			wantSession:  "3333",
		},
		{
			name:         "With Param And With Session",
			param:        "444A",
			sessionValue: "444B",
			wantErr:      false,
			want:         "444A",
			wantSession:  "444A",
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {

			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			c.params[ClientIDParam] = test.param

			if test.sessionValue != "" {
				c.Session().Set(ClientIDSessionKey, test.sessionValue)
				c.Session().Save()
			} else {
				c.Session().Clear()
				c.Session().Save()
			}

			results, err := getOrSetClientID(c)

			if test.wantErr && err == nil {
				t.Errorf("for test \"%s\" expected an error but did not get one.", test.name)
				return
			}

			if !test.wantErr && err != nil {
				t.Errorf("unexpected error for test \"%s\" ...  %v", test.name, err)
				return
			}

			expected := test.want

			if results != expected {
				t.Errorf("bad results for test \"%s\". \nExpected %s\n but got %s",
					test.name, expected, results)
				return
			}

			expected = test.wantSession
			if expected != "" {
				results = fmt.Sprintf("%v", c.sess.Get(ClientIDSessionKey))
				if results != expected {
					t.Errorf("bad results for test \"%s\". \nExpected %s\n but got %s",
						test.name, expected, results)
					return
				}
			}
		})
	}
}

func (as *ActionSuite) TestGetOrSetAuthEmail() {
	t := as.T()

	tests := []struct {
		name         string
		param        string
		sessionValue string
		returnTo     string
		wantErr      bool
		want         string
		wantSession  string
	}{
		{
			name:         "No Param No Session",
			param:        "",
			sessionValue: "",
			wantErr:      true,
		},
		{
			name:         "No Param But With Session",
			param:        "",
			sessionValue: "sess@example.com",
			wantErr:      false,
			want:         "sess@example.com",
			wantSession:  "sess@example.com",
		},
		{
			name:         "With Param But Not With Session",
			param:        "param@example.com",
			sessionValue: "",
			wantErr:      false,
			want:         "param@example.com",
			wantSession:  "param@example.com",
		},
		{
			name:         "With Param And With Session",
			param:        "param@example.com",
			sessionValue: "sess@example.com",
			wantErr:      false,
			want:         "sess@example.com",
			wantSession:  "sess@example.com",
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {

			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			c.params[AuthEmailParam] = test.param

			if test.sessionValue != "" {
				c.Session().Set(AuthEmailSessionKey, test.sessionValue)
				c.Session().Save()
			} else {
				c.Session().Clear()
				c.Session().Save()
			}

			results, err := getOrSetAuthEmail(c)

			if test.wantErr && err == nil {
				t.Errorf("for test \"%s\" expected an error but did not get one.", test.name)
				return
			}

			if !test.wantErr && err != nil {
				t.Errorf("unexpected error for test \"%s\" ...  %v", test.name, err)
				return
			}

			expected := test.want

			if results != expected {
				t.Errorf("bad results for test \"%s\". \nExpected %s\n but got %s",
					test.name, expected, results)
				return
			}

			expected = test.wantSession
			if expected != "" {
				results = fmt.Sprintf("%v", c.sess.Get(AuthEmailSessionKey))
				if results != expected {
					t.Errorf("bad results for test \"%s\". \nExpected %s\n but got %s",
						test.name, expected, results)
					return
				}
			}
		})
	}
}

func (as *ActionSuite) TestGetOrSetReturnTo() {
	t := as.T()

	tests := []struct {
		name         string
		param        string
		sessionValue string
		returnTo     string
		want         string
		wantSession  string
	}{
		{
			name:         "No Param No Session",
			param:        "",
			sessionValue: "",
			want:         "/#",
		},
		{
			name:         "No Param But With Session",
			param:        "",
			sessionValue: "sess.example.com",
			want:         "sess.example.com",
			wantSession:  "sess.example.com",
		},
		{
			name:         "With Param But Not With Session",
			param:        "param.example.com",
			sessionValue: "",
			want:         "param.example.com",
			wantSession:  "param.example.com",
		},
		{
			name:         "With Param And With Session",
			param:        "param.example.com",
			sessionValue: "sess.example.com",
			want:         "param.example.com",
			wantSession:  "param.example.com",
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {

			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			c.params[ReturnToParam] = test.param

			if test.sessionValue != "" {
				c.Session().Set(ReturnToSessionKey, test.sessionValue)
				c.Session().Save()
			} else {
				c.Session().Clear()
				c.Session().Save()
			}

			results := getOrSetReturnTo(c)
			expected := test.want

			if results != expected {
				t.Errorf("bad results for test \"%s\". \nExpected %s\n but got %s",
					test.name, expected, results)
				return
			}

			expected = test.wantSession
			if expected != "" {
				results = fmt.Sprintf("%v", c.sess.Get(ReturnToSessionKey))
				if results != expected {
					t.Errorf("bad results for test \"%s\". \nExpected %s\n but got %s",
						test.name, expected, results)
					return
				}
			}
		})
	}
}

// This doesn't test for errors, since it's too complicated with the call to domain.Error()
func (as *ActionSuite) TestGetOrgAndUserOrgs() {
	t := as.T()
	models.ResetTables(t, as.DB)

	fixtures := Fixtures_GetOrgAndUserOrgs(as, t)
	orgFixture := fixtures.orgs[0]
	userOrgFixtures := fixtures.userOrgs

	tests := []struct {
		name             string
		authEmail        string
		param            string
		wantOrg          string
		wantUserOrg      string
		wantUserOrgCount int
	}{
		{
			name:             "No org_id Param But With UserOrg For AuthEmail",
			authEmail:        userOrgFixtures[0].AuthEmail,
			param:            "",
			wantOrg:          orgFixture.Name,
			wantUserOrgCount: 1,
			wantUserOrg:      userOrgFixtures[0].AuthEmail,
		},
		{
			name:             "With bad org_id Param But With UserOrg for AuthEmail",
			authEmail:        userOrgFixtures[0].AuthEmail,
			param:            "11",
			wantOrg:          orgFixture.Name,
			wantUserOrgCount: 0,
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {

			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			c.params[OrgIDParam] = test.param

			resultOrg, resultUserOrgs, _ := getOrgAndUserOrgs(test.authEmail, c)

			expected := test.wantOrg
			results := resultOrg.Name

			if results != expected {
				t.Errorf("bad Org results for test \"%s\". \nExpected %s\n but got %s",
					test.name, expected, results)
				return
			}

			if len(resultUserOrgs) != test.wantUserOrgCount {
				t.Errorf("bad results for test \"%s\". \nExpected %v UserOrg but got %v ... \n %+v\n",
					test.name, test.wantUserOrgCount, len(resultUserOrgs), resultUserOrgs)
				return
			}

			if test.wantUserOrgCount == 1 {

				expected = test.wantUserOrg
				results = resultUserOrgs[0].AuthEmail

				if results != expected {
					t.Errorf("bad UserOrg results for test \"%s\". \nExpected %s\n but got %s",
						test.name, expected, results)
					return
				}
			}
		})
	}
}

func (as *ActionSuite) TestCreateAuthUser() {
	t := as.T()
	models.ResetTables(t, as.DB)
	orgFixture := Fixtures_CreateAuthUser(as, t).orgs[0]
	c := &bufTestCtx{
		sess:   as.Session,
		params: map[string]string{},
	}

	newEmail := "new@example.com"

	authUser := auth.User{
		Email:     newEmail,
		FirstName: "First",
		LastName:  "Last",
	}

	var user models.User
	err := user.FindOrCreateFromAuthUser(orgFixture.ID, &authUser)
	if err != nil {
		t.Errorf("could not run test because of error calling user.FindOrCreateFromAuthUser ...\n %v", err)
		return
	}

	resultsAuthUser, err := createAuthUser(newEmail, "12345678", user, orgFixture, c)

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		return
	}

	expected := newEmail
	results := resultsAuthUser.Email

	if results != expected {
		t.Errorf("bad email results: expected %v but got %v", expected, results)
	}

}
