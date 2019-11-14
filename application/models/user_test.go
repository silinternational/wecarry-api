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
	createFixture(ms, org)

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
				if err := ms.DB.Where(fmt.Sprintf("access_token='%v'", hash)).First(&dbToken); err != nil {
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
	count, _ := ms.DB.Where("user_id = ?", users[0].ID).Count(uat)
	if count != 2 {
		t.Errorf("did not find correct number of user access tokens, want 2, got %v", count)
	}
}

func (ms *ModelSuite) TestUser_GetOrgIDs() {
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
			want: []int{orgs[1].ID, orgs[0].ID},
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

func (ms *ModelSuite) TestUser_GetOrganizations() {
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
			want: []string{orgs[1].Name, orgs[0].Name},
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
			if err := ms.DB.Where("email = ?", test.args.user.Email).First(&user); err != nil {
				t.Errorf("couldn't find test user '%v'", test.args.user.Email)
			}

			var org Organization
			if err := ms.DB.Where("name = ?", test.args.org.Name).First(&org); err != nil {
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
	f := CreateFixturesForUserGetPosts(ms)

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
				user:     f.Users[0],
				postRole: PostRoleCreatedby,
			},
			want: []uuid.UUID{f.Posts[3].Uuid, f.Posts[2].Uuid, f.Posts[1].Uuid, f.Posts[0].Uuid},
		},
		{
			name: "providing by",
			args: args{
				user:     f.Users[1],
				postRole: PostRoleProviding,
			},
			want: []uuid.UUID{f.Posts[1].Uuid, f.Posts[0].Uuid},
		},
		{
			name: "receiving by",
			args: args{
				user:     f.Users[1],
				postRole: PostRoleReceiving,
			},
			want: []uuid.UUID{f.Posts[3].Uuid, f.Posts[2].Uuid},
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

func (ms *ModelSuite) TestUser_CanEditOrganization() {
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
		createFixture(ms, &orgFixtures[i])
	}

	user := User{
		Email:     "test@com.com",
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		AdminRole: nulls.String{},
		Uuid:      domain.GetUuid(),
	}
	createFixture(ms, &user)

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
		createFixture(ms, &userOrgFixtures[i])
	}

	if !user.CanEditOrganization(orgFixtures[0].ID) {
		t.Error("user unable to edit org that they should be able to edit")
	}
	if user.CanEditOrganization(orgFixtures[1].ID) {
		t.Error("user is able to edit org that they should not be able to edit")
	}
}

func (ms *ModelSuite) TestUser_CanEditAllPosts() {
	t := ms.T()
	f := CreateUserFixtures_CanEditAllPosts(ms)

	tests := []struct {
		name string
		user User
		want bool
	}{
		{name: "super admin & org admin", user: f.Users[0], want: true},
		{name: "sales admin & org admin", user: f.Users[1], want: false},
		{name: "org admin", user: f.Users[2], want: false},
		{name: "super admin", user: f.Users[3], want: true},
		{name: "sales admin", user: f.Users[4], want: false},
		{name: "user", user: f.Users[5], want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms.Equal(test.want, test.user.CanEditAllPosts(), "CanEditAllPosts() incorrect result")
		})
	}
}

func (ms *ModelSuite) TestUser_CanUpdatePostStatus() {
	t := ms.T()

	tests := []struct {
		name      string
		post      Post
		user      User
		newStatus string
		want      bool
	}{
		{
			name: "Creator",
			post: Post{CreatedByID: 1},
			user: User{ID: 1},
			want: true,
		},
		{
			name: "SuperDuperAdmin",
			post: Post{},
			user: User{AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin)},
			want: true,
		},
		{
			name:      "Open",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusOpen,
			want:      false,
		},
		{
			name:      "Committed",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusCommitted,
			want:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms.Equal(test.want, test.user.CanUpdatePostStatus(test.post, test.newStatus),
				"CanEditAllPosts() incorrect result")
		})
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

	if err := ms.DB.Load(&user); err != nil {
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
	createFixture(ms, &user)

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
	ms.Equal(locationFixtures[1], *locationFromDB, "user location data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)
}

func (ms *ModelSuite) TestUser_UnreadMessageCount() {
	t := ms.T()

	f := CreateUserFixtures_UnreadMessageCount(ms, t)

	tests := []struct {
		name      string
		user      User
		want      int
		wantErr   bool
		wantTotal int
	}{
		//{
		//	name: "Eager User",
		//	user: f.Users[0],
		//	want: 0,
		//},
		{
			name:      "Lazy User",
			user:      f.Users[1],
			want:      2,
			wantTotal: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			got, err := test.user.UnreadMessageCount()
			if test.wantErr {
				ms.Error(err, "did not get expected error")
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, len(got))

			gotTotal := 0
			for _, g := range got {
				gotTotal += g.Count
			}
			ms.Equal(test.wantTotal, gotTotal)
		})
	}
}

func (ms *ModelSuite) TestUser_GetThreads() {
	t := ms.T()

	f := CreateUserFixtures_GetThreads(ms)

	tests := []struct {
		name string
		user User
		want []uuid.UUID
	}{
		{name: "no threads", user: f.Users[1], want: []uuid.UUID{}},
		{name: "two threads", user: f.Users[0], want: []uuid.UUID{f.Threads[1].Uuid, f.Threads[0].Uuid}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.user.GetThreads()
			ms.NoError(err)

			ids := make([]uuid.UUID, len(got))
			for i := range got {
				ids[i] = got[i].Uuid
			}
			ms.Equal(test.want, ids, "incorrect list of threads returned")
		})
	}
}
