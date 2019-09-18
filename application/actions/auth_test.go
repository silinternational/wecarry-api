package actions

import (
	"testing"
)

func (as *ActionSuite) TestGetLoginSuccessRedirectURL() {
	t := as.T()

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
			wantBeginning: "/#/welcome?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=new",
		},
		{
			name:          "New With ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: true, AccessToken: "new"},
			returnTo:      "/posts",
			wantBeginning: "/#/welcome?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=new&ReturnTo=/posts",
		},
		{
			name:          "Not New ReturnTo Without a Slash",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old1"},
			returnTo:      "posts",
			wantBeginning: "/#/posts?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=old1",
		},
		{
			name:          "Not New With a Good ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old2"},
			returnTo:      "/posts",
			wantBeginning: "/#/posts?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=old2",
		},
		{
			name:          "Not New With No ReturnTo",
			authUser:      AuthUser{ID: "1", IsNew: false, AccessToken: "old3"},
			returnTo:      "",
			wantBeginning: "/#?token_type=Bearer&expires_utc=",
			wantEnd:       "&access_token=old3",
		},
	}
	for _, test := range tests {
		// Test the first part and last part of the resulting urls
		t.Run(test.name, func(t *testing.T) {
			allResults := getLoginSuccessRedirectURL(test.authUser, test.returnTo)

			expected := test.wantBeginning
			beginningResults := allResults[0: len(expected)]

			if beginningResults != expected {
				t.Errorf("Bad results at beginning for test \"%s\". \nExpected %s\n  but got %s",
					test.name, expected, allResults)
				return
			}

			expected = test.wantEnd
			endResults := allResults[len(allResults) - len(expected) : len(allResults)]
			if endResults != expected {
				t.Errorf("Bad results at end for test \"%s\". \nExpected %s\n  but got %s",
					test.name, expected, allResults)
			}
		})
	}
}
