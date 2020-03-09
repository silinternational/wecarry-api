package actions

import (
	"fmt"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/logger"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) TestVerifyEmails() {
	t := as.T()

	tests := []struct {
		name      string
		authEmail string
		respEmail string
		wantErr   bool
	}{
		{
			name:      "matching",
			authEmail: "good@example.com",
			respEmail: "good@example.com",
		},
		{
			name:      "matching domains",
			authEmail: "good@example.com",
			respEmail: "other@example.com",
		},
		{
			name:      "different domains",
			authEmail: "good@example.com",
			respEmail: "good@other.com",
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			gotErr := verifyEmails(c, test.authEmail, test.respEmail)
			if test.wantErr {
				as.Error(gotErr, "missing expected error")
				return
			}
			as.NoError(gotErr, "unexpected error")
		})
	}
}

func (as *ActionSuite) TestGetLoginSuccessRedirectURL() {
	t := as.T()

	uiURL := domain.Env.UIURL

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
			wantBeginning: uiURL + "/#/welcome?" + TokenTypeParam + "=Bearer&" + ExpiresUTCParam + "=",
			wantEnd:       "&" + AccessTokenParam + "=new",
		},
		{
			name:          "New With Invalid ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: true, AccessToken: "new"},
			returnTo:      "/posts",
			wantBeginning: uiURL + "/#/welcome?" + TokenTypeParam + "=Bearer&" + ExpiresUTCParam + "=",
			wantEnd:       "&" + AccessTokenParam + "=new&" + ReturnToParam + "=%2Fposts",
		},
		{
			name:          "New With Valid ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: true, AccessToken: "new"},
			returnTo:      "/#/posts",
			wantBeginning: uiURL + "/#/welcome?" + TokenTypeParam + "=Bearer&" + ExpiresUTCParam + "=",
			wantEnd:       "&" + AccessTokenParam + "=new&" + ReturnToParam + "=%2Fposts",
		},
		{
			name:          "Not New With Invalid ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old1"},
			returnTo:      "/posts",
			wantBeginning: uiURL + "/#/posts?" + TokenTypeParam + "=Bearer&" + ExpiresUTCParam + "=",
			wantEnd:       "&" + AccessTokenParam + "=old1",
		},
		{
			name:          "Not New With a Good ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old2"},
			returnTo:      "/#/posts",
			wantBeginning: uiURL + "/#/posts?" + TokenTypeParam + "=Bearer&" + ExpiresUTCParam + "=",
			wantEnd:       "&" + AccessTokenParam + "=old2",
		},
		{
			name:          "Not New With No ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old3"},
			returnTo:      "",
			wantBeginning: uiURL + "/#?" + TokenTypeParam + "=Bearer&" + ExpiresUTCParam + "=",
			wantEnd:       "&" + AccessTokenParam + "=old3",
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
				return
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

func (b *bufTestCtx) Value(key interface{}) interface{} {
	return nil
}

func (b *bufTestCtx) Logger() logger.FieldLogger {
	return testLogger{}
}

type testLogger struct{}

func (l testLogger) Debugf(s string, i ...interface{}) {}
func (l testLogger) Infof(s string, i ...interface{})  {}
func (l testLogger) Printf(s string, i ...interface{}) {}
func (l testLogger) Warnf(s string, i ...interface{})  {}
func (l testLogger) Errorf(s string, i ...interface{}) {}
func (l testLogger) Fatalf(s string, i ...interface{}) {}
func (l testLogger) Debug(i ...interface{})            {}
func (l testLogger) Info(i ...interface{})             {}
func (l testLogger) Warn(i ...interface{})             {}
func (l testLogger) Error(i ...interface{})            {}
func (l testLogger) Fatal(i ...interface{})            {}
func (l testLogger) Panic(i ...interface{})            {}

func (l testLogger) WithField(s string, i interface{}) logger.FieldLogger {
	return testLogger{}
}
func (l testLogger) WithFields(m map[string]interface{}) logger.FieldLogger {
	return testLogger{}
}

func (as *ActionSuite) TestGetAuthInviteResponse() {
	f := createFixturesForAuthInvite(as)
	meetings := f.Meetings

	testCases := []struct {
		name        string
		param       string
		want        authInviteResponse
		wantObjUUID string
		wantErr     bool
		wantErrCode string
	}{
		{
			name:        "No Invite Code",
			wantErr:     true,
			wantErrCode: "400",
		},
		{
			name:  "Valid Invite Code",
			param: meetings[1].InviteCode.UUID.String(),
			want: authInviteResponse{
				Type:     InviteTypeMeetingParam,
				Name:     meetings[1].Name,
				ImageURL: f.File.URL,
			},
			wantObjUUID: meetings[1].UUID.String(),
		},
		{
			name:        "Invalid Invite Code",
			param:       domain.GetUUID().String(),
			wantErr:     true,
			wantErrCode: "404",
		},
	}
	for _, tc := range testCases {
		// Test the first part and last part of the resulting urls
		as.T().Run(tc.name, func(t *testing.T) {
			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			if tc.param != "" {
				c.params[InviteCodeParam] = tc.param
			}

			got, err := getAuthInviteResponse(c)
			if tc.wantErr {
				as.Error(err, "expected and error but didn't get one")
				as.Equal(tc.wantErrCode, err.Error(), "invalid error code")
				return
			}
			as.NoError(err)
			as.Equal(tc.want, got, "incorrect authInviteResponse")
			as.Equal(InviteTypeMeetingParam, c.Session().Get(InviteTypeSessionKey), "incorrect session Invite Type")
			as.Equal(tc.wantObjUUID, c.Session().Get(InviteObjectUUIDSessionKey), "incorrect session Invite Object UUID")
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
			want:         domain.DefaultUIPath,
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
					t.Errorf("bad session results for test \"%s\". \nExpected %s\n but got %s",
						test.name, expected, results)
					return
				}
			}
		})
	}
}

// This doesn't test for errors, since it's too complicated with the call to domain.Error()
func (as *ActionSuite) TestGetUserOrgs() {
	t := as.T()

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

			resultUserOrgs, _ := getUserOrgs(c, test.authEmail)

			expected := test.wantOrg

			if len(resultUserOrgs) != test.wantUserOrgCount {
				t.Errorf("bad results for test \"%s\". \nExpected %v UserOrg(s) but got %v ... \n %+v\n",
					test.name, test.wantUserOrgCount, len(resultUserOrgs), resultUserOrgs)
				return
			}

			if test.wantUserOrgCount == 1 {

				expected = test.wantUserOrg
				results := resultUserOrgs[0].AuthEmail

				if results != expected {
					t.Errorf("bad UserOrg results for test \"%s\". \nExpected %s\n but got %s",
						test.name, expected, results)
					return
				}
			}
		})
	}
}

func (as *ActionSuite) Test_newAuthUser() {
	t := as.T()
	orgFixture := Fixtures_newAuthUser(as, t).orgs[0]

	newEmail := "new@example.com"

	authUser := auth.User{
		Email:     newEmail,
		FirstName: "First",
		LastName:  "Last",
		UserID:    newEmail,
	}

	var user models.User
	err := user.FindOrCreateFromAuthUser(orgFixture.ID, &authUser)
	if err != nil {
		t.Errorf("could not run test because of error calling user.FindOrCreateFromAuthUser ...\n %v", err)
		return
	}

	resultsAuthUser, err := newOrgBasedAuthUser("12345678", user, orgFixture)

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

func (as *ActionSuite) TestEnsureMeetingParticipant() {
	f := createFixturesForEnsureMeetingParticipant(as)

	testCases := []struct {
		name            string
		meeting         models.Meeting
		user            models.User
		wantParticipant bool
	}{
		{
			name:            "With Invite but no Participant",
			meeting:         f.Meetings[1],
			user:            f.Users[1],
			wantParticipant: true,
		},
		{
			name:            "With Invite and Participant already created",
			meeting:         f.Meetings[1],
			user:            f.Users[1],
			wantParticipant: true,
		},
		{
			name:            "No Invite",
			meeting:         f.Meetings[1],
			user:            f.Users[0],
			wantParticipant: false,
		},
	}
	for _, tc := range testCases {
		// Test the first part and last part of the resulting urls
		as.T().Run(tc.name, func(t *testing.T) {
			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			ensureMeetingParticipant(c, tc.meeting.UUID.String(), tc.user)

			var participants models.MeetingParticipants
			if tc.wantParticipant {
				err := models.DB.Where("meeting_id = ? and user_id = ?", tc.meeting.ID, tc.user.ID).All(&participants)
				as.NoError(err, "unexpected error finding MeetingParticipants")
				as.Equal(1, len(participants), "incorrect number of MeetingParticipants")
				return
			}

			// Don't want a new Participant
			err := models.DB.Where("meeting_id = ? and user_id = ?", tc.meeting.ID, tc.user.ID).All(&participants)
			as.NoError(err, "unexpected error finding no MeetingParticipants")
			as.Equal(0, len(participants), "expected to find no MeetingParticipants")
		})
	}
}
