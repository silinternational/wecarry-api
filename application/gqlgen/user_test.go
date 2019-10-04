package gqlgen

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// UserQueryFixtures is for returning fixtures from Fixtures_UserQuery
type UserQueryFixtures struct {
	Users       models.Users
	CurrentUser models.User
	ClientID    string
	AccessToken string
}

// Fixtures_UserQuery creates fixtures for Test_UserQuery
func Fixtures_UserQuery(as *ActionSuite, t *testing.T) UserQueryFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   models.AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
			AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for i := range users {
		if err := as.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         "auth_user1",
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         "auth_user2",
			AuthEmail:      users[1].Email,
		},
	}

	for i := range userOrgs {
		if err := as.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	clientID := "12345678"
	accessToken := "ABCDEFGHIJKLMONPQRSTUVWXYZ123456"
	hash := models.HashClientIdAccessToken(clientID + accessToken)

	userAccessToken := models.UserAccessToken{
		UserID:             users[0].ID,
		UserOrganizationID: userOrgs[0].ID,
		AccessToken:        hash,
		ExpiresAt:          time.Now().Add(time.Hour),
	}

	if err := as.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var f models.File

	if err := f.Store("photo.gif", []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := users[1].AttachPhoto(f.UUID.String()); err != nil {
		t.Errorf("failed to attach photo to user, %s", err)
		t.FailNow()
	}

	return UserQueryFixtures{
		Users:       users,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}

// Test_UserQuery tests the User GraphQL query
func (as *ActionSuite) Test_UserQuery() {
	t := as.T()
	models.ResetTables(t, as.DB)

	queryFixtures := Fixtures_UserQuery(as, t)
	userFixtures := queryFixtures.Users

	c := getGqlClient()

	query := `{user(id: "` + userFixtures[1].Uuid.String() + `") {id nickname photoURL}}`

	var usersResp struct {
		User struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
			PhotoURL string `json:"photoURL"`
		} `json:"user"`
	}

	TestUser = userFixtures[0]
	TestUser.AdminRole = nulls.NewString(domain.AdminRoleSuperDuperAdmin)
	c.MustPost(query, &usersResp)

	if err := as.DB.Load(&(userFixtures[1]), "PhotoFile"); err != nil {
		t.Errorf("failed to load user fixture, %s", err)
	}
	as.Equal(userFixtures[1].Uuid.String(), usersResp.User.ID)
	as.Equal(userFixtures[1].Nickname, usersResp.User.Nickname)
	as.Equal(userFixtures[1].PhotoFile.URL, usersResp.User.PhotoURL)
	as.Regexp("^https?", usersResp.User.PhotoURL)
}
