package actions

import (
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
	"net/http"
	"testing"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) Test_GQLAddMeAsPotentialProvider() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	requests := f.Requests

	const qTemplate = `mutation {request: addMeAsPotentialProvider (requestID: "%s")` +
		` {id title potentialProviders{id nickname}}}`

	// Add one to Request with none
	query := fmt.Sprintf(qTemplate, requests[2].UUID.String())

	var resp RequestResponse

	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)
	as.Equal(requests[2].UUID.String(), resp.Request.ID, "incorrect Request UUID")
	as.Equal(requests[2].Title, resp.Request.Title, "incorrect Request title")

	want := []PotentialProvider{{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname}}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")

	// Add one to Request with two already
	query = fmt.Sprintf(qTemplate, requests[1].UUID.String())

	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)
	as.Equal(requests[1].UUID.String(), resp.Request.ID, "incorrect Request UUID")
	as.Equal(requests[1].Title, resp.Request.Title, "incorrect Request title")

	want = []PotentialProvider{
		{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
	}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")

	// Adding a repeat gives an error
	query = fmt.Sprintf(qTemplate, requests[1].UUID.String())

	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an error (unique together) but didn't get one")

	want = []PotentialProvider{
		{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
	}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")

	// Adding one for a different Org gives an error
	err = as.testGqlQuery(query, f.Users[4].Nickname, &resp)
	as.Error(err, "expected an error (unauthorized) but didn't get one")
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")
}

func (as *ActionSuite) Test_GQLRemoveMeAsPotentialProvider() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	requests := f.Requests

	const qTemplate = `mutation {request: removeMeAsPotentialProvider (requestID: "%s")` +
		` {id title potentialProviders{id nickname}}}`

	var resp RequestResponse

	query := fmt.Sprintf(qTemplate, requests[1].UUID.String())

	err := as.testGqlQuery(query, f.Users[2].Nickname, &resp)
	as.NoError(err)
	as.Equal(requests[1].UUID.String(), resp.Request.ID, "incorrect Request UUID")
	as.Equal(requests[1].Title, resp.Request.Title, "incorrect Request title")

	want := []PotentialProvider{}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")
}

func (as *ActionSuite) Test_GQLRejectPotentialProvider() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	requests := f.Requests

	const qTemplate = `mutation {request: rejectPotentialProvider (requestID: "%s", userID: "%s")` +
		` {id title potentialProviders{id nickname}}}`

	var resp RequestResponse

	// remove third User as a potential provider on second Request
	query := fmt.Sprintf(qTemplate, requests[1].UUID.String(), f.Users[2].UUID.String())

	// Called by requester
	err := as.testGqlQuery(query, f.Users[0].Nickname, &resp)
	as.NoError(err)
	as.Equal(requests[1].UUID.String(), resp.Request.ID, "incorrect Request UUID")
	as.Equal(requests[1].Title, resp.Request.Title, "incorrect Request title")

	want := []PotentialProvider{{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}
	as.Equal(want, resp.Request.PotentialProviders, "incorrect potential providers")
}

func (as *ActionSuite) verifyPotentialProviders(expected models.Users, actual api.Users, msg string) {
	as.Equal(len(expected), len(actual), msg+", length is not correct")

	for i := range expected {
		as.verifyUser(expected[i], actual[i], fmt.Sprintf("%s, potential provider %d is not correct", msg, i))
	}
}

func (as *ActionSuite) Test_AddMeAsPotentialProvider() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	user := f.Users[1]

	noProviders := f.Requests[2]
	twoProviders := f.Requests[1]

	type testCase struct {
		name           string
		request        models.Request
		user           models.User
		wantHttpStatus int
		wantContains   []string
	}

	testCases := []testCase{
		{
			name:           "Wrong Organization",
			request:        twoProviders,
			user:           f.Users[4],
			wantHttpStatus: http.StatusNotFound,
			wantContains:   []string{api.ErrorGetRequest.String()},
		},
		{
			name:           "No Other Providers",
			request:        noProviders,
			user:           f.Users[1],
			wantHttpStatus: http.StatusOK,
			wantContains: []string{
				fmt.Sprintf(`{"id":"%s"`, noProviders.UUID),
				fmt.Sprintf(`"title":"%s"`, noProviders.Title),
				fmt.Sprintf(`"potential_providers":[{"id":"%s"`, user.UUID),
				fmt.Sprintf(`"nickname":"%s"`, user.Nickname),
			},
		},
		{
			name:           "Two Other Providers",
			request:        twoProviders,
			user:           f.Users[1],
			wantHttpStatus: http.StatusOK,
			wantContains: []string{
				fmt.Sprintf(`{"id":"%s"`, twoProviders.UUID),
				fmt.Sprintf(`"title":"%s"`, twoProviders.Title),
				fmt.Sprintf(`"potential_providers":[{"id":"%s"`, user.UUID),
				fmt.Sprintf(`"nickname":"%s"`, user.Nickname),
			},
		},
		{
			name:           "Repeat Provider Gives Error",
			request:        twoProviders,
			user:           f.Users[1],
			wantHttpStatus: http.StatusBadRequest,
			wantContains:   []string{api.ErrorAddPotentialProviderDuplicate.String()},
		},
	}

	for _, tc := range testCases {
		as.T().Run(tc.name, func(t *testing.T) {
			req := as.JSON("/requests/%s/potentialprovider", tc.request.UUID.String())
			req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tc.user.Nickname)
			req.Headers["content-type"] = "application/json"
			res := req.Post(nil)

			body := res.Body.String()
			as.Equal(tc.wantHttpStatus, res.Code, "incorrect status code returned, body: %s", body)
			as.verifyResponseData(tc.wantContains, body, "")
		})
	}
}

func (as *ActionSuite) Test_RejectPotentialProvider() {
	f := test.CreatePotentialProvidersFixtures(as.DB)

	// Called by requester
	requestID := f.Requests[1].UUID.String()
	requester := f.Users[0]
	pproviderUserID := f.Users[2].UUID.String()
	pprovider := f.PotentialProviders[3]

	req := as.JSON("/requests/%s/potentialprovider/%s", requestID, pproviderUserID)
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", requester.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Delete()

	body := res.Body.String()
	as.Equal(http.StatusNoContent, res.Code, "incorrect status code returned, body: %s", body)
	as.Empty(body, "incorrect body returned.")

	var dbPProvider models.PotentialProvider
	err := as.DB.Find(&dbPProvider, pprovider.ID)

	as.NotNil(err, "expected the PotentialProvider to be missing from the database")
	as.False(domain.IsOtherThanNoRows(err), "got unexpected error fetching PotentialProvider from database")
}

func (as *ActionSuite) Test_RemoveMeAsPotentialProvider() {
	f := test.CreatePotentialProvidersFixtures(as.DB)
	request := f.Requests[1]
	pprovider := f.PotentialProviders[3]

	req := as.JSON("/requests/%s/potentialprovider", request.UUID.String())
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", f.Users[2].Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Delete()

	body := res.Body.String()
	as.Equal(http.StatusNoContent, res.Code, "incorrect status code returned, body: %s", body)
	as.Empty(body, "incorrect body returned.")

	var dbPProvider models.PotentialProvider
	err := as.DB.Find(&dbPProvider, pprovider.ID)

	as.NotNil(err, "expected the PotentialProvider to be missing from the database")
	as.False(domain.IsOtherThanNoRows(err), "got unexpected error fetching PotentialProvider from database")
}
