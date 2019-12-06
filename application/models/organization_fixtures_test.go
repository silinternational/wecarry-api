package models

import (
	"strconv"

	"github.com/silinternational/wecarry-api/domain"
)

type OrganizationFixtures struct {
	Organizations
	OrganizationDomains
	Users
}

func createFixturesForOrganizationGetUsers(ms *ModelSuite) OrganizationFixtures {
	org := Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, &org)

	nicknames := []string{"alice", "john", "bob"}
	unique := org.UUID.String()
	users := make(Users, len(nicknames))
	for i := range users {
		users[i].Email = "user" + strconv.Itoa(i) + unique + "example.com"
		users[i].Nickname = nicknames[i] + unique
		users[i].UUID = domain.GetUUID()

		createFixture(ms, &users[i])
	}

	userOrgFixtures := make(UserOrganizations, len(nicknames))
	for i := range userOrgFixtures {
		userOrgFixtures[i].OrganizationID = org.ID
		userOrgFixtures[i].UserID = users[i].ID
		userOrgFixtures[i].AuthID = users[i].Email
		userOrgFixtures[i].AuthEmail = users[i].Email

		createFixture(ms, &userOrgFixtures[i])
	}

	return OrganizationFixtures{
		Organizations: Organizations{org},
		Users:         users,
	}
}

func CreateFixturesForOrganizationGetDomains(ms *ModelSuite) OrganizationFixtures {
	org := Organization{
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		UUID:       domain.GetUUID(),
	}
	createFixture(ms, &org)

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

	user := User{
		Email:     "user1@example.com",
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  "Existing User ",
		UUID:      domain.GetUUID(),
	}
	createFixture(ms, &user)

	return OrganizationFixtures{
		Organizations:       Organizations{org},
		OrganizationDomains: orgDomains,
		Users:               Users{user},
	}
}
