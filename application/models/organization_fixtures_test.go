package models

import (
	"strconv"

	"github.com/silinternational/wecarry-api/domain"
)

type OrganizationFixtures struct {
	Organizations
	Users
}

func createFixturesForOrganizationGetUsers(ms *ModelSuite) OrganizationFixtures {
	org := Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, &org)

	nicknames := []string{"alice", "john", "bob"}
	unique := org.Uuid.String()
	users := make(Users, len(nicknames))
	for i := range users {
		users[i].Email = "user" + strconv.Itoa(i) + unique + "example.com"
		users[i].Nickname = nicknames[i] + unique
		users[i].Uuid = domain.GetUuid()

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
