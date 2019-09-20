package actions

import (
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"

	"testing"

	"github.com/gobuffalo/nulls"
)

type UserOrgFixtures struct {
	users    models.Users
	orgs     models.Organizations
	userOrgs models.UserOrganizations
}

func Fixtures_GetOrgAndUserOrgs(as *ActionSuite, t *testing.T) UserOrgFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	orgDomain := models.OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "example.com",
	}

	if err := as.DB.Create(&orgDomain); err != nil {
		t.Errorf("could not create test org domain ... %v", err)
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

	return UserOrgFixtures{
		orgs:     models.Organizations{*org},
		userOrgs: userOrgs,
	}
}

func Fixtures_CreateAuthUser(as *ActionSuite, t *testing.T) UserOrgFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	orgDomain := models.OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "example.com",
	}

	if err := as.DB.Create(&orgDomain); err != nil {
		t.Errorf("could not create test org domain ... %v", err)
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

	return UserOrgFixtures{
		users: users,
		orgs:  models.Organizations{*org},
	}
}
