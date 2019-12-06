package gqlgen

import (
	"fmt"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// OrganizationResponse is for marshalling Organization query and mutation responses
type OrganizationResponse struct {
	Organization struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		URL       string `json:"url"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		Domains   []struct {
			Domain         string `json:"domain"`
			OrganizationID string `json:"organizationID"`
		} `json:"domains"`
	} `json:"organization"`
}

// OrganizationFixtures is for returning fixtures from Fixtures_UserQuery
type OrganizationFixtures struct {
	models.Organizations
	models.Users
}

// Fixtures_CreateOrganization creates fixtures for Test_CreateOrganization
func Fixtures_CreateOrganization(gs *GqlgenSuite) OrganizationFixtures {
	org := models.Organization{
		Name:       "Acme",
		Url:        nulls.NewString("example.com"),
		AuthType:   "saml",
		AuthConfig: "{}",
	}
	// Don't save to the database, that's for the test to do.

	unique := domain.GetUUID().String()
	users := models.Users{
		{
			UUID:      domain.GetUUID(),
			Email:     unique + "_user1@example.com",
			Nickname:  unique + " User1",
			AdminRole: models.UserAdminRoleSuperAdmin,
		},
	}
	for i := range users {
		createFixture(gs, &users[i])
	}

	return OrganizationFixtures{
		Users:         users,
		Organizations: models.Organizations{org},
	}
}

// Test_CreateOrganization tests the CreateOrganization GraphQL mutation
func (gs *GqlgenSuite) Test_CreateOrganization() {
	f := Fixtures_CreateOrganization(gs)
	c := getGqlClient()

	var resp OrganizationResponse

	allFieldsQuery := `id domains { domain organizationID } name url`
	allFieldsInput := fmt.Sprintf(
		`name:"%s" url:"%s" authType:"%s" authConfig:"%s"`,
		f.Organizations[0].Name,
		f.Organizations[0].Url.String,
		f.Organizations[0].AuthType,
		f.Organizations[0].AuthConfig)

	TestUser = f.Users[0]
	err := c.Post(fmt.Sprintf("mutation{organization: createOrganization(input: {%s}) {%s}}",
		allFieldsInput, allFieldsQuery), &resp)
	gs.NoError(err)

	gs.Equal(f.Organizations[0].Name, resp.Organization.Name, "received wrong name")
	gs.Equal(f.Organizations[0].Url.String, resp.Organization.URL, "received wrong URL")
	gs.Equal(0, len(resp.Organization.Domains))

	var orgs models.Organizations
	err = gs.DB.Where("name = ?", f.Organizations[0].Name).All(&orgs)
	gs.NoError(err)

	gs.GreaterOrEqual(1, len(orgs), "no Organization record created")
	gs.Equal(f.Organizations[0].Name, orgs[0].Name, "Name doesn't match")
	gs.Equal(f.Organizations[0].Url, orgs[0].Url, "URL doesn't match")
	gs.Equal(f.Organizations[0].AuthType, orgs[0].AuthType, "AuthType doesn't match")
	gs.Equal(f.Organizations[0].AuthConfig, orgs[0].AuthConfig, "AuthConfig doesn't match")

	domains, _ := orgs[0].GetDomains()
	gs.Equal(0, len(domains), "new organization has unexpected domains")

	gs.Equal(resp.Organization.ID, orgs[0].UUID.String(), "UUID doesn't match")
}
