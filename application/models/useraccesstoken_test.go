package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/domain"
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

	_, users, userOrgs := CreateUserFixtures(ms, t)
	tokens := CreateUserAccessTokenFixtures(ms, users[0], userOrgs)

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

	_, users, userOrgs := CreateUserFixtures(ms, t)
	tokens := CreateUserAccessTokenFixtures(ms, users[0], userOrgs)

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
				} else if u.User.UUID != test.want.UUID {
					t.Errorf("found %v, expected %v", u, test.want)
				}
			}
		})
	}
}

func CreateUserAccessTokenFixtures(ms *ModelSuite, user User, userOrgs UserOrganizations) []string {
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

	for i := range tokens {
		createFixture(ms, &tokens[i])
	}

	return rawTokens
}

func CreateUserFixtures_GetOrg(ms *ModelSuite, t *testing.T) ([]Organization, Users, UserOrganizations) {
	unique := domain.GetUUID().String()

	// Load Organization test fixtures
	orgs := []Organization{
		{
			Name:     fmt.Sprintf("ACME-%s", unique),
			AuthType: AuthTypeSaml,
		},
		{
			Name:     fmt.Sprintf("Starfleet Academy-%s", unique),
			AuthType: AuthTypeGoogle,
		},
	}
	for i := range orgs {
		orgs[i].AuthConfig = "{}"
		orgs[i].UUID = domain.GetUUID()
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
		},
		{
			Email:     fmt.Sprintf("user2-%s@example.com", unique),
			FirstName: "Another",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Another User %s", unique),
		},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		if err := ms.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user %v ... %v", users[i], err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{{}, {}}
	for i := range userOrgs {
		userOrgs[i].OrganizationID = orgs[i].ID
		userOrgs[i].UserID = users[i].ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		if err := ms.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrgs[i])
			t.FailNow()
		}
	}

	return orgs, users, userOrgs
}

func CreateUserAccessTokenFixtures_GetOrgs(ms *ModelSuite, users Users, userOrgs UserOrganizations) UserAccessTokens {
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

	for i := range tokens {
		createFixture(ms, &tokens[i])
	}

	return tokens
}

func (ms *ModelSuite) TestUserAccessToken_GetOrganization() {
	t := ms.T()

	orgs, users, userOrgs := CreateUserFixtures_GetOrg(ms, t)

	tokens := CreateUserAccessTokenFixtures_GetOrgs(ms, users, userOrgs)

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

type AccessTokenFixtures struct {
	Users
	UserAccessTokens
}

// CreateFixtures_GetUser creates test fixtures for the GetUser test function
func CreateFixtures_GetUser(ms *ModelSuite, t *testing.T) AccessTokenFixtures {
	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	users := Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1", UUID: domain.GetUUID()},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2", UUID: domain.GetUUID()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{{}, {}}
	for i := range userOrgs {
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].UserID = users[i].ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		createFixture(ms, &userOrgs[i])
	}

	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[1].ID,
			UserOrganizationID: userOrgs[1].ID,
			AccessToken:        HashClientIdAccessToken("xyz789"),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	for i := range tokens {
		createFixture(ms, &tokens[i])
	}

	return AccessTokenFixtures{
		Users:            users,
		UserAccessTokens: tokens,
	}
}

func (ms *ModelSuite) TestUserAccessToken_GetUser() {
	t := ms.T()

	f := CreateFixtures_GetUser(ms, t)

	tests := []struct {
		name    string
		token   UserAccessToken
		want    string
		wantErr bool
	}{
		{name: "user1", token: f.UserAccessTokens[0], want: f.Users[0].UUID.String()},
		{name: "user2", token: f.UserAccessTokens[1], want: f.Users[1].UUID.String()},
		{name: "noUserOrg", token: UserAccessToken{}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u := test.token
			got, err := u.GetUser()
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("GetUser() returned an error: %v", err)
				} else if got.UUID.String() != test.want {
					t.Errorf("found %v, expected %v", got.UUID.String(), test.want)
				}
			}
		})
	}
}

// CreateFixtures_DeleteIfExpired creates test fixtures for the DeleteIfExpired test function
func CreateFixtures_DeleteIfExpired(ms *ModelSuite, t *testing.T) AccessTokenFixtures {
	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	users := Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1"},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{{}, {}}
	for i := range userOrgs {
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].UserID = users[i].ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		createFixture(ms, &userOrgs[i])
	}

	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[1].ID,
			UserOrganizationID: userOrgs[1].ID,
			AccessToken:        HashClientIdAccessToken("xyz789"),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	for i := range tokens {
		createFixture(ms, &tokens[i])
	}

	return AccessTokenFixtures{
		Users:            users,
		UserAccessTokens: tokens,
	}
}

func (ms *ModelSuite) TestUserAccessToken_DeleteIfExpired() {
	t := ms.T()
	f := CreateFixtures_DeleteIfExpired(ms, t)

	tests := []struct {
		name  string
		token UserAccessToken
		want  bool
	}{
		{name: "expired", token: f.UserAccessTokens[0], want: true},
		{name: "not expired", token: f.UserAccessTokens[1], want: false},
		{name: "empty token", token: UserAccessToken{}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u := test.token
			got, err := u.DeleteIfExpired()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			ms.Equal(test.want, got)
		})
	}
}

// CreateFixtures_Renew creates test fixtures for the Renew test function
func CreateFixtures_Renew(ms *ModelSuite, t *testing.T) AccessTokenFixtures {
	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	users := Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1"},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{{}, {}}
	for i := range userOrgs {
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].UserID = users[i].ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		createFixture(ms, &userOrgs[i])
	}

	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[1].ID,
			UserOrganizationID: userOrgs[1].ID,
			AccessToken:        HashClientIdAccessToken("xyz789"),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	for i := range tokens {
		createFixture(ms, &tokens[i])
	}

	return AccessTokenFixtures{
		Users:            users,
		UserAccessTokens: tokens,
	}
}

func (ms *ModelSuite) TestUserAccessToken_Renew() {
	t := ms.T()

	f := CreateFixtures_Renew(ms, t)

	tests := []struct {
		name  string
		token UserAccessToken
	}{
		{name: "expired", token: f.UserAccessTokens[0]},
		{name: "not expired", token: f.UserAccessTokens[1]},
		{name: "empty token", token: UserAccessToken{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u := test.token
			_ = u.Renew()
			ms.False(u.DeleteIfExpired())
		})
	}
}

func createFixtures_UserAccessTokensDeleteExpired(ms *ModelSuite, t *testing.T) AccessTokenFixtures {

	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	users := Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1"},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{{}, {}}
	for i := range userOrgs {
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].UserID = users[i].ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		createFixture(ms, &userOrgs[i])
	}

	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[1].ID,
			UserOrganizationID: userOrgs[1].ID,
			AccessToken:        HashClientIdAccessToken("xyz789"),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	for i := range tokens {
		createFixture(ms, &tokens[i])
	}

	return AccessTokenFixtures{
		Users:            users,
		UserAccessTokens: tokens,
	}
}

func (ms *ModelSuite) TestUserAccessToken_DeleteExpired() {
	t := ms.T()

	f := createFixtures_UserAccessTokensDeleteExpired(ms, t)

	var uats UserAccessTokens
	got, err := uats.DeleteExpired()

	ms.NoError(err)

	want := 1

	ms.Equal(want, got, "Wrong count of deleted tokens.")

	var uatsLeft UserAccessTokens
	if err := ms.DB.All(&uatsLeft); err != nil {
		t.Errorf("Couldn't get tokens in order to complete test ... %v", err)
		return
	}

	got = len(uatsLeft)
	ms.Equal(want, got, "Deleted wrong number of tokens.")

	gotToken := uatsLeft[0].AccessToken
	wantToken := f.UserAccessTokens[1].AccessToken

	ms.Equal(wantToken, gotToken, "Wrong token remaining.")
}

func (ms *ModelSuite) TestGetRandomToken() {
	got1, _ := getRandomToken()
	got2, err := getRandomToken()

	ms.NoError(err)

	ms.True(got1 != got2, "Expected two different access tokens, but got the same one")

	got := len(got1)

	thirds := TokenBytes / 3
	if TokenBytes%3 > 0 {
		thirds++
	}

	want := thirds * 4 // 44 = 32 / 3 (rounded up)  * 4

	ms.Equal(want, got, "Wrong length of access token. Got ... %s", got1)
}
