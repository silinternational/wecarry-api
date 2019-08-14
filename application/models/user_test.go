package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/auth/saml"
	"github.com/silinternational/handcarry-api/domain"
)

func TestUser_FindOrCreateFromSamlUser(t *testing.T) {
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
		samlUser saml.SamlUser
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
				tx:    models.DB,
				orgID: org.ID,
				samlUser: saml.SamlUser{
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
				tx:    models.DB,
				orgID: org.ID,
				samlUser: saml.SamlUser{
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
				tx:    models.DB,
				orgID: org.ID,
				samlUser: saml.SamlUser{
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
			err := u.FindOrCreateFromSamlUser(tt.args.tx, tt.args.orgID, tt.args.samlUser)
			if err != nil && !tt.wantErr {
				t.Errorf("FindOrCreateFromSamlUser() error = %v, wantErr %v", err, tt.wantErr)
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
		Email:      "user@example.com",
		FirstName:  "Existing",
		LastName:   "User",
		Nickname:   "Existing User",
		AuthOrgID:  org.ID,
		AuthOrgUid: "existing_user",
		Uuid:       domain.GetUuid(),
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
				Email:      "user@example.com",
				FirstName:  "A",
				LastName:   "User",
				Nickname:   "A User",
				AuthOrgID:  1,
				AuthOrgUid: "a_user",
				Uuid:       domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing email",
			user: User{
				FirstName:  "A",
				LastName:   "User",
				Nickname:   "A User",
				AuthOrgID:  1,
				AuthOrgUid: "a_user",
				Uuid:       domain.GetUuid(),
			},
			wantErr:  true,
			errField: "email",
		},
		{
			name: "missing first_name",
			user: User{
				Email:      "user@example.com",
				LastName:   "User",
				Nickname:   "A User",
				AuthOrgID:  1,
				AuthOrgUid: "a_user",
				Uuid:       domain.GetUuid(),
			},
			wantErr:  true,
			errField: "first_name",
		},
		{
			name: "missing last_name",
			user: User{
				Email:      "user@example.com",
				FirstName:  "A",
				Nickname:   "A User",
				AuthOrgID:  1,
				AuthOrgUid: "a_user",
				Uuid:       domain.GetUuid(),
			},
			wantErr:  true,
			errField: "last_name",
		},
		{
			name: "missing nickname",
			user: User{
				Email:      "user@example.com",
				FirstName:  "A",
				LastName:   "User",
				AuthOrgID:  1,
				AuthOrgUid: "a_user",
				Uuid:       domain.GetUuid(),
			},
			wantErr:  true,
			errField: "nickname",
		},
		{
			name: "missing auth_org_id",
			user: User{
				Email:      "user@example.com",
				FirstName:  "A",
				LastName:   "User",
				Nickname:   "A User",
				AuthOrgUid: "a_user",
				Uuid:       domain.GetUuid(),
			},
			wantErr:  true,
			errField: "auth_org_id",
		},
		{
			name: "missing auth_org_uid",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				AuthOrgID: 1,
				Uuid:      domain.GetUuid(),
			},
			wantErr:  true,
			errField: "auth_org_uid",
		},
		{
			name: "missing uuid",
			user: User{
				Email:      "user@example.com",
				FirstName:  "A",
				LastName:   "User",
				Nickname:   "A User",
				AuthOrgID:  1,
				AuthOrgUid: "a_user",
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
