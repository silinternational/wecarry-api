package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
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
		},
		{
			name:         "With Param But Not With Session",
			param:        "3333",
			sessionValue: "",
			wantErr:      false,
			want:         "3333",
		},
		{
			name:         "With Param And With Session",
			param:        "444A",
			sessionValue: "444B",
			wantErr:      false,
			want:         "444A",
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {

			c := &bufTestCtx{
				sess:   as.Session,
				params: map[string]string{},
			}

			c.params["client_id"] = test.param

			if test.sessionValue != "" {
				c.Session().Set("ClientID", test.sessionValue)
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
		})
	}

}
