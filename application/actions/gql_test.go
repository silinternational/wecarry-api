package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

type gqlErrorResponse struct {
	Errors []struct {
		Message string   `json:"message"`
		Path    []string `json:"path"`
	} `json:"errors"`
	Data interface{} `json:"data"`
}

func (as *ActionSuite) TestQueryAUser() {
	t := as.T()
	models.ResetTables(t, as.DB)

	queryFixtures := Fixtures_QueryAUser(as, t)
	userFixtures := queryFixtures.Users

	tUuid := userFixtures[1].Uuid.String()

	uq := map[string]string{
		"query": `{user(id: "` + tUuid + `") {id nickname}}`,
	}

	bearer := queryFixtures.ClientID + queryFixtures.AccessToken
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + bearer,
	}

	hj := as.JSON("/gql")
	hj.Headers = headers
	res := hj.Post(uq)

	as.Equal(200, res.Code)

	u2Uuid := userFixtures[1].Uuid.String()
	u2Nname := userFixtures[1].Nickname
	expectedBody := `{"data":{"user":{"id":"` + u2Uuid + `","nickname":"` + u2Nname + `"}}}`
	as.Equal(expectedBody, res.Body.String())

}

func (as *ActionSuite) Test_CreateOrganization() {
	t := as.T()
	models.ResetTables(as.T(), as.DB)

	// Array indexes for convenience in references
	const (
		SalesAdmin    = 0
		OrgMember     = 1
		OrgAdmin      = 2
		OtherOrgAdmin = 3
		Org1          = 0
		Org2          = 1
	)

	userFixtures := []models.User{
		{
			Email:     "sales_admin@domain.com",
			FirstName: "Sales",
			LastName:  "Admin",
			Nickname:  "sales_admin",
			AdminRole: nulls.NewString(domain.AdminRoleSalesAdmin),
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "member@domain.com",
			FirstName: "Org",
			LastName:  "Member",
			Nickname:  "org_member",
			AdminRole: nulls.String{},
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "admin@domain.com",
			FirstName: "Org",
			LastName:  "Admin",
			Nickname:  "org_admin",
			AdminRole: nulls.String{},
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "admin@other.com",
			FirstName: "Other Org",
			LastName:  "Admin",
			Nickname:  "other_org_admin",
			AdminRole: nulls.String{},
			Uuid:      domain.GetUuid(),
		},
	}
	for i := range userFixtures {
		err := as.DB.Create(&userFixtures[i])
		if err != nil {
			t.Errorf("unable to create user fixture %s: %s", userFixtures[i].Nickname, err)
		}
	}

	orgFixtures := []models.Organization{
		{
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   "saml2",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
		{
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   "saml2",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
	}
	for i := range orgFixtures {
		err := as.DB.Create(&orgFixtures[i])
		if err != nil {
			t.Errorf("unable to create org fixture named %s: %s", orgFixtures[i].Name, err)
		}
	}

	userOrgFixtures := []models.UserOrganization{
		{
			OrganizationID: orgFixtures[Org1].ID,
			UserID:         userFixtures[SalesAdmin].ID,
			Role:           models.UserOrganizationRoleMember,
			AuthID:         userFixtures[SalesAdmin].Nickname,
			AuthEmail:      userFixtures[SalesAdmin].Email,
		},
		{
			OrganizationID: orgFixtures[Org1].ID,
			UserID:         userFixtures[OrgMember].ID,
			Role:           models.UserOrganizationRoleMember,
			AuthID:         userFixtures[OrgMember].Nickname,
			AuthEmail:      userFixtures[OrgMember].Email,
		},
		{
			OrganizationID: orgFixtures[Org1].ID,
			UserID:         userFixtures[OrgAdmin].ID,
			Role:           models.UserOrganizationRoleAdmin,
			AuthID:         userFixtures[OrgAdmin].Nickname,
			AuthEmail:      userFixtures[OrgAdmin].Email,
		},
		{
			OrganizationID: orgFixtures[Org2].ID,
			UserID:         userFixtures[OtherOrgAdmin].ID,
			Role:           models.UserOrganizationRoleAdmin,
			AuthID:         userFixtures[OtherOrgAdmin].Nickname,
			AuthEmail:      userFixtures[OtherOrgAdmin].Email,
		},
	}
	for i := range userOrgFixtures {
		err := as.DB.Create(&userOrgFixtures[i])
		if err != nil {
			t.Errorf("unable to create user org fixture for %s: %s", userOrgFixtures[i].AuthID, err)
		}
	}

	accessTokenFixtures := []models.UserAccessToken{
		{
			UserID:             userFixtures[SalesAdmin].ID,
			UserOrganizationID: userOrgFixtures[SalesAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(userFixtures[SalesAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             userFixtures[OrgMember].ID,
			UserOrganizationID: userOrgFixtures[OrgMember].ID,
			AccessToken:        models.HashClientIdAccessToken(userFixtures[OrgMember].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             userFixtures[OrgAdmin].ID,
			UserOrganizationID: userOrgFixtures[OrgAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(userFixtures[OrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
		{
			UserID:             userFixtures[OtherOrgAdmin].ID,
			UserOrganizationID: userOrgFixtures[OtherOrgAdmin].ID,
			AccessToken:        models.HashClientIdAccessToken(userFixtures[OtherOrgAdmin].Nickname),
			ExpiresAt:          time.Now().Add(time.Minute * 60),
		},
	}
	for i := range accessTokenFixtures {
		err := as.DB.Create(&accessTokenFixtures[i])
		if err != nil {
			t.Errorf("unable to create access token fixture for index %v: %s", i, err)
		}
	}

	type testCase struct {
		Name            string
		Token           string
		Payload         string
		ExpectError     bool
		ExpectSubString string
	}

	createOrgPayload := `{"query": "mutation { createOrganization(input: { name: \"new org\", url: \"http://test.com\", authType: \"saml2\", authConfig: \"{}\", }){id} }"}`
	updateOrgPayload := fmt.Sprintf(`{"query": "mutation { updateOrganization(input: { id: \"%s\" name: \"updated org\", url: \"http://test.com\", authType: \"saml2\", authConfig: \"{}\", }){id} }"}`, orgFixtures[Org1].Uuid.String())
	createOrgDomainPayload := fmt.Sprintf(`{"query": "mutation { createOrganizationDomain(input: { organizationID: \"%s\", domain: \"newdomain.com\"}){domain} }"}`, orgFixtures[Org1].Uuid.String())
	removeOrgDomainPayload := fmt.Sprintf(`{"query": "mutation { removeOrganizationDomain(input: { organizationID: \"%s\", domain: \"newdomain.com\"}){domain} }"}`, orgFixtures[Org1].Uuid.String())

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
