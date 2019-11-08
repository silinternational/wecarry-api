package gqlgen

import (
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type OrganizationDomainFixtures struct {
	models.Organization
	models.Users
}

type OrganizationDomainResponse struct {
	OrganizationDomain []struct {
		ID     string `json:"organizationID"`
		Domain string `json:"domain"`
	} `json:"organizationDomain"`
}

func createFixtures_OrganizationDomain(gs *GqlgenSuite) OrganizationDomainFixtures {
	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(gs, &org)

	unique := org.Uuid.String()
	users := models.Users{
		{Email: unique + "_user@example.com", Nickname: unique + " User ", Uuid: domain.GetUuid()},
		{Email: unique + "_admin@example.com", Nickname: unique + " Admin ", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(gs, &(users[i]))
	}

	userOrgs := models.UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			Role:           "user",
			AuthID:         users[0].Email,
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			Role:           "admin",
			AuthID:         users[1].Email,
			AuthEmail:      users[1].Email,
		},
	}
	for i := range userOrgs {
		createFixture(gs, &(userOrgs[i]))
	}

	return OrganizationDomainFixtures{
		Organization: org,
		Users:        users,
	}
}

func (gs *GqlgenSuite) Test_CreateOrganizationDomain() {
	f := createFixtures_OrganizationDomain(gs)
	c := getGqlClient()

	input := `organizationID: "` + f.Organization.Uuid.String() + `" 
		domain: "example.org"`
	query := `mutation { organizationDomain: createOrganizationDomain(input: {` + input + `})
		{ domain organizationID }}`
	fmt.Printf("------- query=%s\n", query)

	var resp OrganizationDomainResponse

	TestUser = f.Users[0]
	gs.Error(c.Post(query, &resp))

	TestUser = f.Users[1]
	gs.NoError(c.Post(query, &resp))

	gs.Equal(1, len(resp.OrganizationDomain), "received wrong number of domains")
	gs.Equal(f.Organization.Uuid.String(), resp.OrganizationDomain[0].ID, "received incorrect org UUID")
	gs.Equal("example.org", resp.OrganizationDomain[0].Domain, "received incorrect domain")

	// removeOrganization is tested in `actions/gql_test.go`. This tests the returned data, whereas the `actions`
	// test is more thorough with error cases.
}
