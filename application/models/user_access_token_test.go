package models

import (
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
	"testing"
	"time"
)

func (ms *ModelSuite) TestUserAccessToken_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		token    UserAccessToken
		wantErr  bool
		errField string
	}{
		{
			name:    "minimum",
			token:   UserAccessToken{UserID: 1, AccessToken: "abc123", ExpiresAt: time.Now()},
			wantErr: false,
		},
		{
			name:     "missing user_id",
			token:    UserAccessToken{AccessToken: "abc123", ExpiresAt: time.Now()},
			wantErr:  true,
			errField: "user_id",
		},
		{
			name:     "missing access_token",
			token:    UserAccessToken{UserID: 1, ExpiresAt: time.Now()},
			wantErr:  true,
			errField: "access_token",
		},
		{
			name:     "missing expires_at",
			token:    UserAccessToken{UserID: 1, AccessToken: "abc123"},
			wantErr:  true,
			errField: "expires_at",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.token.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestUserAccessToken_DeleteByBearerToken() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, userOrgs := CreateUserFixtures(ms, t)
	tokens := CreateUserAccessTokenFixtures(t, users[0], userOrgs)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{name: "success1", token: tokens[0], wantErr: false},
		{name: "success2", token: tokens[1], wantErr: false},
		{name: "failure", token: "------", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var uat UserAccessToken
			err := uat.DeleteByBearerToken(test.token)
			if err != nil && !test.wantErr {
				t.Errorf("DeleteAccessToken() returned an unexpected error: %s", err)
			} else if err == nil && test.wantErr {
				t.Errorf("expected an error, but DeleteAccessToken() did not return an error")
			}
		})
	}
}

func (ms *ModelSuite) TestUserAccessToken_FindByBearerToken() {
	t := ms.T()
	ResetTables(t, ms.DB)

	_, users, userOrgs := CreateUserFixtures(ms, t)
	tokens := CreateUserAccessTokenFixtures(t, users[0], userOrgs)

	tests := []struct {
		name    string
		token   string
		want    User
		wantErr bool
	}{
		{name: "valid0", token: tokens[0], want: users[0]},
		{name: "valid1", token: tokens[1], want: users[0]},
		{name: "invalid", token: "000000", wantErr: true},
		{name: "empty", token: "", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u UserAccessToken
			err := u.FindByBearerToken(test.token)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindByAccessToken() returned an error: %v", err)
				} else if u.User.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", u, test.want)
				}
			}
		})
	}
}

func CreateUserAccessTokenFixtures(t *testing.T, user User, userOrgs UserOrganizations) []string {
	rawTokens := []string{"abc123", "xyz789"}
	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:             user.ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken(rawTokens[0]),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             user.ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken(rawTokens[1]),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := CreateUserAccessTokens(tokens); err != nil {
		t.Errorf("could not create access tokens ... %v", err)
		t.FailNow()
	}

	return rawTokens
}

func CreateUserFixtures_GetOrg(ms *ModelSuite, t *testing.T) ([]Organization, Users, UserOrganizations) {
	ResetTables(t, ms.DB)

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	orgs := []Organization{
		{
			Name:       fmt.Sprintf("ACME-%s", unique),
			Uuid:       domain.GetUuid(),
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
		},
		{
			Name:       fmt.Sprintf("Starfleet Academy-%s", unique),
			Uuid:       domain.GetUuid(),
			AuthType:   AuthTypeGoogle,
			AuthConfig: "{}",
		},
	}
	for i := range orgs {
		if err := ms.DB.Create(&orgs[i]); err != nil {
			t.Errorf("error creating org %+v ...\n %v \n", orgs[i], err)
			t.FailNow()
		}
	}

	// Load User test fixtures
	users := Users{
		{
			Email:     fmt.Sprintf("user1-%s@example.com", unique),
			FirstName: "Existing",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Existing User %s", unique),
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     fmt.Sprintf("user2-%s@example.com", unique),
			FirstName: "Another",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Another User %s", unique),
			Uuid:      domain.GetUuid(),
		},
	}
	for i := range users {
		if err := ms.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user %v ... %v", users[i], err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{
		{
			OrganizationID: orgs[0].ID,
			UserID:         users[0].ID,
			AuthID:         users[0].Email,
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: orgs[1].ID,
			UserID:         users[1].ID,
			AuthID:         users[1].Email,
			AuthEmail:      users[1].Email,
		},
	}
	for i := range userOrgs {
		if err := ms.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrgs[i])
			t.FailNow()
		}
	}

	return orgs, users, userOrgs
}

func CreateUserAccessTokenFixtures_GetOrgs(t *testing.T, users Users, userOrgs UserOrganizations) UserAccessTokens {
	rawTokens := []string{"abc123", "xyz789"}
	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken(rawTokens[0]),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[1].ID,
			UserOrganizationID: userOrgs[1].ID,
			AccessToken:        HashClientIdAccessToken(rawTokens[1]),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := CreateUserAccessTokens(tokens); err != nil {
		t.Errorf("could not create access tokens ... %v", err)
		t.FailNow()
	}

	return tokens
}

func (ms *ModelSuite) TestUserAccessToken_GetOrganization() {
	t := ms.T()
	ResetTables(t, ms.DB)

	orgs, users, userOrgs := CreateUserFixtures_GetOrg(ms, t)

	tokens := CreateUserAccessTokenFixtures_GetOrgs(t, users, userOrgs)

	tests := []struct {
		name    string
		token   UserAccessToken
		want    string
		wantErr bool
	}{
		{name: "org0", token: tokens[0], want: orgs[0].AuthType},
		{name: "org1", token: tokens[1], want: orgs[1].AuthType},
		{name: "noUserOrg", token: UserAccessToken{}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u := test.token
			got, err := u.GetOrganization()
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("GetOrganization() returned an error: %v", err)
				} else if got.AuthType != test.want {
					t.Errorf("found %v, expected %v", u, test.want)
				}
			}
		})
	}
}
