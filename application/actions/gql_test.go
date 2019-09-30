package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/silinternational/wecarry-api/gqlgen"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"

	"github.com/gobuffalo/httptest"
	"github.com/silinternational/wecarry-api/models"
)

type gqlErrorResponse struct {
	Errors []struct {
		Message string   `json:"message"`
		Path    []string `json:"path"`
	} `json:"errors"`
	Data interface{} `json:"data"`
}

func (as *ActionSuite) Test_UserQuery() {
	t := as.T()
	models.ResetTables(t, as.DB)

	queryFixtures := Fixtures_UserQuery(as, t)
	userFixtures := queryFixtures.Users

	c := getGqlClient()

	query := `{user(id: "` + userFixtures[1].Uuid.String() + `") {id nickname photoURL}}`

	var usersResp struct {
		User struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
			PhotoURL string `json:"photoURL"`
		} `json:"user"`
	}

	gqlgen.TestUser = userFixtures[0]
	gqlgen.TestUser.AdminRole = nulls.NewString(domain.AdminRoleSuperDuperAdmin)
	c.MustPost(query, &usersResp)

	if err := as.DB.Load(&(userFixtures[1]), "PhotoFile"); err != nil {
		t.Errorf("failed to load user fixture, %s", err)
	}
	as.Equal(userFixtures[1].Uuid.String(), usersResp.User.ID)
	as.Equal(userFixtures[1].Nickname, usersResp.User.Nickname)
	as.Equal(userFixtures[1].PhotoFile.URL.String, usersResp.User.PhotoURL)
	as.Regexp("^https?", usersResp.User.PhotoURL)
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
	)

	testFixtures := Fixtures_CreateOrganization(as, t)
	userFixtures := testFixtures.Users
	orgFixtures := testFixtures.Orgs

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

func (as *ActionSuite) Test_PostQuery() {
	t := as.T()
	models.ResetTables(t, as.DB)

	queryFixtures := Fixtures_PostQuery(as, t)
	userFixtures := queryFixtures.Users
	postFixtures := queryFixtures.Posts

	c := getGqlClient()

	query := `{ post(id: "` + postFixtures[1].Uuid.String() + `") { id photo { id } files { id } }}`

	var postsResp struct {
		Post struct {
			ID    string `json:"id"`
			Photo struct {
				ID string `json:"id"`
			} `json:"photo"`
			Files []struct {
				ID string `json:"id"`
			} `json:"files"`
		} `json:"post"`
	}

	gqlgen.TestUser = userFixtures[0]
	c.MustPost(query, &postsResp)

	if err := as.DB.Load(&(postFixtures[1]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
	}

	as.Equal(postFixtures[1].Uuid.String(), postsResp.Post.ID)
	as.Equal(postFixtures[1].PhotoFile.UUID.String(), postsResp.Post.Photo.ID)
	as.Equal(1, len(postsResp.Post.Files))
	as.Equal(postFixtures[1].Files[0].File.UUID.String(), postsResp.Post.Files[0].ID)
}

func getGqlClient() *client.Client {
	h := handler.GraphQL(gqlgen.NewExecutableSchema(gqlgen.Config{Resolvers: &gqlgen.Resolver{}}))
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)
	return c
}
