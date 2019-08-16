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
	resetTables(t) // in case other tests don't clean up

	// Load Organization test fixtures
	org := Organization{
		Name:       "ACME",
		Uuid:       domain.GetUuid(),
		AuthType:   "saml2",
		AuthConfig: "[]",
	}
	if err := DB.Create(&org); err != nil {
		t.Errorf("could not create test org ... %v", err)
		t.FailNow()
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

	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:      user.ID,
			AccessToken: hashClientIdAccessToken("abc123"),
			ExpiresAt:   time.Unix(0, 0),
		},
		{
			UserID:      user.ID,
			AccessToken: hashClientIdAccessToken("xyz789"),
			ExpiresAt:   time.Date(2099, time.December, 31, 0, 0, 0, 0, time.UTC),
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
	resetTables(t)

	// Load Organization test fixtures
	org := Organization{
		Name:       "ACME",
		Uuid:       domain.GetUuid(),
		AuthType:   "saml2",
		AuthConfig: "[]",
	}
	if err := DB.Create(&org); err != nil {
		t.Errorf("could not create test org ... %v", err)
		t.FailNow()
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

	type args struct {
		user     User
		clientID string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "basic",
			args: args{
				user:     user,
				clientID: "abc123",
			},
		},
		{
			name: "empty client ID",
			args: args{
				user:     user,
				clientID: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedExpiry := createAccessTokenExpiry().Unix()
			token, expiry, err := test.args.user.CreateAccessToken(org, test.args.clientID)
			if err != nil {
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
		})
	}

	resetTables(t) // Pack it in, Pack it out a/k/a "Leave No Trace"
}

func TestGetOrgIDs(t *testing.T) {
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

	for i := range orgs {
		uo := UserOrganization{
			OrganizationID: orgs[i].ID,
			UserID:         user.ID,
			Role:           "user",
		}
		if err := DB.Create(&uo); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

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
			if err := DB.Load(&test.user); err != nil {
				t.Errorf("Failed to load related records on user %v", user.ID)
				t.FailNow()
			}
			got := test.user.GetOrgIDs()

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", got, test.want)
			}
		})
	}
	resetTables(t) // Pack it in, Pack it out a/k/a "Leave No Trace"
}
