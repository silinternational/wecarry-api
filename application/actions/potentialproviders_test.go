package actions

import (
	"fmt"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) Test_AddMeAsPotentialProvider() {
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

func (as *ActionSuite) Test_RemoveMeAsPotentialProvider() {
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

func (as *ActionSuite) Test_RejectPotentialProvider() {
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
