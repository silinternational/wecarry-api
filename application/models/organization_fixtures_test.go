package models

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
)

type OrganizationFixtures struct {
	Organizations
	OrganizationDomains
	Users
}

func createOrgFixtures(ms *ModelSuite, t *testing.T) (Organization, OrganizationDomain) {
	// Load Organization test fixtures
	orgs := createOrganizationFixtures(ms.DB, 1)

	// Load Organization Domains test fixtures
	orgDomain := OrganizationDomain{
		OrganizationID: orgs[0].ID,
		Domain:         "example.org",
	}
	err := ms.DB.Create(&orgDomain)
	ms.NoError(err, "could not create org domain fixtures")

	return orgs[0], orgDomain
}

func createFixturesForOrganizationGetUsers(ms *ModelSuite) OrganizationFixtures {
	uf := createUserFixtures(ms.DB, 3)
	org := uf.Organization
	users := uf.Users

	// nicknames in unsorted order
	nicknames := []string{"alice", "john", "bob"}
	for i := range users {
		users[i].Nickname = nicknames[i]
		ms.NoError(ms.DB.Save(&users[i]))
	}

	return OrganizationFixtures{
		Organizations: Organizations{org},
		Users:         users,
	}
}

func CreateFixturesForOrganizationGetDomains(ms *ModelSuite) OrganizationFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	orgDomains := OrganizationDomains{
		{
			OrganizationID: org.ID,
			Domain:         "example.org",
		},
		{
			OrganizationID: org.ID,
			Domain:         "1.example.org",
		},
		{
			OrganizationID: org.ID,
			Domain:         "example.com",
		},
	}
	for i := range orgDomains {
		createFixture(ms, &orgDomains[i])
	}

	return OrganizationFixtures{
		Organizations:       Organizations{org},
		OrganizationDomains: orgDomains,
		Users:               Users{user},
	}
}

func createFixturesForOrganization_AllWhereUserIsOrgAdmin(ms *ModelSuite) OrganizationFixtures {
	orgs := []Organization{
		{ID: 1, Name: "NoAdmin"},
		{ID: 2, Name: "NoAdmin"},
		{ID: 3, Name: "NoAdmin"},
		{ID: 4, Name: "Admin Users 4 & 5"},
		{ID: 5, Name: "Admin User 5"},
	}
	for i := range orgs {
		orgs[i].CreatedAt = time.Time{}
		orgs[i].UpdatedAt = time.Time{}
		orgs[i].Url = nulls.String{}
		orgs[i].AuthType = "na"
		orgs[i].AuthConfig = "{}"
		orgs[i].UUID = domain.GetUUID()

		err := ms.DB.Create(&orgs[i])
		ms.NoError(err, "Unable to create org fixture")
	}

	const userCount = 5
	users := make(Users, userCount)
	userOrgs := make(UserOrganizations, userCount+1)
	unique := domain.GetUUID().String()

	users[0].AdminRole = UserAdminRoleSuperAdmin
	users[1].AdminRole = UserAdminRoleSalesAdmin

	uOrgRoles := [userCount]string{
		UserOrganizationRoleUser, UserOrganizationRoleUser, UserOrganizationRoleUser,
		UserOrganizationRoleAdmin, UserOrganizationRoleAdmin,
	}

	for i := range users {
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].FirstName = "first" + strconv.Itoa(i)
		users[i].LastName = "last" + strconv.Itoa(i)
		mustCreate(ms.DB, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = orgs[i].ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		userOrgs[i].Role = uOrgRoles[i]
		mustCreate(ms.DB, &userOrgs[i])

		if err := ms.DB.Load(&users[i], "Organizations"); err != nil {
			panic(fmt.Sprintf("failed to load organizations on users[%d] fixture, %s", i, err))
		}
	}

	repeatAdmin := userOrgs[5]

	// Make the last (fifth) user also an admin for the fourth organization
	repeatAdmin.UserID = users[4].ID
	repeatAdmin.OrganizationID = orgs[3].ID
	repeatAdmin.AuthID = "extra_" + users[4].Email
	repeatAdmin.AuthEmail = users[4].Email
	repeatAdmin.Role = UserOrganizationRoleAdmin
	mustCreate(ms.DB, &repeatAdmin)

	if err := ms.DB.Load(&users[4], "Organizations"); err != nil {
		panic(fmt.Sprintf("failed to load organizations on users[4] fixture, %s", err))
	}

	return OrganizationFixtures{
		Organizations: orgs,
		Users:         users,
	}
}
