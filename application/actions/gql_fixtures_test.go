package actions

import (
	buffalo_models "github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"testing"
	"time"
)

type QueryFixtures struct {
	Users       models.Users
	CurrentUser models.User
	ClientID    string
	AccessToken string
}

func Fixtures_QueryAUser(t *testing.T) QueryFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		ID:         1,
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := buffalo_models.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			ID:        1,
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
			AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin),
		},
		{
			ID:        2,
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for _, user := range users {
		if err := models.DB.Create(&user); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
		{
			ID:             1,
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         "auth_user1",
			AuthEmail:      users[0].Email,
		},
		{
			ID:             2,
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         "auth_user2",
			AuthEmail:      users[1].Email,
		},
	}

	for _, uOrg := range userOrgs {
		if err := models.DB.Create(&uOrg); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	clientID := "12345678"
	accessToken := "ABCDEFGHIJKLMONPQRSTUVWXYZ123456"
	hash := models.HashClientIdAccessToken(clientID + accessToken)

	userAccessToken := models.UserAccessToken{
		ID:                 1,
		UserID:             users[0].ID,
		UserOrganizationID: userOrgs[0].ID,
		AccessToken:        hash,
		ExpiresAt:          time.Now().Add(time.Hour),
	}

	if err := models.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	return QueryFixtures{
		Users:       users,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}
