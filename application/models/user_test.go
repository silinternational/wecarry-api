package models

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/auth"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestUser_FindOrCreateFromAuthUser() {
	t := ms.T()

	unique := domain.GetUuid().String()

	// create org for test
	org := &Organization{
		Name:       "TestOrg1-" + unique,
		Url:        nulls.String{},
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	createFixture(t, org)

	type args struct {
		tx       *pop.Connection
		orgID    int
		authUser *auth.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create new user: test_user1",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     fmt.Sprintf("test_user1-%s@domain.com", unique),
					UserID:    fmt.Sprintf("test_user1-%s", unique),
				},
			},
			wantErr: false,
		},
		{
			name: "find existing user: test_user1",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     fmt.Sprintf("test_user1-%s@domain.com", unique),
					UserID:    fmt.Sprintf("test_user1-%s", unique),
				},
			},
			wantErr: false,
		},
		{
			name: "conflicting user",
			args: args{
				orgID: org.ID,
				authUser: &auth.User{
					FirstName: "Test",
					LastName:  "User",
					Email:     fmt.Sprintf("test_user1-%s@domain.com", unique),
					UserID:    "test_user2",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{}
			err := u.FindOrCreateFromAuthUser(tt.args.orgID, tt.args.authUser)
			if tt.wantErr {
				ms.Error(err, "FindOrCreateFromAuthUser() did not return expected error")
			} else {
				ms.NoError(err, "FindOrCreateFromAuthUser() error: %s", err)
				ms.NotEqual(0, u.ID, "Did not get a new user ID")
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
		{
			name: "good photoURL",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
				PhotoURL:  nulls.NewString("http://example.com/user/7/avatar"),
			},
			wantErr: false,
		},
		{
			name: "blank photoURL",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
				PhotoURL:  nulls.NewString(""),
			},
			wantErr: false,
		},
		{
			name: "bad photoURL",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				Uuid:      domain.GetUuid(),
				PhotoURL:  nulls.NewString("badone"),
			},
			wantErr:  true,
			errField: "photo_url",
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

	orgs, users, _ := CreateUserFixtures(ms, t)

	type args struct {
		user     *User
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
				user:     &users[0],
				clientID: "abc123",
			},
			wantErr: false,
		},
		{
			name: "123abc",
			args: args{
				user:     &users[0],
				clientID: "123abc",
			},
			wantErr: false,
		},
		{
			name: "empty client ID",
			args: args{
				user:     &users[0],
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
	orgs, users, _ := CreateUserFixtures(ms, t)

	tests := []struct {
		name string
		user User
		want []int
	}{
		{
			name: "basic",
			user: users[0],
			want: []int{orgs[0].ID, orgs[1].ID},
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

func CreateUserFixtures(ms *ModelSuite, t *testing.T) ([]Organization, Users, UserOrganizations) {

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
			AuthType:   AuthTypeSaml,
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
		{
			Email:     fmt.Sprintf("not_participating-%s@example.com", unique),
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  fmt.Sprintf("Not Participating %s", unique),
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
			UserID:         users[0].ID,
			AuthID:         users[0].Email,
			AuthEmail:      users[0].Email,
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

func (ms *ModelSuite) TestGetOrganizations() {
	t := ms.T()
	orgs, users, _ := CreateUserFixtures(ms, t)

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
				t.Errorf("GetOrganizations() = \"%v\", want \"%v\"", orgNames, test.want)
			}
		})
	}
}

func (ms *ModelSuite) Test_FindUserOrganization() {
	t := ms.T()
	createUserOrganizationFixtures(ms, t)

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
	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

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

	orgFixtures := []Organization{
		{
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
		{
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
	}
	for i := range orgFixtures {
		createFixture(t, &orgFixtures[i])
	}

	user := User{
		Email:     "test@com.com",
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		AdminRole: nulls.String{},
		Uuid:      domain.GetUuid(),
	}
	createFixture(t, &user)

	userOrgFixtures := []UserOrganization{
		{
			OrganizationID: orgFixtures[0].ID,
			UserID:         user.ID,
			Role:           UserOrganizationRoleAdmin,
			AuthID:         "abc123",
			AuthEmail:      "test@com.com",
		},
		{
			OrganizationID: orgFixtures[1].ID,
			UserID:         user.ID,
			Role:           UserOrganizationRoleUser,
			AuthID:         "123abc",
			AuthEmail:      "test@com.com",
		},
	}
	for i := range userOrgFixtures {
		createFixture(t, &userOrgFixtures[i])
	}

	if !user.CanEditOrganization(orgFixtures[0].ID) {
		t.Error("user unable to edit org that they should be able to edit")
	}
	if user.CanEditOrganization(orgFixtures[1].ID) {
		t.Error("user is able to edit org that they should not be able to edit")
	}
}

func (ms *ModelSuite) TestUser_AttachPhoto() {
	t := ms.T()

	user := User{}
	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("failed to create user fixture, %s", err)
	}

	var photoFixture File
	const filename = "photo.gif"
	if err := photoFixture.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

	if attachedFile, err := user.AttachPhoto(photoFixture.UUID.String()); err != nil {
		t.Errorf("failed to attach photo to user, %s", err)
	} else {
		ms.Equal(filename, attachedFile.Name)
		ms.NotEqual(0, attachedFile.ID)
		ms.NotEqual(domain.EmptyUUID, attachedFile.UUID.String())
	}

	if err := DB.Load(&user); err != nil {
		t.Errorf("failed to load photo relation for test user, %s", err)
	}

	ms.Equal(filename, user.PhotoFile.Name)

	if got, err := user.GetPhotoURL(); err == nil {
		ms.Regexp("^https?", got)
	} else {
		ms.Fail("user.GetPhotoURL failed, %s", err)
	}
}

func CreateUserFixturesForNicknames(ms *ModelSuite, t *testing.T) User {
	prefix := allPrefixes()[0]

	// Load User test fixtures
	user := User{
		Email:     fmt.Sprintf("user1-%s@example.com", t.Name()),
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  prefix + "ExistingU",
		Uuid:      domain.GetUuid(),
	}

	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("could not create test user %v ... %v", user, err)
		t.FailNow()
	}

	return user
}

func (ms *ModelSuite) TestUniquifyNickname() {
	t := ms.T()
	existingUser := CreateUserFixturesForNicknames(ms, t)
	prefix := allPrefixes()[0]

	tests := []struct {
		name     string
		user     User
		want     string
		dontWant string
	}{
		{
			name: "No Change, Blank Last Name",
			user: User{FirstName: "New"},
			want: prefix + "New",
		},
		{
			name: "No Change, OK Last Name",
			user: User{FirstName: "New", LastName: "User"},
			want: prefix + "NewU",
		},
		{
			name: "Expect Change",
			user: User{
				FirstName: existingUser.FirstName,
				LastName:  existingUser.LastName,
				Nickname:  existingUser.Nickname[len(prefix):], //remove the prefix so it can be added back on
			},
			dontWant: existingUser.Nickname,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.user.uniquifyNickname()
			if err != nil {
				t.Errorf("uniquifyNickname() returned error: %s", err)
			}

			got := test.user.Nickname

			if test.want != "" {
				ms.Equal(test.want, got)
				return
			}

			ms.NotEqual(test.dontWant, got)
		})
	}
}

func (ms *ModelSuite) TestUser_SetLocation() {
	t := ms.T()

	user := User{Uuid: domain.GetUuid(), Email: t.Name() + "_user@example.com", Nickname: t.Name() + "_User"}
	createFixture(t, &user)

	locationFixtures := Locations{
		{
			Description: "a place",
			Country:     "XY",
			Latitude:    nulls.NewFloat64(1.1),
			Longitude:   nulls.NewFloat64(2.2),
		},
		{
			Description: "another place",
			Country:     "AB",
			Latitude:    nulls.Float64{},
			Longitude:   nulls.Float64{},
		},
	}

	err := user.SetLocation(locationFixtures[0])
	ms.NoError(err, "unexpected error from user.SetLocation()")

	locationFromDB, err := user.GetLocation()
	ms.NoError(err, "unexpected error from user.GetLocation()")

	locationFixtures[0].ID = locationFromDB.ID
	ms.Equal(locationFixtures[0], *locationFromDB, "user location data doesn't match new location")

	err = user.SetLocation(locationFixtures[1])
	ms.NoError(err, "unexpected error from user.SetLocation()")

	locationFromDB, err = user.GetLocation()
	ms.NoError(err, "unexpected error from user.GetLocation()")
	ms.Equal(locationFixtures[0].ID, locationFromDB.ID,
		"Location ID doesn't match -- location record was probably not reused")

	locationFixtures[1].ID = locationFromDB.ID
	ms.Equal(locationFixtures[1], *locationFromDB, "destination data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)
}
