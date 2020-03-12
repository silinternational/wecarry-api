package models

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestUser_FindOrCreateFromAuthUser() {
	t := ms.T()

	unique := domain.GetUUID().String()

	// create org for test
	org := &Organization{
		Name:       "TestOrg1-" + unique,
		Url:        nulls.String{},
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		UUID:       domain.GetUUID(),
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
				ms.True(u.ID != 0, "Did not get a new user ID")
			}
		})
	}
}

func (ms *ModelSuite) TestUser_FindOrCreateFromOrglessAuthUser() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)

	unique := domain.GetUUID().String()

	tests := []struct {
		name     string
		authType string
		authUser *auth.User
	}{
		{
			name:     "create new user: test_user1",
			authType: AuthTypeAzureAD.String(),
			authUser: &auth.User{
				FirstName: "Test",
				LastName:  "User",
				Email:     fmt.Sprintf("test_user1-%s@domain.com", unique),
				UserID:    fmt.Sprintf("test_user1-%s", unique),
			},
		},
		{
			name:     "find existing user: test_user1",
			authType: AuthTypeAzureAD.String(),
			authUser: &auth.User{
				FirstName: "Test",
				LastName:  "User",
				Email:     fmt.Sprintf("test_user1-%s@domain.com", unique),
				UserID:    fmt.Sprintf("test_user1-%s", unique),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{}
			err := u.FindOrCreateFromOrglessAuthUser(tc.authUser, tc.authType)
			ms.NoError(err, "unexpected error")
			ms.True(u.ID != 0, "Did not get a new user ID")
			ms.Equal(tc.authType, u.SocialAuthProvider.String, "incorrect SocialAuthProvider.")
		})
	}
}

func (ms *ModelSuite) TestUser_Validate() {
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
				UUID:      domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "missing email",
			user: User{
				FirstName: "A",
				LastName:  "User",
				Nickname:  "A User",
				UUID:      domain.GetUUID(),
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
				UUID:     domain.GetUUID(),
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
				UUID:      domain.GetUUID(),
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
				UUID:      domain.GetUUID(),
			},
			wantErr:  true,
			errField: "nickname",
		},
		{
			name: "invisible nickname",
			user: User{
				Email:     "user@example.com",
				FirstName: "A",
				LastName:  "User",
				UUID:      domain.GetUUID(),
				Nickname:  string([]rune{0xfe0f}),
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
				Email:        "user@example.com",
				FirstName:    "A",
				LastName:     "User",
				Nickname:     "A User",
				UUID:         domain.GetUUID(),
				AuthPhotoURL: nulls.NewString("http://example.com/user/7/avatar"),
			},
			wantErr: false,
		},
		{
			name: "blank photoURL",
			user: User{
				Email:        "user@example.com",
				FirstName:    "A",
				LastName:     "User",
				Nickname:     "A User",
				UUID:         domain.GetUUID(),
				AuthPhotoURL: nulls.NewString(""),
			},
			wantErr: false,
		},
		{
			name: "bad authPhotoURL",
			user: User{
				Email:        "user@example.com",
				FirstName:    "A",
				LastName:     "User",
				Nickname:     "A User",
				UUID:         domain.GetUUID(),
				AuthPhotoURL: nulls.NewString("badone"),
			},
			wantErr:  true,
			errField: "auth_photo_url",
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
func (ms *ModelSuite) TestUser_CreateAccessToken() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 1)
	users := uf.Users

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
			token, expiry, err := test.args.user.CreateAccessToken(uf.Organization, test.args.clientID)
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
	ms.Equal(3, count, "did not find correct number of user access tokens")
}

func (ms *ModelSuite) TestUser_CreateOrglessAccessToken() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 1)
	users := uf.Users

	tests := []struct {
		name     string
		user     *User
		clientID string
		wantErr  bool
	}{
		{
			name:     "abc123",
			user:     &users[0],
			clientID: "abc123",
			wantErr:  false,
		},
		{
			name:     "123abc",
			user:     &users[0],
			clientID: "123abc",
			wantErr:  false,
		},
		{
			name:     "empty client ID",
			user:     &users[0],
			clientID: "",
			wantErr:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expectedExpiry := createAccessTokenExpiry().Unix()
			token, expiry, err := tc.user.CreateOrglessAccessToken(tc.clientID)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, but did not get one")
				}
			} else {
				if err != nil && !tc.wantErr {
					t.Errorf("CreateAccessToken() returned error: %v", err)
				}
				hash := HashClientIdAccessToken(tc.clientID + token)

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
	ms.Equal(3, count, "did not find correct number of user access tokens")
}

func (ms *ModelSuite) TestUser_GetOrgIDs() {
	t := ms.T()

	orgs, users := createFixturesForUserGetOrganizations(ms)

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

	orgs, users := createFixturesForUserGetOrganizations(ms)

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

func (ms *ModelSuite) TestUser_FindUserOrganization() {
	t := ms.T()
	users, orgs := createUserOrganizationFixtures(ms)

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
				user: users[0],
				org:  orgs[0],
			},
			wantErr: false,
		},
		{
			name: "user 2, org 1",
			args: args{
				user: users[1],
				org:  orgs[0],
			},
			wantErr: false,
		},
		{
			name: "user 1, org 2",
			args: args{
				user: users[0],
				org:  orgs[1],
			},
			wantErr: false,
		},
		{
			name: "user 2, org 2",
			args: args{
				user: users[1],
				org:  orgs[1],
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user := test.args.user
			org := test.args.org
			uo, err := user.FindUserOrganization(org)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one, %v", uo.ID)
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
				postRole: PostsCreated,
			},
			want: []uuid.UUID{f.Posts[3].UUID, f.Posts[2].UUID, f.Posts[1].UUID, f.Posts[0].UUID},
		},
		{
			name: "providing",
			args: args{
				user:     f.Users[1],
				postRole: PostsProviding,
			},
			want: []uuid.UUID{f.Posts[1].UUID, f.Posts[0].UUID},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.args.user.Posts(test.args.postRole)
			if err != nil {
				t.Errorf("Posts() returned error: %s", err)
			}

			ids := make([]uuid.UUID, len(got))
			for i, p := range got {
				ids[i] = p.UUID
			}
			if !reflect.DeepEqual(ids, test.want) {
				t.Errorf("GetOrgIDs() = \"%v\", want \"%v\"", ids, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestUser_CanCreateOrganization() {
	t := ms.T()

	user := User{AdminRole: UserAdminRoleUser}
	admin := User{AdminRole: UserAdminRoleAdmin}
	salesAdmin := User{AdminRole: UserAdminRoleSalesAdmin}
	superAdmin := User{AdminRole: UserAdminRoleSuperAdmin}

	if !salesAdmin.CanCreateOrganization() {
		t.Error("sales admin should be able to create orgs")
	}
	if !superAdmin.CanCreateOrganization() {
		t.Error("super admin should be able to create orgs")
	}
	if admin.CanCreateOrganization() {
		t.Error("admin should not be able to create orgs")
	}
	if user.CanCreateOrganization() {
		t.Error("user should not be able to create orgs")
	}
}

func (ms *ModelSuite) TestUser_CanEditOrganization() {
	t := ms.T()

	orgFixtures := createOrganizationFixtures(ms.DB, 2)

	user := User{
		Email:     "test@com.com",
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		AdminRole: UserAdminRoleUser,
		UUID:      domain.GetUUID(),
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
			ms.Equal(test.want, test.user.canEditAllPosts(), "canEditAllPosts() incorrect result")
		})
	}
}

func (ms *ModelSuite) TestUser_CanUpdatePostStatus() {
	t := ms.T()

	tests := []struct {
		name      string
		post      Post
		user      User
		newStatus PostStatus
		want      bool
	}{
		{
			name: "Creator",
			post: Post{CreatedByID: 1},
			user: User{ID: 1},
			want: true,
		},
		{
			name: "SuperAdmin",
			post: Post{},
			user: User{AdminRole: UserAdminRoleSuperAdmin},
			want: true,
		},
		{
			name:      "Open",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusOpen,
			want:      false,
		},
		{
			name:      "Open to Accepted",
			post:      Post{CreatedByID: 1, Status: PostStatusOpen},
			user:      User{ID: 1},
			newStatus: PostStatusAccepted,
			want:      true,
		},
		{
			name:      "Accepted to Accepted",
			post:      Post{CreatedByID: 1, Status: PostStatusAccepted},
			newStatus: PostStatusAccepted,
			want:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms.Equal(test.want, test.user.CanUpdatePostStatus(test.post, test.newStatus),
				"incorrect result")
		})
	}
}

func (ms *ModelSuite) TestUser_CanViewPost() {
	t := ms.T()

	f := CreateFixturesForUserCanViewPost(ms)
	users := f.Users
	posts := f.Posts

	tests := []struct {
		name      string
		post      Post
		user      User
		newStatus PostStatus
		want      bool
	}{
		{
			name: "Creator",
			post: posts[0],
			user: users[0],
			want: true,
		},
		{
			name: "SuperAdmin",
			post: posts[0],
			user: users[3],
			want: true,
		},
		{
			name: "User's Org",
			post: posts[0],
			user: users[1],
			want: true,
		},
		{
			name: "All with User's untrusted Org",
			post: posts[2],
			user: users[0],
			want: true,
		},
		{
			name: "Trusted with User's trusted Org",
			post: posts[3],
			user: users[1],
			want: true,
		},
		{
			name: "Trusted with User's untrusted Org",
			post: posts[3],
			user: users[0],
			want: false,
		},
		{
			name: "Same with User's untrusted Org",
			post: posts[4],
			user: users[0],
			want: false,
		},
		{
			name: "Same with User's trusted Org",
			post: posts[4],
			user: users[1],
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms.Equal(test.want, test.user.canViewPost(test.post),
				"incorrect result")
		})
	}
}

func (ms *ModelSuite) TestUser_CanViewOrganization() {
	t := ms.T()

	orgFixtures := createOrganizationFixtures(ms.DB, 2)

	user := User{
		Email:     "test@com.com",
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		AdminRole: UserAdminRoleUser,
		UUID:      domain.GetUUID(),
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

	if !user.CanViewOrganization(orgFixtures[0].ID) {
		t.Error("user unable to view org that they should be able to view")
	}
	if user.CanViewOrganization(orgFixtures[1].ID) {
		t.Error("user is able to view org that they should not be able to view")
	}
}

func (ms *ModelSuite) TestUser_FindByUUID() {
	t := ms.T()

	f := createUserFixtures(ms.DB, 1)

	tests := []struct {
		name    string
		UUID    string
		wantErr string
	}{
		{
			name:    "Good",
			UUID:    f.Users[0].UUID.String(),
			wantErr: "",
		},
		{
			name:    "Bad",
			UUID:    "",
			wantErr: "uuid must not be blank",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u User
			err := u.FindByUUID(test.UUID)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.Equal(test.UUID, u.UUID.String())
		})
	}
}

func (ms *ModelSuite) TestUser_FindByID() {
	t := ms.T()

	f := createUserFixtures(ms.DB, 1)

	tests := []struct {
		name    string
		ID      int
		wantErr string
	}{
		{
			name:    "Good",
			ID:      f.Users[0].ID,
			wantErr: "",
		},
		{
			name:    "Bad",
			ID:      0,
			wantErr: "id must be a positive number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u User
			err := u.FindByID(test.ID)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}
			ms.Equal(test.ID, u.ID)
		})
	}
}

func (ms *ModelSuite) TestUser_FindByEmail() {
	t := ms.T()

	f := createUserFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		email   string
		wantErr string
	}{
		{
			name:    "Good",
			email:   f.Users[0].Email,
			wantErr: "",
		},
		{
			name:    "Bad",
			email:   "bad@example.com",
			wantErr: sql.ErrNoRows.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u User
			err := u.FindByEmail(tc.email)
			if tc.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tc.wantErr)
				return
			}
			ms.Equal(tc.email, u.Email)
		})
	}
}

func (ms *ModelSuite) TestUser_FindBySocialAuthProvider() {
	t := ms.T()

	f := createUserFixtures(ms.DB, 2)
	user := f.Users[0]
	user.SocialAuthProvider = nulls.NewString(auth.AuthTypeFacebook)
	ms.NoError(user.Save(), "error saving User for test prep.")

	tests := []struct {
		name               string
		email              string
		socialAuthProvider string
		wantErr            string
	}{
		{
			name:               "Good",
			email:              user.Email,
			socialAuthProvider: user.SocialAuthProvider.String,
			wantErr:            "",
		},
		{
			name:               "Bad Email",
			email:              "basd@example.com",
			socialAuthProvider: user.SocialAuthProvider.String,
			wantErr:            "sql: no rows in result set",
		},
		{
			name:               "Bad SocialAuthProvider",
			email:              user.Email,
			socialAuthProvider: "badOne",
			wantErr:            "sql: no rows in result set",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var u User
			err := u.FindByEmailAndSocialAuthProvider(tc.email, tc.socialAuthProvider)
			if tc.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tc.wantErr)
				return
			}
			ms.Equal(tc.email, u.Email)
		})
	}
}

func (ms *ModelSuite) TestUser_AttachPhoto() {
	uf := createUserFixtures(ms.DB, 3)
	users := uf.Users

	files := createFileFixtures(3)
	users[1].PhotoFileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&users[1], "photo_file_id"))

	tests := []struct {
		name     string
		user     User
		oldImage *File
		newImage string
		want     File
		wantErr  string
	}{
		{
			name:     "no previous file",
			user:     users[0],
			newImage: files[1].UUID.String(),
			want:     files[1],
		},
		{
			name:     "previous file",
			user:     users[1],
			oldImage: &files[0],
			newImage: files[2].UUID.String(),
			want:     files[2],
		},
		{
			name:     "bad ID",
			user:     users[2],
			newImage: uuid.UUID{}.String(),
			wantErr:  "no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.user.AttachPhoto(tt.newImage)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(tt.want.UUID.String(), got.UUID.String(), "wrong file returned")
			ms.Equal(true, got.Linked, "new user photo file is not marked as linked")
			if tt.oldImage != nil {
				ms.Equal(false, tt.oldImage.Linked, "old user photo file is not marked as unlinked")
			}
		})
	}
}

func (ms *ModelSuite) TestUser_RemovePhoto() {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	files := createFileFixtures(2)
	users[1].PhotoFileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&users[1], "photo_file_id"))

	tests := []struct {
		name     string
		user     User
		oldImage *File
		want     File
		wantErr  string
	}{
		{
			name: "no file",
			user: users[0],
		},
		{
			name:     "has a file",
			user:     users[1],
			oldImage: &files[0],
		},
		{
			name:    "bad ID",
			user:    User{},
			wantErr: "invalid User ID",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			err := tt.user.RemovePhoto()
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			if tt.oldImage != nil {
				ms.Equal(false, tt.oldImage.Linked, "old user photo file is not marked as unlinked")
			}
		})
	}
}

func (ms *ModelSuite) TestUser_GetPhotoID() {
	t := ms.T()
	f := createFixturesForTestUserGetPhoto(ms)
	photoID2 := f.Users[2].PhotoFile.UUID.String()
	photoID3 := f.Users[3].PhotoFile.UUID.String()

	tests := []struct {
		name    string
		user    User
		wantID  *string
		wantErr string
	}{
		{
			name:   "no AuthPhoto, no photo attachment",
			user:   f.Users[0],
			wantID: nil,
		},
		{
			name:   "AuthPhoto, and no photo attachment",
			user:   f.Users[1],
			wantID: nil,
		},
		{
			name:   "no AuthPhoto, but photo attachment",
			user:   f.Users[2],
			wantID: &photoID2,
		},
		{
			name:   "AuthPhoto and photo attachment",
			user:   f.Users[3],
			wantID: &photoID3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			photoID, err := test.user.GetPhotoID()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.Equal(test.wantID, photoID, "incorrect photo id.")
		})
	}
}

func (ms *ModelSuite) TestUser_GetPhoto() {
	t := ms.T()
	f := createFixturesForTestUserGetPhoto(ms)

	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(f.Users[0].Email))))
	gravatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%x.jpg?s=200&d=mp", hash)

	tests := []struct {
		name    string
		user    User
		wantURL string
		wantErr string
	}{
		{
			name:    "no AuthPhoto, no photo attachment",
			user:    f.Users[0],
			wantURL: gravatarURL,
		},
		{
			name:    "AuthPhoto, and no photo attachment",
			user:    f.Users[1],
			wantURL: f.Users[1].AuthPhotoURL.String,
		},
		{
			name:    "no AuthPhoto, but photo attachment",
			user:    f.Users[2],
			wantURL: f.Users[2].PhotoFile.URL,
		},
		{
			name:    "AuthPhoto and photo attachment",
			user:    f.Users[3],
			wantURL: f.Users[3].PhotoFile.URL,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, err := test.user.GetPhotoURL()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.NotNil(url)
			ms.Equal(test.wantURL, *url)
		})
	}
}

func (ms *ModelSuite) TestUser_Save() {
	t := ms.T()
	f := createFixturesForTestUserSave(ms)

	tests := []struct {
		name    string
		user    User
		wantErr string
	}{
		{
			name:    "no uuid",
			user:    f.Users[0],
			wantErr: "",
		},
		{
			name:    "no uuid, should not conflict with first",
			user:    f.Users[1],
			wantErr: "",
		},
		{
			name:    "uuid given",
			user:    f.Users[2],
			wantErr: "",
		},
		{
			name:    "update existing",
			user:    f.Users[3],
			wantErr: "",
		},
		{
			name:    "validation error",
			user:    f.Users[4],
			wantErr: "first_name: FirstName can not be blank.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.user.Save()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.True(test.user.UUID.Version() != 0)
			var u User
			ms.NoError(u.FindByID(test.user.ID))
		})
	}
}

func (ms *ModelSuite) TestUser_UniquifyNickname() {
	t := ms.T()
	allPrefs := getShuffledPrefixes()
	prefix := allPrefs[0]
	prefix2 := allPrefs[1]
	existingUser := CreateUserFixturesForNicknames(ms, t, prefix)

	testCases := []struct {
		name string
		user User
		want string
	}{
		{
			name: "No Change, Blank Last Name",
			user: User{FirstName: "New"},
			want: prefix + " New",
		},
		{
			name: "No Change, OK Last Name",
			user: User{FirstName: "New", LastName: "User"},
		},
		{
			name: "Expect Change",
			user: User{
				FirstName: existingUser.FirstName,
				LastName:  existingUser.LastName,
				Nickname:  existingUser.Nickname[len(prefix)+1:], //remove the prefix so it can be added back on
			},
			want: prefix2 + " " + existingUser.FirstName + " " + existingUser.LastName[:1],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.uniquifyNickname(allPrefs)
			if err != nil {
				t.Errorf("uniquifyNickname() returned error: %s", err)
			}

			got := tc.user.Nickname

			if tc.want != "" {
				ms.Equal(tc.want, got)
				return
			}
		})
	}
}

func (ms *ModelSuite) TestUser_SetLocation() {
	uf := createUserFixtures(ms.DB, 1)
	user := uf.Users[0]
	user.LocationID = nulls.Int{}
	ms.NoError(ms.DB.Save(&user))

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

	tests := []struct {
		name        string
		newLocation Location
		wantNil     bool
	}{
		{
			name:        "previously unset",
			newLocation: locationFixtures[0],
		},
		{
			name:        "overwrite location with new info",
			newLocation: locationFixtures[1],
		},
	}

	for _, test := range tests {
		ms.T().Run(test.name, func(t *testing.T) {
			err := user.SetLocation(test.newLocation)
			ms.NoError(err, "unexpected error from user.SetLocation()")

			locationFromDB, err := user.GetLocation()
			ms.NoError(err, "unexpected error from user.GetLocation()")

			if test.wantNil {
				ms.Nil(locationFromDB)
			} else {
				test.newLocation.ID = locationFromDB.ID
				ms.Equal(test.newLocation, *locationFromDB, "user location data doesn't match new location")
			}
		})
	}
}

func (ms *ModelSuite) TestUser_RemoveLocation() {
	uf := createUserFixtures(ms.DB, 2)
	uf.Users[0].LocationID = nulls.Int{}
	ms.NoError(ms.DB.Save(&uf.Users[0]))

	tests := []struct {
		name    string
		user    User
		wantNil bool
	}{
		{
			name: "user has no location",
			user: uf.Users[0],
		},
		{
			name: "user has a location",
			user: uf.Users[1],
		},
	}

	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			id := tt.user.LocationID
			fmt.Print(id)
			err := tt.user.RemoveLocation()
			ms.NoError(err, "unexpected error from user.RemoveLocation()")

			locationFromDB, err := tt.user.GetLocation()
			ms.NoError(err, "unexpected error from user.GetLocation()")
			ms.Nil(locationFromDB)

			loc := Location{}
			err = ms.DB.Find(&loc, id)
			ms.Error(err)
			ms.Equal("sql: no rows in result set", err.Error())
		})
	}
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
		{name: "two threads", user: f.Users[0], want: []uuid.UUID{f.Threads[1].UUID, f.Threads[0].UUID}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.user.GetThreads()
			ms.NoError(err)

			ids := make([]uuid.UUID, len(got))
			for i := range got {
				ids[i] = got[i].UUID
			}
			ms.Equal(test.want, ids, "incorrect list of threads returned")
		})
	}
}

func (ms *ModelSuite) TestUser_WantsPostNotification() {
	t := ms.T()
	f := CreateFixturesForUserWantsPostNotification(ms)

	tests := []struct {
		name string
		user User
		post Post
		want bool
	}{
		{
			name: "no, I created it",
			user: f.Users[0],
			post: f.Posts[0],
			want: false,
		},
		{
			name: "no, it's a request for something not near me",
			user: f.Users[1],
			post: f.Posts[0],
			want: false,
		},
		{
			name: "yes, it's a request for something near me",
			user: f.Users[1],
			post: f.Posts[1],
			want: true,
		},
		{
			name: "no, there is no request origin",
			user: f.Users[1],
			post: f.Posts[2],
			want: false,
		},
		{
			name: "yes, I have a watch for that location",
			user: f.Users[2],
			post: f.Posts[2],
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.user.WantsPostNotification(test.post)

			ms.Equal(test.want, got, "incorrect result from WantsPostNotification()")
		})
	}
}

func (ms *ModelSuite) TestUser_UpdateStandardPreferences() {
	t := ms.T()

	f := CreateUserFixtures_TestGetPreference(ms)

	tests := []struct {
		name   string
		user   User
		SPrefs StandardPreferences
		want   StandardPreferences
	}{
		{
			name: "Change Lang, Leave KG, Add TimeZone",
			user: f.Users[0],
			SPrefs: StandardPreferences{
				Language:   domain.UserPreferenceLanguageFrench,
				WeightUnit: domain.UserPreferenceWeightUnitKGs,
				TimeZone:   "America/New_York",
			},
			want: StandardPreferences{
				Language:   domain.UserPreferenceLanguageFrench,
				WeightUnit: domain.UserPreferenceWeightUnitKGs,
				TimeZone:   "America/New_York",
			},
		},
		{
			name: "Start with none then add one",
			user: f.Users[1],
			SPrefs: StandardPreferences{
				Language: domain.UserPreferenceLanguageFrench,
			},
			want: StandardPreferences{
				Language: domain.UserPreferenceLanguageFrench,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.user.UpdateStandardPreferences(test.SPrefs)
			ms.NoError(err)

			ms.Equal(test.want, got, "incorrect result from UpdateStandardPreferences()")
		})
	}
}

func (ms *ModelSuite) TestUser_GetPreferences() {
	t := ms.T()

	f := CreateUserFixtures_TestGetPreference(ms)

	tests := []struct {
		name string
		user User
		want StandardPreferences
	}{
		{
			name: "english and kgs",
			user: f.Users[0],
			want: StandardPreferences{
				Language:   domain.UserPreferenceLanguageEnglish,
				WeightUnit: domain.UserPreferenceWeightUnitKGs,
			},
		},
		{
			name: "none",
			user: f.Users[1],
			want: StandardPreferences{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.user.GetPreferences()
			ms.NoError(err)

			ms.Equal(test.want, got, "incorrect result from GetPreferences()")
		})
	}
}

func (ms *ModelSuite) TestUser_GetLanguagePreference() {
	t := ms.T()

	users := CreateUserFixtures_TestGetLanguagePreference(ms)

	tests := []struct {
		name string
		user User
		want string
	}{
		{
			name: "english",
			user: users[0],
			want: domain.UserPreferenceLanguageEnglish,
		},
		{
			name: "none so english default",
			user: users[1],
			want: domain.UserPreferenceLanguageEnglish,
		},
		{
			name: "",
			user: users[2],
			want: domain.UserPreferenceLanguageFrench,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.user.GetLanguagePreference()

			ms.Equal(test.want, got, "incorrect result from GetLanguagePreference()")
		})
	}
}

func (ms *ModelSuite) TestUser_GetRealName() {
	t := ms.T()

	tests := []struct {
		name string
		user User
		want string
	}{
		{
			name: "first and last",
			user: User{
				FirstName: "John",
				LastName:  "Doe",
			},
			want: "John Doe",
		},
		{
			name: "first only",
			user: User{FirstName: "Cher"},
			want: "Cher",
		},
		{
			name: "last only",
			user: User{LastName: "Bono"},
			want: "Bono",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.user.GetRealName()

			ms.Equal(test.want, got, "incorrect result from GetRealName()")
		})
	}
}

func (ms *ModelSuite) TestUser_HasOrganization() {
	f := createUserFixtures(ms.DB, 2)
	users := f.Users
	ms.NoError(ms.DB.Destroy(&f.UserOrganizations[1]))

	tests := []struct {
		name string
		user User
		want bool
	}{
		{
			name: "true",
			user: users[0],
			want: true,
		},
		{
			name: "false",
			user: users[1],
			want: false,
		},
	}
	for _, test := range tests {
		ms.T().Run(test.name, func(t *testing.T) {
			got := test.user.HasOrganization()

			ms.Equal(test.want, got, "incorrect result from HasOrganization()")
		})
	}
}

func (ms *ModelSuite) TestUser_MeetingsAsParticipant() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		user    User
		want    []int
		wantErr string
	}{
		{
			name: "creator",
			user: f.Users[0],
			want: []int{},
		},
		{
			name: "organizer",
			user: f.Users[1],
			want: []int{f.Meetings[0].ID},
		},
		{
			name: "invited participant",
			user: f.Users[2],
			want: []int{f.Meetings[0].ID},
		},
		{
			name: "self-joined participant",
			user: f.Users[3],
			want: []int{f.Meetings[0].ID},
		},
		{
			name: "invalid",
			user: User{},
			want: []int{},
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			// exercise
			got, err := tt.user.MeetingsAsParticipant(createTestContext(tt.user))

			// verify
			if tt.wantErr != "" {
				ms.Error(err, `expected error "%s" but got none`, tt.wantErr)
				ms.Contains(err.Error(), tt.wantErr, "unexpected error message")
				return
			}
			ids := make([]int, len(got))
			for i := range got {
				ids[i] = got[i].ID
			}
			ms.Equal(tt.want, ids)

			// teardown
		})
	}
}

func (ms *ModelSuite) TestUser_CanCreateMeetingParticipant() {
	f := createMeetingFixtures(ms.DB, 2)

	tests := []struct {
		name    string
		user    User
		meeting Meeting
		want    bool
	}{
		{
			name:    "creator",
			user:    f.Users[0],
			meeting: f.Meetings[0],
			want:    true,
		},
		{
			name:    "organizer",
			user:    f.Users[1],
			meeting: f.Meetings[0],
			want:    true,
		},
		{
			name:    "invited participant",
			user:    f.Users[2],
			meeting: f.Meetings[0],
			want:    true,
		},
		{
			name:    "self-joined participant",
			user:    f.Users[3],
			meeting: f.Meetings[0],
			want:    true,
		},
		{
			name:    "not yet participant, cannot see meeting",
			user:    f.Users[4],
			meeting: f.Meetings[0],
			want:    true, // will be false when meeting visibility is implemented
		},
		{
			name:    "not yet participant, can see meeting",
			user:    f.Users[4],
			meeting: f.Meetings[0],
			want:    true,
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			// exercise
			got := tt.user.CanCreateMeetingParticipant(createTestContext(tt.user), tt.meeting)

			// verify
			ms.Equal(tt.want, got)
		})
	}
}
