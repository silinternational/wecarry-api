package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/silinternational/handcarry-api/auth"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/domain"
)

func (ms *ModelSuite) TestUser_FindOrCreateFromAuthUser() {
	t := ms.T()
	ResetTables(t, ms.DB)

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

func (ms *ModelSuite) TestUser_FindByAccessToken() {
	t := ms.T()
	ResetTables(t, ms.DB)
	_, users, userOrgs := CreateUserFixtures(t)

	// Load access token test fixtures
	tokens := UserAccessTokens{
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("abc123"),
			ExpiresAt:          time.Unix(0, 0),
		},
		{
			UserID:             users[0].ID,
			UserOrganizationID: userOrgs[0].ID,
			AccessToken:        HashClientIdAccessToken("xyz789"),
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
			want: users[0],
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
			var got User
			err := got.FindByAccessToken(test.args.token)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindByAccessToken() returned an error: %v", err)
				} else if got.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", got, test.want)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestValidateUser() {
	t := ms.T()
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

// Ensure multiple access tokens for same organization are allowed (to support multiple tabs/browsers)
func (ms *ModelSuite) TestCreateAccessToken() {
	t := ms.T()
	ResetTables(t, ms.DB)
	orgs, users, _ := CreateUserFixtures(t)

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
			name: "abc123",
			args: args{
				user:     users[0],
				clientID: "abc123",
			},
			wantErr: false,
		},
		{
			name: "123abc",
			args: args{
				user:     users[0],
				clientID: "123abc",
			},
			wantErr: false,
		},
		{
			name: "empty client ID",
			args: args{
				user:     users[0],
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
				hash := HashClientIdAccessToken(test.args.clientID + token)

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

	uat := &UserAccessToken{}
	count, _ := DB.Where("user_id = ?", users[0].ID).Count(uat)
	if count != 2 {
		t.Errorf("did not find correct number of user access tokens, want 2, got %v", count)
	}
}

func (ms *ModelSuite) TestGetOrgIDs() {
	t := ms.T()
	ResetTables(t, ms.DB)
	_, users, _ := CreateUserFixtures(t)

	tests := []struct {
		name string
		user User
		want []int
	}{
		{
			name: "basic",
			user: users[0],
			want: []int{1, 2},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.user.GetOrgIDs()

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", got, test.want)
			}
		})
	}
}

func CreateUserFixtures(t *testing.T) ([]Organization, Users, UserOrganizations) {
	ResetTables(t, DB)

	// Load Organization test fixtures
	orgs := []Organization{
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
	for _, org := range orgs {
		if err := DB.Create(&org); err != nil {
			t.Errorf("error creating org %+v ...\n %v \n", org, err)
			t.FailNow()
		}
	}

	// Load User test fixtures
	users := Users{
		{
			Email:     "user1@example.com",
			FirstName: "Existing",
			LastName:  "User",
			Nickname:  "Existing User",
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Another",
			LastName:  "User",
			Nickname:  "Another User",
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "not_participating@example.com",
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  "Not Participating",
			Uuid:      domain.GetUuid(),
		},
	}
	for _, user := range users {
		if err := DB.Create(&user); err != nil {
			t.Errorf("could not create test user %v ... %v", user, err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{
		{
			OrganizationID: orgs[0].ID,
			UserID:         users[0].ID,
			AuthID:         "existing_user",
			AuthEmail:      "user@example.com",
		},
		{
			OrganizationID: orgs[1].ID,
			UserID:         users[0].ID,
			AuthID:         "existing_user",
			AuthEmail:      "user@example.com",
		},
	}
	for _, uo := range userOrgs {
		if err := DB.Create(&uo); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	return orgs, users, userOrgs
}

func (ms *ModelSuite) TestGetOrganizations() {
	t := ms.T()
	ResetTables(t, ms.DB)
	orgs, users, _ := CreateUserFixtures(t)

	tests := []struct {
		name string
		user User
		want []string
	}{
		{
			name: "basic",
			user: users[0],
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
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", orgNames, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestFindUserOrganization() {
	t := ms.T()
	resetTables(t)
	createUserOrganizationFixtures(t)

	type args struct {
		user User
		org  Organization
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "user 1, org 1",
			args: args{
				user: User{Email: "single@domain.com"},
				org:  Organization{Name: "Org1"},
			},
			wantErr: false,
		},
		{
			name: "user 2, org 1",
			args: args{
				user: User{Email: "two@domain.com"},
				org:  Organization{Name: "Org1"},
			},
			wantErr: false,
		},
		{
			name: "user 2, org 2",
			args: args{
				user: User{Email: "two@domain.com"},
				org:  Organization{Name: "Org2"},
			},
			wantErr: false,
		},
		{
			name: "user 1, org 2",
			args: args{
				user: User{Email: "single@domain.com"},
				org:  Organization{Name: "Org2"},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var user User
			if err := DB.Where("email = ?", test.args.user.Email).First(&user); err != nil {
				t.Errorf("couldn't find test user '%v'", test.args.user.Email)
			}

			var org Organization
			if err := DB.Where("name = ?", test.args.org.Name).First(&org); err != nil {
				t.Errorf("couldn't find test org '%v'", test.args.org.Name)
			}

			uo, err := user.FindUserOrganization(org)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindOrgByUUID() returned an error: %v", err)
				} else if (uo.UserID != user.ID) || (uo.OrganizationID != org.ID) {
					t.Errorf("received wrong UserOrganization (UserID=%v, OrganizationID=%v), expected (user.ID=%v, org.ID=%v)",
						uo.UserID, uo.OrganizationID, user.ID, org.ID)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestUser_GetPosts() {
	t := ms.T()
	resetTables(t)
	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

	type args struct {
		user     User
		postRole string
	}
	tests := []struct {
		name string
		args args
		want []uuid.UUID
	}{
		{
			name: "created by",
			args: args{
				user:     users[0],
				postRole: PostRoleCreatedby,
			},
			want: []uuid.UUID{posts[0].Uuid, posts[1].Uuid},
		},
		{
			name: "providing by",
			args: args{
				user:     users[1],
				postRole: PostRoleProviding,
			},
			want: []uuid.UUID{posts[0].Uuid},
		},
		{
			name: "receiving by",
			args: args{
				user:     users[1],
				postRole: PostRoleReceiving,
			},
			want: []uuid.UUID{posts[1].Uuid},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.args.user.GetPosts(test.args.postRole)
			if err != nil {
				t.Errorf("GetPosts() returned error: %s", err)
			}

			ids := make([]uuid.UUID, len(got))
			for i, p := range got {
				ids[i] = p.Uuid
			}
			if !reflect.DeepEqual(ids, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", ids, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestCanEditOrganization() {
	t := ms.T()
	ResetTables(t, ms.DB)

	orgFixtures := []Organization{
		{
			ID:         1,
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   "saml2",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
		{
			ID:         2,
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   "saml2",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
	}
	for _, of := range orgFixtures {
		err := DB.Create(&of)
		if err != nil {
			t.Errorf("failed to create org fixtures: %s", err)
		}
	}

	user := User{
		ID:        1,
		Email:     "test@com.com",
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		AdminRole: nulls.String{},
		Uuid:      domain.GetUuid(),
	}
	err := DB.Create(&user)
	if err != nil {
		t.Errorf("failed to create user fixture: %s", err)
	}

	userOrgFixtures := []UserOrganization{
		{
			ID:             1,
			OrganizationID: 1,
			UserID:         1,
			Role:           UserOrganizationRoleAdmin,
			AuthID:         "abc123",
			AuthEmail:      "test@com.com",
		},
		{
			ID:             2,
			OrganizationID: 2,
			UserID:         1,
			Role:           UserOrganizationRoleMember,
			AuthID:         "123abc",
			AuthEmail:      "test@com.com",
		},
	}
	for _, uof := range userOrgFixtures {
		err := DB.Create(&uof)
		if err != nil {
			t.Errorf("failed to create user org fixtures: %s", err)
		}
	}

	if !user.CanEditOrganization(orgFixtures[0].ID) {
		t.Error("user unable to edit org that they should be able to edit")
	}
	if user.CanEditOrganization(orgFixtures[1].ID) {
		t.Error("user is able to edit org that they should not be able to edit")
	}
}
