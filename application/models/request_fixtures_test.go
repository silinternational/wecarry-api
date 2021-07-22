package models

import (
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type RequestFixtures struct {
	Users
	Requests
	RequestHistories
	Files
	Locations
	PotentialProviders
}

func CreateFixturesValidateUpdate_RequestStatus(status RequestStatus, ms *ModelSuite, t *testing.T) Request {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	request := Request{
		CreatedByID:    user.ID,
		OrganizationID: org.ID,
		DestinationID:  location.ID,
		Title:          "Test Request",
		Size:           RequestSizeMedium,
		UUID:           domain.GetUUID(),
		Status:         status,
	}

	createFixture(ms, &request)

	return request
}

func createFixturesForTestRequestCreate(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	requests := Requests{
		{UUID: domain.GetUUID(), Title: "title0"},
		{Title: "title1"},
		{},
	}
	locations := make(Locations, len(requests))
	for i := range requests {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		requests[i].Status = RequestStatusOpen
		requests[i].Size = RequestSizeTiny
		requests[i].CreatedByID = user.ID
		requests[i].OrganizationID = org.ID
		requests[i].DestinationID = locations[i].ID
	}
	createFixture(ms, &requests[2])

	return RequestFixtures{
		Users:    Users{user},
		Requests: requests,
	}
}

func createFixturesForTestRequestUpdate(ms *ModelSuite) RequestFixtures {
	requests := createRequestFixtures(ms.DB, 2, false)
	requests[0].Title = "new title"
	requests[1].Title = ""

	return RequestFixtures{
		Requests: requests,
	}
}

func createFixturesForTestRequest_manageStatusTransition_forwardProgression(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	requests := createRequestFixtures(ms.DB, 4, false, users[0].ID)
	requests[1].Status = RequestStatusAccepted
	requests[1].CreatedByID = users[1].ID
	requests[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&requests[1]))

	// Give the last Request an intermediate RequestHistory
	pHistory := RequestHistory{
		Status:    RequestStatusAccepted,
		RequestID: requests[3].ID,
	}
	mustCreate(ms.DB, &pHistory)

	// Give these new statuses while by-passing the status transition validation
	for i, status := range [2]RequestStatus{RequestStatusAccepted, RequestStatusDelivered} {
		index := i + 2

		requests[index].Status = status
		ms.NoError(ms.DB.Save(&requests[index]))
	}

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func createFixturesForTestRequest_manageStatusTransition_backwardProgression(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	requests := createRequestFixtures(ms.DB, 4, false, users[0].ID)

	// Put the first two requests into ACCEPTED status (also give them matching RequestHistory entries)
	requests[0].Status = RequestStatusAccepted
	requests[0].CreatedByID = users[0].ID
	requests[0].ProviderID = nulls.NewInt(users[1].ID)
	requests[1].Status = RequestStatusAccepted
	requests[1].CreatedByID = users[1].ID
	requests[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&requests))

	// add in a RequestHistory entry as if it had already happened
	pHistory := RequestHistory{
		Status:     RequestStatusAccepted,
		RequestID:  requests[2].ID,
		ReceiverID: nulls.NewInt(users[0].ID),
		ProviderID: nulls.NewInt(users[1].ID),
	}
	mustCreate(ms.DB, &pHistory)

	pHistory = RequestHistory{
		Status:     RequestStatusDelivered,
		RequestID:  requests[3].ID,
		ReceiverID: nulls.NewInt(users[0].ID),
		ProviderID: nulls.NewInt(users[1].ID),
	}
	mustCreate(ms.DB, &pHistory)

	for i := 2; i < 4; i++ {
		// AfterUpdate creates the new RequestHistory for this status
		requests[i].Status = RequestStatusCompleted
		requests[i].CompletedOn = nulls.NewTime(time.Now())
		requests[i].ProviderID = nulls.NewInt(users[1].ID)
		ms.NoError(ms.DB.Save(&requests[i]))
	}

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func CreateFixturesForRequestsGetFiles(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 1)
	organization := uf.Organization
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	request := Request{CreatedByID: user.ID, OrganizationID: organization.ID, DestinationID: location.ID}
	createFixture(ms, &request)

	files := createFileFixtures(ms.DB, 3)

	for i := range files {
		_, err := request.AttachFile(ms.DB, files[i].UUID.String())
		ms.NoError(err, "failed to attach file to request fixture")
	}

	return RequestFixtures{
		Users:    Users{user},
		Requests: Requests{request},
		Files:    files,
	}
}

//func createFixturesForRequestFindByUserAndUUID(ms *ModelSuite) RequestFixtures {
//	orgs := Organizations{{}, {}}
//	for i := range orgs {
//		orgs[i].UUID = domain.GetUUID()
//		orgs[i].AuthConfig = "{}"
//		createFixture(ms, &orgs[i])
//	}
//
//	users := createUserFixtures(ms.DB, 2).Users
//
//	// both users are in org 0, but need user 0 to also be in org 1
//	createFixture(ms, &UserOrganization{
//		OrganizationID: orgs[1].ID,
//		UserID:         users[0].ID,
//		AuthID:         users[0].Email,
//		AuthEmail:      users[0].Email,
//	})
//
//	requests := createRequestFixtures(ms.DB, 3, false, users[0].ID)
//	requests[1].OrganizationID = orgs[1].ID
//	requests[2].Status = RequestStatusRemoved
//	ms.NoError(ms.DB.Save(&requests))
//
//	return RequestFixtures{
//		Users: users,
//		Requests: requests,
//	}
//}

//        Org0                Org1           Org2
//        |  |                | | |          | |
//        |  +----+-----------+ | +----+-----+ +
//        |       |             |      |       |
//       User1  User0        User3   Trust   User2
//
// Org0: Request0 (SAME), Request2 (SAME, COMPLETED), Request3 (SAME, REMOVED), Request4 (SAME)
// Org1: Request1 (SAME)
// Org2: Request5 (ALL), Request6 (TRUSTED), Request7 (SAME)
//
func CreateFixtures_Requests_FindByUser(ms *ModelSuite) RequestFixtures {
	orgs := createOrganizationFixtures(ms.DB, 3)

	trusts := OrganizationTrusts{
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[2].ID},
		{PrimaryID: orgs[2].ID, SecondaryID: orgs[1].ID},
	}
	createFixture(ms, &trusts)

	users := createUserFixtures(ms.DB, 4).Users

	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	uo, err := users[2].FindUserOrganization(ms.DB, orgs[0])
	ms.NoError(err)
	uo.OrganizationID = orgs[2].ID
	ms.NoError(DB.UpdateColumns(&uo, "organization_id"))

	uo, err = users[3].FindUserOrganization(ms.DB, orgs[0])
	ms.NoError(err)
	uo.OrganizationID = orgs[1].ID
	ms.NoError(DB.UpdateColumns(&uo, "organization_id"))

	requests := createRequestFixtures(ms.DB, 8, false, users[0].ID)
	requests[1].OrganizationID = orgs[1].ID
	requests[2].Status = RequestStatusOpen
	requests[3].Status = RequestStatusRemoved
	requests[4].CreatedByID = users[1].ID
	requests[5].OrganizationID = orgs[2].ID
	requests[5].Visibility = RequestVisibilityAll
	requests[6].OrganizationID = orgs[2].ID
	requests[6].Visibility = RequestVisibilityTrusted
	requests[7].OrganizationID = orgs[2].ID
	ms.NoError(ms.DB.Save(&requests))

	// can't go directly to "completed"
	requests[2].Status = RequestStatusAccepted
	ms.NoError(ms.DB.Save(&requests[2]))
	requests[2].Status = RequestStatusCompleted
	ms.NoError(ms.DB.Save(&requests[2]))

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func createFixtures_Requests_FindByUser_SearchText(ms *ModelSuite) RequestFixtures {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	unique := domain.GetUUID().String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{
		{OrganizationID: orgs[0].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[1].ID, AuthID: users[1].Email, AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(ms, &userOrgs[i])
	}

	locations := make([]Location, 6)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	requests := Requests{
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Title: "With Match"},
		{
			CreatedByID: users[0].ID, OrganizationID: orgs[1].ID, Title: "MXtch In Description",
			Description: nulls.NewString("This has the lower case match in it."),
		},
		{
			CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: RequestStatusCompleted,
			Title: "With Match But Completed",
		},
		{
			CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: RequestStatusRemoved,
			Title: "With Match But Removed",
		},
		{CreatedByID: users[1].ID, OrganizationID: orgs[1].ID, Title: "User1 No MXtch"},
		{CreatedByID: users[1].ID, OrganizationID: orgs[1].ID, Title: "User1 With MATCH"},
	}

	for i := range requests {
		requests[i].UUID = domain.GetUUID()
		requests[i].DestinationID = locations[i].ID
		createFixture(ms, &requests[i])
	}

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func CreateFixtures_Request_IsEditable(ms *ModelSuite) RequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	requests := createRequestFixtures(ms.DB, 2, false, users[0].ID)
	requests[1].Status = RequestStatusRemoved

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func createFixturesForRequestGetAudience(ms *ModelSuite) RequestFixtures {
	orgs := make(Organizations, 2)
	for i := range orgs {
		orgs[i] = Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
		createFixture(ms, &orgs[i])
	}

	users := createUserFixtures(ms.DB, 2).Users

	requests := createRequestFixtures(ms.DB, 2, false, users[0].ID)
	requests[1].OrganizationID = orgs[1].ID
	ms.NoError(ms.DB.Save(&requests[1]))

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

// CreateFixtures_Request_AddUserAsPotentialProvider generates
//   five PotentialProvider records for testing.
// If necessary, five User and four Request fixtures will also be created.
//  (The fifth User will be with a different organization)
// The Requests will all be created by the first user.
// The first Request will have all but the first user as a potential provider.
// The second Request will have the last two users as potential providers.
// The third Request won't have any potential providers
// The fourth Request won't have any potential providers but will not be OPEN
func CreateFixtures_Request_AddUserAsPotentialProvider(ms *ModelSuite) potentialProvidersFixtures {
	uf := createUserFixtures(ms.DB, 5)

	extraOrg := Organization{AuthConfig: "{}"}
	mustCreate(ms.DB, &extraOrg)

	otherOrgUser := uf.UserOrganizations[4]
	otherOrgUser.OrganizationID = extraOrg.ID
	ms.NoError(ms.DB.Save(&otherOrgUser), "failed saving OrganizationUser with new Org")

	requests := createRequestFixtures(ms.DB, 4, false, uf.Users[0].ID)
	providers := PotentialProviders{}

	// ensure the first user is actually the creator (timing issues tend to make this unreliable otherwise)
	for i := range requests {
		requests[i].CreatedByID = uf.Users[0].ID
	}
	requests[3].Status = RequestStatusAccepted

	ms.DB.Update(&requests)

	for i, p := range requests[:2] {
		for _, u := range uf.Users[i+1:] {
			c := PotentialProvider{RequestID: p.ID, UserID: u.ID}
			c.Create(ms.DB)
			providers = append(providers, c)
		}
	}

	return potentialProvidersFixtures{
		Users:              uf.Users,
		Requests:           requests,
		PotentialProviders: providers,
	}
}
