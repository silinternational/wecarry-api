package models

import (
	"strconv"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type RequestHistoryFixtures struct {
	Users
	Requests
	RequestHistories
	Files
	Locations
}

func createFixturesForTestRequestHistory_Load(ms *ModelSuite) RequestHistoryFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	requests := createRequestFixtures(ms.DB, 2, false)

	pHistory := RequestHistory{
		Status:     RequestStatusOpen,
		RequestID:  requests[0].ID,
		ReceiverID: nulls.NewInt(requests[0].CreatedByID),
	}
	createFixture(ms, &pHistory)

	return RequestHistoryFixtures{
		Users:            users,
		Requests:         requests,
		RequestHistories: RequestHistories{pHistory},
	}
}

func createFixturesForTestRequestHistory_pop(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	requests := Requests{
		{Title: "Request1 Title", ProviderID: nulls.NewInt(users[1].ID)},
		{Title: "Request2 Title"},
	}
	locations := make(Locations, len(requests))
	for i := range requests {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		requests[i].UUID = domain.GetUUID()
		requests[i].Status = RequestStatusAccepted
		requests[i].Size = RequestSizeTiny
		requests[i].CreatedByID = users[0].ID
		requests[i].OrganizationID = org.ID
		requests[i].DestinationID = locations[i].ID
		createFixture(ms, &requests[i])
	}

	pHistories := RequestHistories{
		{
			Status:     RequestStatusOpen,
			RequestID:  requests[0].ID,
			ReceiverID: nulls.NewInt(requests[0].CreatedByID),
		},
		{
			Status:     RequestStatusAccepted,
			RequestID:  requests[0].ID,
			ReceiverID: nulls.NewInt(requests[0].CreatedByID),
			ProviderID: nulls.NewInt(users[1].ID),
		},
	}

	for i := range pHistories {
		createFixture(ms, &pHistories[i])
	}

	return RequestFixtures{
		Users:            users,
		Requests:         requests,
		RequestHistories: pHistories,
	}
}

func createFixturesForTestRequestHistory_createForRequest(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	requests := createRequestFixtures(ms.DB, 2, false)

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}
