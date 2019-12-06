package actions

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type OrgFixtures struct {
	Users models.Users
	Orgs  models.Organizations
}

func Fixtures_CreateOrganization(as *ActionSuite, t *testing.T) OrgFixtures {

	// Array indexes for convenience in references
	const (
		SalesAdmin    = 0
		OrgMember     = 1
		OrgAdmin      = 2
		OtherOrgAdmin = 3
		Org1          = 0
		Org2          = 1
	)

	users := models.Users{
		{
			Email:     "sales_admin@domain.com",
			FirstName: "Sales",
			LastName:  "Admin",
			Nickname:  "sales_admin",
			AdminRole: models.UserAdminRoleSalesAdmin,
		},
		{
			Email:     "member@domain.com",
			FirstName: "Org",
			LastName:  "Member",
			Nickname:  "org_member",
			AdminRole: models.UserAdminRoleUser,
		},
		{
			Email:     "admin@domain.com",
			FirstName: "Org",
			LastName:  "Admin",
			Nickname:  "org_admin",
			AdminRole: models.UserAdminRoleUser,
		},
		{
			Email:     "admin@other.com",
			FirstName: "Other Org",
			LastName:  "Admin",
			Nickname:  "other_org_admin",
			AdminRole: models.UserAdminRoleUser,
		},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		err := as.DB.Create(&users[i])
		if err != nil {
			t.Errorf("unable to create user fixture %s: %s", users[i].Nickname, err)
		}
	}

	orgs := []models.Organization{
		{
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   models.AuthTypeSaml,
			AuthConfig: "{}",
		},
		{
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   models.AuthTypeSaml,
			AuthConfig: "{}",
		},
	}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		err := as.DB.Create(&orgs[i])
		if err != nil {
			t.Errorf("unable to create orgs fixture named %s: %s", orgs[i].Name, err)
		}
	}

	userOrgs := []models.UserOrganization{
		{
			OrganizationID: orgs[Org1].ID,
			UserID:         users[SalesAdmin].ID,
			Role:           models.UserOrganizationRoleUser,
			AuthID:         users[SalesAdmin].Nickname,
			AuthEmail:      users[SalesAdmin].Email,
		},
		{
			OrganizationID: orgs[Org1].ID,
			UserID:         users[OrgMember].ID,
			Role:           models.UserOrganizationRoleUser,
			AuthID:         users[OrgMember].Nickname,
			AuthEmail:      users[OrgMember].Email,
		},
		{
			OrganizationID: orgs[Org1].ID,
			UserID:         users[OrgAdmin].ID,
			Role:           models.UserOrganizationRoleAdmin,
			AuthID:         users[OrgAdmin].Nickname,
			AuthEmail:      users[OrgAdmin].Email,
		},
		{
			OrganizationID: orgs[Org2].ID,
			UserID:         users[OtherOrgAdmin].ID,
			Role:           models.UserOrganizationRoleAdmin,
			AuthID:         users[OtherOrgAdmin].Nickname,
			AuthEmail:      users[OtherOrgAdmin].Email,
		},
	}
	for i := range userOrgs {
		err := as.DB.Create(&userOrgs[i])
		if err != nil {
			t.Errorf("unable to create user orgs fixture for %s: %s", userOrgs[i].AuthID, err)
		}
	}

	accessTokenFixtures := []models.UserAccessToken{
		{
			UserID:             users[SalesAdmin].ID,
			UserOrganizationID: userOrgs[SalesAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(users[SalesAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OrgMember].ID,
			UserOrganizationID: userOrgs[OrgMember].ID,
			AccessToken:        models.HashClientIdAccessToken(users[OrgMember].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OrgAdmin].ID,
			UserOrganizationID: userOrgs[OrgAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(users[OrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OtherOrgAdmin].ID,
			UserOrganizationID: userOrgs[OtherOrgAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(users[OtherOrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
	}
	for i := range accessTokenFixtures {
		err := as.DB.Create(&accessTokenFixtures[i])
		if err != nil {
			t.Errorf("unable to create access token fixture for index %v: %s", i, err)
		}
	}

	return OrgFixtures{
		Users: users,
		Orgs:  orgs,
	}
}
