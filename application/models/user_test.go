package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/silinternational/handcarry-api/auth"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/domain"
)

func TestUser_FindOrCreateFromAuthUser(t *testing.T) {
	resetTables(t)

	// create org for test
	org := &Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := models.DB.Create(org)
	if err != nil {
		t.Errorf("Failed to create organization for test, error: %s", err)
		t.FailNow()
	}

	type args struct {
		tx       *pop.Connection
		orgID    int
		authUser *auth.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantID  int
	}{
		{
			name: "create new user: test_user1",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     "test_user1@domain.com",
					UserID:    "test_user1",
				},
			},
			wantErr: false,
			wantID:  1,
		},
		{
			name: "find existing user: test_user1",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     "test_user1@domain.com",
					UserID:    "test_user1",
				},
			},
			wantErr: false,
			wantID:  1,
		},
		{
			name: "conflicting user",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     "test_user1@domain.com",
					UserID:    "test_user2",
				},
			},
			wantErr: true,
			wantID:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{}
			err := u.FindOrCreateFromAuthUser(tt.args.orgID, tt.args.authUser)
			if err != nil && !tt.wantErr {
				t.Errorf("FindOrCreateFromAuthUser() error = %v, wantErr %v", err, tt.wantErr)
			} else if u.ID != tt.wantID {
				t.Errorf("ID on user is not what was wanted. Got %v, wanted %v", u.ID, tt.wantID)
			}
		})
	}
}

func TestFindUserByAccessToken(t *testing.T) {
	_, user, userOrgs := createUserFixtures(t)

	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:             user.ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        hashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             user.ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        hashClientIdAccessToken("xyz789"),
			ExpiresAt:          time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := CreateUserAccessTokens(tokens); err != nil {
		t.Errorf("could not create access tokens ... %v", err)
		t.FailNow()
	}

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    User
		wantErr bool
	}{
		{
			name:    "expired",
			args:    args{"abc123"},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{"xyz789"},
			want: user,
		},
		{
			name:    "invalid",
			args:    args{"000000"},
			wantErr: true,
		},
		{
			name:    "empty",
			args:    args{""},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FindUserByAccessToken(test.args.token)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindUserByAccessToken() returned an error: %v", err)
				} else if got.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", got, test.want)
				}
			}
		})
	}

	resetTables(t) // Pack it in, Pack it out a/k/a "Leave No Trace"
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing email",
			user: User{
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "missing first_name",
			user: User{
				Email:    "user@example.com",
				LastName: "User",
				Nickname: "A User",
				Uuid:     domain.GetUuid(),
			},
			wantErr:  true,
			errField: "first_name",
		},
		{
			name: "missing last_name",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "last_name",
		},
		{
			name: "missing nickname",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "nickname",
		},
		{
			name: "missing uuid",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
			},
			wantErr:  true,
			errField: "uuid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.user.Validate(DB)
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

func TestCreateAccessToken(t *testing.T) {
	orgs, user, _ := createUserFixtures(t)

	type args struct {
		user     User
		clientID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				user:     user,
				clientID: "abc123",
			},
			wantErr: false,
		},
		{
			name: "empty client ID",
			args: args{
				user:     user,
				clientID: "",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedExpiry := createAccessTokenExpiry().Unix()
			token, expiry, err := test.args.user.CreateAccessToken(orgs[0], test.args.clientID)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, but did not get one")
				}
			} else {
				if err != nil && !test.wantErr {
					t.Errorf("CreateAccessToken() returned error: %v", err)
				}
				hash := hashClientIdAccessToken(test.args.clientID + token)

				var dbToken UserAccessToken
				if err := DB.Where(fmt.Sprintf("access_token='%v'", hash)).First(&dbToken); err != nil {
					t.Errorf("Can't find new token (%v)", err)
				}

				if expiry-expectedExpiry > 1 {
					t.Errorf("Unexpected token expiry: %v, expected %v", expiry, expectedExpiry)
				}

				if dbToken.ExpiresAt.Unix()-expectedExpiry > 1 {
					t.Errorf("Unexpected token expiry: %v, expected %v", dbToken.ExpiresAt.Unix(), expectedExpiry)
				}
			}
		})
	}

	resetTables(t) // Pack it in, Pack it out a/k/a "Leave No Trace"
}

func TestGetOrgIDs(t *testing.T) {
	_, user, _ := createUserFixtures(t)

	tests := []struct {
		name string
		user User
		want []int
	}{
		{
			name: "basic",
			user: user,
			want: []int{1, 2},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.user.GetOrgIDs()
			ints := make([]int, len(got))
			for i, id := range got {
				ints[i] = id.(int)
			}

			if !reflect.DeepEqual(ints, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", ints, test.want)
			}
		})
	}
	resetTables(t) // Pack it in, Pack it out a/k/a "Leave No Trace"
}

func createUserFixtures(t *testing.T) (Organizations, User, UserOrganizations) {
	// in case other tests don't clean up
	resetTables(t)

	// Load Organization test fixtures
	orgs := Organizations{
		{
			Name:       "ACME",
			Uuid:       domain.GetUuid(),
			AuthType:   "saml2",
			AuthConfig: "[]",
		},
		{
			Name:       "Starfleet Academy",
			Uuid:       domain.GetUuid(),
			AuthType:   "saml2",
			AuthConfig: "[]",
		},
	}
	for i := range orgs {
		if err := DB.Create(&orgs[i]); err != nil {
			t.Errorf("error creating org %+v ...\n %v \n", orgs[i], err)
			t.FailNow()
		}
	}

	// Load User test fixtures
	user := User{
		Email:     "user@example.com",
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  "Existing User",
		Uuid:      domain.GetUuid(),
	}
	if err := DB.Create(&user); err != nil {
		t.Errorf("could not create test user ... %v", err)
		t.FailNow()
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{
		{
			OrganizationID: orgs[0].ID,
			UserID:         user.ID,
			AuthID:         "existing_user",
			AuthEmail:      "user@example.com",
		},
		{
			OrganizationID: orgs[1].ID,
			UserID:         user.ID,
			AuthID:         "existing_user",
			AuthEmail:      "user@example.com",
		},
	}
	for i := range userOrgs {
		if err := DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	return orgs, user, userOrgs
}

// Ensure multiple access tokens for same organization are allowed (to support multiple tabs/browsers)
func TestUser_CreateAccessToken(t *testing.T) {
	resetTables(t)

	// setup fixtures for test
	user := &User{
		ID:        1,
		Email:     "test@test.com",
		FirstName: "test",
		LastName:  "user",
		Nickname:  "Tester",
		AdminRole: nulls.String{},
		Uuid:      domain.GetUuid(),
	}
	_ = DB.Create(user)

	org := &Organization{
		ID:         1,
		Name:       "testorg",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	_ = DB.Create(org)

	userOrg := &UserOrganization{
		ID:             1,
		OrganizationID: 1,
		UserID:         1,
		Role:           UserOrganizationRoleMember,
		AuthID:         "abc123",
		AuthEmail:      "test@test.com",
		LastLogin:      time.Now(),
	}
	_ = DB.Create(userOrg)

	accessToken1, _, err := user.CreateAccessToken(*org, "abc1234")
	if err != nil {
		t.Errorf("unable to create access token 1: %s", err)
	}
	accessToken2, _, err := user.CreateAccessToken(*org, "1234abc")
	if err != nil {
		t.Errorf("unable to create access token 2: %s", err)
	}

	if accessToken1 == accessToken2 {
		t.Error("got same access tokens after two calls")
	}

	uat := &UserAccessToken{}
	count, err := DB.Where("user_id = ? and user_organization_id = ?", user.ID, userOrg.ID).Count(uat)
	if count != 2 {
		t.Errorf("did not find correct number of user access tokens, want 2, got %v", count)
	}

}

func TestGetOrganizations(t *testing.T) {
	orgs, user, _ := createUserFixtures(t)

	tests := []struct {
		name string
		user User
		want []string
	}{
		{
			name: "basic",
			user: user,
			want: []string{orgs[0].Name, orgs[1].Name},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.user.GetOrganizations()
			if err != nil {
				t.Errorf("GetOrganizations() returned error: %s", err)
			}

			orgNames := make([]string, len(got))
			for i, o := range got {
				orgNames[i] = o.Name
			}
			if !reflect.DeepEqual(orgNames, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", got, test.want)
			}
		})
	}
	resetTables(t) // Pack it in, Pack it out a/k/a "Leave No Trace"
}
