package actions

import (
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type OrganizationFixtures struct {
	models.Users
	models.Organizations
	models.File
	models.OrganizationDomains
}

func fixturesForCreateOrganization(as *ActionSuite) OrganizationFixtures {
	orgs := models.Organizations{
		{
			Name:       "Primary",
			Url:        nulls.NewString("example.com"),
			AuthType:   models.AuthTypeSaml,
			AuthConfig: "{}",
		},
		{
			Name:       "New",
			Url:        nulls.NewString("example.org"),
			AuthType:   models.AuthTypeSaml,
			AuthConfig: "{}",
		},
	}
	// Don't save "New" to the database, that's for the test to do.
	createFixture(as, &orgs[0])

	userFixtures := test.CreateUserFixtures(as.DB, 1)
	users := userFixtures.Users

	users[0].AdminRole = models.UserAdminRoleSuperAdmin
	as.NoError(as.DB.Save(&users[0]))

	var file models.File
	as.Nil(file.Store("photo.gif", []byte("GIF89a")), "unexpected error storing file")

	return OrganizationFixtures{
		Users:         users,
		Organizations: orgs,
		File:          file,
	}
}

func fixturesForOrganizationCreateRemoveUpdate(as *ActionSuite, t *testing.T) OrganizationFixtures {
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
			UserOrganizationID: nulls.NewInt(userOrgs[SalesAdmin].ID),
			AccessToken:        models.HashClientIdAccessToken(users[SalesAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OrgMember].ID,
			UserOrganizationID: nulls.NewInt(userOrgs[OrgMember].ID),
			AccessToken:        models.HashClientIdAccessToken(users[OrgMember].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OrgAdmin].ID,
			UserOrganizationID: nulls.NewInt(userOrgs[OrgAdmin].ID),
			AccessToken:        models.HashClientIdAccessToken(users[OrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             users[OtherOrgAdmin].ID,
			UserOrganizationID: nulls.NewInt(userOrgs[OtherOrgAdmin].ID),
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

	return OrganizationFixtures{
		Users:         users,
		Organizations: orgs,
	}
}

func fixturesForOrganizationDomain(as *ActionSuite) OrganizationFixtures {
	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(as, &org)

	unique := org.UUID.String()
	users := models.Users{
		{Email: unique + "_user@example.com", Nickname: unique + " User "},
		{Email: unique + "_admin@example.com", Nickname: unique + " Admin "},
	}
	userOrgs := models.UserOrganizations{
		{Role: "user"},
		{Role: "admin"},
	}
	accessTokenFixtures := make([]models.UserAccessToken, len(users))
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(as, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = unique + "_auth_user" + strconv.Itoa(i)
		userOrgs[i].AuthEmail = unique + users[i].Email
		createFixture(as, &userOrgs[i])

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = nulls.NewInt(userOrgs[i].ID)
		accessTokenFixtures[i].AccessToken = models.HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		createFixture(as, &accessTokenFixtures[i])
	}

	return OrganizationFixtures{
		Organizations: models.Organizations{org},
		Users:         users,
	}
}

func fixturesForUpdateOrganization(as *ActionSuite) OrganizationFixtures {
	orgs := make([]models.Organization, 3)
	orgs[0].Name = "default org"
	orgs[1].Name = "trusted org 1"
	orgs[2].Name = "trusted org 2"
	for i := range orgs {
		orgs[i].AuthType = models.AuthTypeSaml
		orgs[i].AuthConfig = "{}"
		orgs[i].Url = nulls.NewString("https://www.example.com")
		test.MustCreate(as.DB, &orgs[i])
	}

	domains := make([]models.OrganizationDomain, 2)
	for i := range domains {
		domains[i] = models.OrganizationDomain{
			OrganizationID: orgs[0].ID,
			Domain:         strconv.Itoa(i) + orgs[0].UUID.String() + ".example.com",
		}
		test.MustCreate(as.DB, &domains[i])
	}

	userFixtures := test.CreateUserFixtures(as.DB, 1)
	users := userFixtures.Users

	users[0].AdminRole = models.UserAdminRoleSuperAdmin
	as.NoError(as.DB.Save(&users[0]))

	var file models.File
	as.Nil(file.Store("photo.gif", []byte("GIF89a")), "unexpected error storing file")

	return OrganizationFixtures{
		Users:               users,
		Organizations:       orgs,
		File:                file,
		OrganizationDomains: domains,
	}
}

func fixturesForCreateTrust(as *ActionSuite) OrganizationFixtures {
	orgs := make([]models.Organization, 3)
	orgs[0].Name = "default org"
	orgs[1].Name = "trusted org 1"
	orgs[2].Name = "trusted org 2"
	for i := range orgs {
		orgs[i].AuthType = models.AuthTypeSaml
		orgs[i].AuthConfig = "{}"
		orgs[i].Url = nulls.NewString("https://www.example.com")
		test.MustCreate(as.DB, &orgs[i])
	}

	trust := models.OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID}
	as.NoError(trust.CreateSymmetric())

	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users
	users[0].AdminRole = models.UserAdminRoleSalesAdmin
	as.NoError(as.DB.Save(&users[0]))
	users[1].AdminRole = models.UserAdminRoleAdmin
	as.NoError(as.DB.Save(&users[1]))

	var file models.File
	as.Nil(file.Store("photo.gif", []byte("GIF89a")), "unexpected error storing file")

	return OrganizationFixtures{
		Users:         users,
		Organizations: orgs,
		File:          file,
	}
}

func fixturesForRemoveTrust(as *ActionSuite) OrganizationFixtures {
	orgs := make([]models.Organization, 3)
	orgs[0].Name = "default org"
	orgs[1].Name = "trusted org 1"
	orgs[2].Name = "trusted org 2"
	for i := range orgs {
		orgs[i].AuthType = models.AuthTypeSaml
		orgs[i].AuthConfig = "{}"
		orgs[i].Url = nulls.NewString("https://www.example.com")
		test.MustCreate(as.DB, &orgs[i])
	}

	trusts := models.OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[2].ID},
	}
	as.NoError(trusts[0].CreateSymmetric())
	as.NoError(trusts[1].CreateSymmetric())

	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users
	users[0].AdminRole = models.UserAdminRoleSalesAdmin
	as.NoError(as.DB.Save(&users[0]))
	users[1].AdminRole = models.UserAdminRoleAdmin
	as.NoError(as.DB.Save(&users[1]))

	var file models.File
	as.Nil(file.Store("photo.gif", []byte("GIF89a")), "unexpected error storing file")

	return OrganizationFixtures{
		Users:         users,
		Organizations: orgs,
		File:          file,
	}
}
