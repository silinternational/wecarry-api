package actions

import (
	"fmt"

	"github.com/silinternational/wecarry-api/models"
)

type OrganizationDomainFixtures struct {
	models.Organization
	models.Users
}

type OrganizationDomainResponse struct {
	OrganizationDomain []struct {
		Organization Organization `json:"organization"`
		Domain       string       `json:"domain"`
	} `json:"domain"`
}

func (as *ActionSuite) Test_CreateUpdateOrganizationDomain() {
	f := fixturesForOrganizationDomain(as)

	testDomain := "example.com"
	allFieldsQuery := `organization { id } domain authType authConfig`
	allFieldsInput := fmt.Sprintf(`organizationID:"%s" domain:"%s" authType:DEFAULT`,
		f.Organizations[0].UUID.String(), testDomain)

	query := fmt.Sprintf("mutation{domain: createOrganizationDomain(input: {%s}) {%s}}",
		allFieldsInput, allFieldsQuery)
	var resp OrganizationDomainResponse
	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)

	as.Equal(1, len(resp.OrganizationDomain), "wrong number of domains in response")
	as.Equal(testDomain, resp.OrganizationDomain[0].Domain, "received wrong domain")
	as.Equal(f.Organizations[0].UUID.String(), resp.OrganizationDomain[0].Organization.ID, "received wrong org ID")

	var orgs models.Organizations
	err = as.DB.Where("name = ?", f.Organizations[0].Name).All(&orgs)
	as.NoError(err)

	as.GreaterOrEqual(1, len(orgs), "no Organization found")
	domains, err := orgs[0].Domains()
	as.NoError(err)
	as.Equal(1, len(domains), "wrong number of domains in DB")
	as.Equal(testDomain, domains[0].Domain, "wrong domain in DB")

	// Test updating orgdomains
	validUpdateInput := fmt.Sprintf(`organizationID:"%s" domain:"%s" authType:SAML, authConfig:"{}"`,
		f.Organizations[0].UUID.String(), testDomain)

	query = fmt.Sprintf("mutation{domain: updateOrganizationDomain(input: {%s}) {%s}}",
		validUpdateInput, allFieldsQuery)
	var resp2 OrganizationDomainResponse
	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp2)
	as.NoError(err)

	// User without admin role - should error
	validUpdateInput = fmt.Sprintf(`organizationID:"%s" domain:"%s" authType:SAML, authConfig:"{}"`,
		f.Organizations[0].UUID.String(), testDomain)

	query = fmt.Sprintf("mutation{domain: updateOrganizationDomain(input: {%s}) {%s}}",
		validUpdateInput, allFieldsQuery)
	var resp3 OrganizationDomainResponse
	err = as.testGqlQuery(query, f.Users[0].Nickname, &resp3)
	as.Error(err)

	// User with admin role but invalid domain - should error
	inValidUpdateInput := fmt.Sprintf(`organizationID:"%s" domain:"%s" authType:SAML, authConfig:"{}"`,
		f.Organizations[0].UUID.String(), "invalid.com")

	query = fmt.Sprintf("mutation{domain: updateOrganizationDomain(input: {%s}) {%s}}",
		inValidUpdateInput, allFieldsQuery)
	var resp4 OrganizationDomainResponse
	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp4)
	as.Error(err)

}
