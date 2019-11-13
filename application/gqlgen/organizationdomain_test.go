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
		OrganizationID string `json:"organizationID"`
		Domain         string `json:"domain"`
	} `json:"domain"`
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

	testDomain := "example.com"
	allFieldsQuery := `organizationID domain`
	allFieldsInput := fmt.Sprintf(`organizationID:"%s" domain:"%s"`,
		f.Organization.Uuid.String(), testDomain)

	var resp OrganizationDomainResponse
	TestUser = f.Users[1]
	err := c.Post(fmt.Sprintf("mutation{domain: createOrganizationDomain(input: {%s}) {%s}}",
		allFieldsInput, allFieldsQuery), &resp)
	gs.NoError(err)

	gs.Equal(1, len(resp.OrganizationDomain), "wrong number of domains in response")
	gs.Equal(testDomain, resp.OrganizationDomain[0].Domain, "received wrong domain")
	gs.Equal(f.Organization.Uuid.String(), resp.OrganizationDomain[0].OrganizationID, "received wrong org ID")

	var orgs models.Organizations
	err = gs.DB.Eager().Where("name = ?", f.Organization.Name).All(&orgs)
	gs.NoError(err)

	gs.GreaterOrEqual(1, len(orgs), "no Organization found")
	gs.Equal(1, len(orgs[0].OrganizationDomains), "wrong number of domains in DB")
	gs.Equal(testDomain, orgs[0].OrganizationDomains[0].Domain, "wrong domain in DB")

	// removeOrganization is tested in `actions/gql_test.go`. This tests the returned data, whereas the `actions`
	// test is more thorough with error cases.
}
