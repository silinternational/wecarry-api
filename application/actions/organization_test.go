package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
	"io/ioutil"
	"strings"
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

// Test_CreateOrganization tests the CreateOrganization GraphQL mutation
func (as *ActionSuite) Test_CreateOrganization() {
	f := fixturesForCreateOrganization(as)

	var resp OrganizationResponse

	allFieldsQuery := `id domains { domain organizationID } name url`
	allFieldsInput := fmt.Sprintf(
		`name:"%s" url:"%s" authType:"%s" authConfig:"%s"`,
		f.Organizations[1].Name,
		f.Organizations[1].Url.String,
		f.Organizations[1].AuthType,
		f.Organizations[1].AuthConfig)

	query := fmt.Sprintf("mutation{organization: createOrganization(input: {%s}) {%s}}",
		allFieldsInput, allFieldsQuery)
	err := as.testGqlQuery(query, f.Users[0].Nickname, &resp)
	as.NoError(err)

	as.Equal(f.Organizations[1].Name, resp.Organization.Name, "received wrong name")
	as.Equal(f.Organizations[1].Url.String, resp.Organization.URL, "received wrong URL")
	as.Equal(0, len(resp.Organization.Domains))

	var orgs models.Organizations
	err = as.DB.Where("name = ?", f.Organizations[1].Name).All(&orgs)
	as.NoError(err)

	as.GreaterOrEqual(1, len(orgs), "no Organization record created")
	as.Equal(f.Organizations[1].Name, orgs[0].Name, "Name doesn't match")
	as.Equal(f.Organizations[1].Url, orgs[0].Url, "URL doesn't match")
	as.Equal(f.Organizations[1].AuthType, orgs[0].AuthType, "AuthType doesn't match")
	as.Equal(f.Organizations[1].AuthConfig, orgs[0].AuthConfig, "AuthConfig doesn't match")

	domains, _ := orgs[0].GetDomains()
	as.Equal(0, len(domains), "new organization has unexpected domains")

	as.Equal(resp.Organization.ID, orgs[0].UUID.String(), "UUID doesn't match")
}

func (as *ActionSuite) Test_OrganizationCreateRemoveUpdate() {
	t := as.T()

	// Array indexes for convenience in references
	const (
		SalesAdmin    = 0
		OrgMember     = 1
		OrgAdmin      = 2
		OtherOrgAdmin = 3
		Org1          = 0
	)

	testFixtures := fixturesForOrganizationCreateRemoveUpdate(as, t)
	userFixtures := testFixtures.Users
	orgFixtures := testFixtures.Organizations

	type testCase struct {
		Name            string
		Token           string
		Payload         string
		ExpectError     bool
		ExpectSubString string
	}

	createOrgPayload := `{"query": "mutation { createOrganization(input: { name: \"new org\", url: \"http://test.com\", authType: \"saml2\", authConfig: \"{}\", }){id} }"}`
	updateOrgPayload := fmt.Sprintf(`{"query": "mutation { updateOrganization(input: { id: \"%s\" name: \"updated org\", url: \"http://test.com\", authType: \"saml2\", authConfig: \"{}\", }){id} }"}`, orgFixtures[Org1].UUID.String())
	createOrgDomainPayload := fmt.Sprintf(`{"query": "mutation { createOrganizationDomain(input: { organizationID: \"%s\", domain: \"newdomain.com\"}){domain} }"}`, orgFixtures[Org1].UUID.String())
	removeOrgDomainPayload := fmt.Sprintf(`{"query": "mutation { removeOrganizationDomain(input: { organizationID: \"%s\", domain: \"newdomain.com\"}){domain} }"}`, orgFixtures[Org1].UUID.String())

	testCases := []testCase{
		{
			Name:        "org member cannot create org",
			Token:       userFixtures[OrgMember].Nickname,
			Payload:     createOrgPayload,
			ExpectError: true,
		},
		{
			Name:        "org admin cannot create org",
			Token:       userFixtures[OrgAdmin].Nickname,
			Payload:     createOrgPayload,
			ExpectError: true,
		},
		{
			Name:            "sales admin can create org",
			Token:           userFixtures[SalesAdmin].Nickname,
			Payload:         createOrgPayload,
			ExpectError:     false,
			ExpectSubString: "createOrganization",
		},
		{
			Name:        "org member cannot update org",
			Token:       userFixtures[OrgMember].Nickname,
			Payload:     updateOrgPayload,
			ExpectError: true,
		},
		{
			Name:            "org admin can update org",
			Token:           userFixtures[OrgAdmin].Nickname,
			Payload:         updateOrgPayload,
			ExpectError:     false,
			ExpectSubString: "updateOrganization",
		},
		{
			Name:        "other org admin cannot update org1",
			Token:       userFixtures[OtherOrgAdmin].Nickname,
			Payload:     updateOrgPayload,
			ExpectError: true,
		},
		{
			Name:            "sales admin can update org",
			Token:           userFixtures[SalesAdmin].Nickname,
			Payload:         updateOrgPayload,
			ExpectError:     false,
			ExpectSubString: "updateOrganization",
		},
		{
			Name:        "org member cannot create org domain",
			Token:       userFixtures[OrgMember].Nickname,
			Payload:     createOrgDomainPayload,
			ExpectError: true,
		},
		{
			Name:            "org admin can create org domain",
			Token:           userFixtures[OrgAdmin].Nickname,
			Payload:         createOrgDomainPayload,
			ExpectError:     false,
			ExpectSubString: "createOrganizationDomain",
		},
		{
			Name:        "org admin cannot create duplicate org domain",
			Token:       userFixtures[OrgAdmin].Nickname,
			Payload:     createOrgDomainPayload,
			ExpectError: true,
		},
		{
			Name:        "org member cannot remove org domain",
			Token:       userFixtures[OrgMember].Nickname,
			Payload:     removeOrgDomainPayload,
			ExpectError: true,
		},
		{
			Name:            "org admin can remove org domain",
			Token:           userFixtures[OrgAdmin].Nickname,
			Payload:         removeOrgDomainPayload,
			ExpectError:     false,
			ExpectSubString: "removeOrganizationDomain",
		},
		{
			Name:        "other org admin cannot create org1 domain",
			Token:       userFixtures[OtherOrgAdmin].Nickname,
			Payload:     createOrgDomainPayload,
			ExpectError: true,
		},
		{
			Name:            "sales admin can create org domain",
			Token:           userFixtures[SalesAdmin].Nickname,
			Payload:         createOrgDomainPayload,
			ExpectError:     false,
			ExpectSubString: "createOrganizationDomain",
		},
		{
			Name:            "sales admin can remove org domain",
			Token:           userFixtures[SalesAdmin].Nickname,
			Payload:         removeOrgDomainPayload,
			ExpectError:     false,
			ExpectSubString: "removeOrganizationDomain",
		},
	}

	for _, tc := range testCases {

		payload := bytes.NewReader([]byte(tc.Payload))
		req := httptest.NewRequest("POST", "/gql", payload)
		resp := httptest.NewRecorder()
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set("content-type", "application/json")

		as.App.ServeHTTP(resp, req)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		if tc.ExpectError {
			var errResp gqlErrorResponse
			err = json.Unmarshal(body, &errResp)
			if err != nil {
				t.Error(err)
			}

			if len(errResp.Errors) == 0 {
				t.Errorf("did not get expected errors in test %s, response: +%v", tc.Name, errResp)
			}

			continue
		}

		if !strings.Contains(string(body), tc.ExpectSubString) {
			t.Errorf("substring \"%s\" not found in response: %s. test case: %s", tc.ExpectSubString, string(body), tc.Name)
		}

	}
}

func (as *ActionSuite) Test_OrganizationViewAndList() {
	t := as.T()

	// create organizations
	org1 := models.Organization{
		Name:                "Org1",
		Url:                 nulls.String{},
		AuthType:            models.AuthTypeSaml,
		AuthConfig:          "{}",
		UUID:                domain.GetUUID(),
	}
	test.MustCreate(as.DB, &org1)
	// create organizations
	org2 := models.Organization{
		Name:                "Org2",
		Url:                 nulls.String{},
		AuthType:            models.AuthTypeSaml,
		AuthConfig:          "{}",
		UUID:                domain.GetUUID(),
	}
	test.MustCreate(as.DB, &org2)

	userFixtures := test.CreateUserFixtures(as.DB, 4)
	for i, _ := range userFixtures.Users {
		_ = as.DB.Load(&userFixtures.Users[i], "UserOrganizations", "AccessTokens")
	}

	// user 0 will be a super admin, user 1 will be a sales admin, and user 2 will be an org admin for org1, user 3 will be a user
	userFixtures.Users[0].AdminRole = models.UserAdminRoleSuperAdmin
	as.NoError(as.DB.Save(&userFixtures.Users[0]))

	userFixtures.Users[1].AdminRole = models.UserAdminRoleSalesAdmin
	as.NoError(as.DB.Save(&userFixtures.Users[1]))

	userFixtures.Users[2].UserOrganizations[0].Role = models.UserOrganizationRoleAdmin
	as.NoError(as.DB.Save(&userFixtures.Users[2].UserOrganizations[0]))

	// Test view org1 for each user. users 0,1,2 should succeed, user 3 should fail
	viewOrg1Payload := fmt.Sprintf(`{"query": "{organization(id: \"%s\"){id}}"}`, org1.UUID.String())
	listOrgsPayload := `{"query": "{organizations{id}}"}`

	type testCase struct {
		Name            string
		Token           string
		Payload         string
		ExpectError     bool
		ExpectSubStrings []string
		DontExpectSubStrings []string
	}

	testCases := []testCase{
		{
			Name: "View org 1 as user 0 (super admin)",
			Token: userFixtures.Users[0].Nickname,
			Payload: viewOrg1Payload,
			ExpectError: false,
			ExpectSubStrings: []string{org1.UUID.String()},
		},
		{
			Name: "View org 1 as user 1 (sales admin)",
			Token: userFixtures.Users[1].Nickname,
			Payload: viewOrg1Payload,
			ExpectError: false,
			ExpectSubStrings: []string{org1.UUID.String()},
		},
		{
			Name: "View org 1 as user 2 (org admin)",
			Token: userFixtures.Users[2].Nickname,
			Payload: viewOrg1Payload,
			ExpectError: false,
			ExpectSubStrings: []string{org1.UUID.String()},
		},
		{
			Name: "View org 1 as user 3 (normal user)",
			Token: userFixtures.Users[3].Nickname,
			Payload: viewOrg1Payload,
			ExpectError: true,
		},
		{
			Name: "List orgs as user 0 (super admin)",
			Token: userFixtures.Users[0].Nickname,
			Payload: listOrgsPayload,
			ExpectError: false,
			ExpectSubStrings: []string{org1.UUID.String(), org2.UUID.String()},
		},
		{
			Name: "List orgs as user 1 (sales admin)",
			Token: userFixtures.Users[1].Nickname,
			Payload: listOrgsPayload,
			ExpectError: false,
			ExpectSubStrings: []string{org1.UUID.String(), org2.UUID.String()},
		},
		{
			Name: "List orgs as user 2 (org admin)",
			Token: userFixtures.Users[2].Nickname,
			Payload: listOrgsPayload,
			ExpectError: false,
			ExpectSubStrings: []string{org1.UUID.String()},
			DontExpectSubStrings: []string{org2.UUID.String()},
		},
		{
			Name: "List orgs as user 3 (normal user)",
			Token: userFixtures.Users[3].Nickname,
			Payload: listOrgsPayload,
			ExpectError: false,
			DontExpectSubStrings: []string{org1.UUID.String(), org2.UUID.String()},
		},
	}

	for _, tc := range testCases {
		payload := bytes.NewReader([]byte(tc.Payload))
		req := httptest.NewRequest("POST", "/gql", payload)
		resp := httptest.NewRecorder()
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set("content-type", "application/json")

		as.App.ServeHTTP(resp, req)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		if tc.ExpectError {
			var errResp gqlErrorResponse
			err = json.Unmarshal(body, &errResp)
			if err != nil {
				t.Error(err)
			}

			if len(errResp.Errors) == 0 {
				t.Errorf("did not get expected errors in test %s, response: +%v", tc.Name, errResp)
			}

			continue
		}

		for _, sub := range tc.ExpectSubStrings {
			if !strings.Contains(string(body), sub) {
				t.Errorf("ExpectSubString \"%s\" not found in response: %s. test case: %s", sub, string(body), tc.Name)
			}
		}

		for _, sub := range tc.DontExpectSubStrings {
			if strings.Contains(string(body), sub) {
				t.Errorf("DontExpectSubString \"%s\" found in response and should not have been: %s. test case: %s", sub, string(body), tc.Name)
			}
		}

	}

}
