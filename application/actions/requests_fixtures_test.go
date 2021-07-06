package actions

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type UpdateRequestFixtures struct {
	models.Requests
	models.Users
	models.Files
	models.Locations
}

type CreateRequestFixtures struct {
	models.Users
	models.Organization
	models.File
	models.Meetings
}

type UpdateRequestStatusFixtures struct {
	models.Requests
	models.Users
}

type RequestsListFixtures struct {
	models.Requests
	models.Users
}

func createFixturesForRequestQuery(as *ActionSuite) RequestQueryFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 3, true)
	requests[0].Status = models.RequestStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Save(&requests))

	requests[2].Status = models.RequestStatusCompleted
	requests[2].CompletedOn = nulls.NewTime(time.Now())
	requests[2].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Save(&requests))

	threads := []models.Thread{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	for i := range threads {
		createFixture(as, &threads[i])
	}

	threadParticipants := []models.ThreadParticipant{
		{ThreadID: threads[0].ID, UserID: requests[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(as, &threadParticipants[i])
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	fileFixture := test.CreateFileFixture(as.DB)

	if _, err := requests[0].AttachFile(as.DB, fileFixture.UUID.String()); err != nil {
		t.Errorf("failed to attach file to request, %s", err)
		t.FailNow()
	}

	return RequestQueryFixtures{
		Organization: org,
		Users:        users,
		Requests:     requests,
		Threads:      threads,
	}
}

func createFixturesForRequestsList(as *ActionSuite) RequestsListFixtures {
	usersFixtures := test.CreateUserFixtures(as.DB, 3)
	requests := test.CreateRequestFixtures(as.DB, 5, false)

	requests[0].Status = models.RequestStatusAccepted
	requests[0].ProviderID = nulls.NewInt(usersFixtures.Users[1].ID)
	as.NoError(as.DB.Save(&requests[0]))

	requests[1].Status = models.RequestStatusCompleted
	requests[1].CompletedOn = nulls.NewTime(time.Now())
	requests[1].ProviderID = nulls.NewInt(usersFixtures.Users[2].ID)
	as.NoError(as.DB.Save(&requests[1]))

	photo := test.CreateFileFixture(as.DB)
	requests[0].FileID = nulls.NewInt(photo.ID)
	as.NoError(as.DB.Save(&requests[0]))

	return RequestsListFixtures{
		Users:    usersFixtures.Users,
		Requests: requests,
	}
}

func createFixturesForSearchRequestsQuery(as *ActionSuite) RequestQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 2, false)
	requests[0].Title = "A Match"
	as.NoError(as.DB.Save(&requests[0]))

	return RequestQueryFixtures{
		Organization: org,
		Users:        users,
		Requests:     requests,
	}
}

func createFixturesForUpdateRequest(as *ActionSuite) UpdateRequestFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 1, false)
	requests[0].OriginID = nulls.Int{}
	as.NoError(as.DB.Save(&requests[0]))

	fileFixture := test.CreateFileFixture(as.DB)

	return UpdateRequestFixtures{
		Requests: requests,
		Users:    users,
		Files:    models.Files{fileFixture},
	}
}

func createFixturesForCreateRequest(as *ActionSuite) CreateRequestFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, 1)
	org := userFixtures.Organization

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	fileFixture := test.CreateFileFixture(as.DB)

	meetingLocations := test.CreateLocationFixtures(as.DB, 1)

	meetings := make(models.Meetings, 1)
	for i := range meetings {
		iString := strconv.Itoa(i)
		start := time.Now().Add(time.Duration(rand.Intn(10000)) * time.Hour)
		meetings[i] = models.Meeting{
			Name:        "Meeting " + iString,
			Description: nulls.NewString("Meeting Description " + iString),
			MoreInfoURL: nulls.NewString("https://example.com/meeting/" + iString),
			StartDate:   start,
			EndDate:     start.Add(time.Duration(rand.Intn(200)) * time.Hour),
			CreatedByID: userFixtures.Users[0].ID,
			FileID:      nulls.Int{},
			LocationID:  meetingLocations[i].ID,
		}
		test.MustCreate(as.DB, &meetings[i])
	}

	return CreateRequestFixtures{
		Users:        userFixtures.Users,
		Organization: org,
		File:         fileFixture,
		Meetings:     meetings,
	}
}

func createFixturesForUpdateRequestStatus(as *ActionSuite) UpdateRequestStatusFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 1, false, users[0].ID)

	return UpdateRequestStatusFixtures{
		Requests: requests,
		Users:    users,
	}
}

func createFixturesForMarkRequestAsDelivered(as *ActionSuite) UpdateRequestStatusFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 2, false)
	requests[0].Status = models.RequestStatusAccepted
	requests[0].CreatedByID = users[0].ID
	requests[0].ProviderID = nulls.NewInt(users[1].ID)

	requests[1].Status = models.RequestStatusCompleted
	requests[1].CreatedByID = users[1].ID
	requests[1].ProviderID = nulls.NewInt(users[0].ID)

	as.NoError(as.DB.Update(&requests))

	as.DB.Save(&requests[0])

	return UpdateRequestStatusFixtures{
		Requests: requests,
		Users:    users,
	}
}

func createFixturesForMarkRequestAsReceived(as *ActionSuite) UpdateRequestStatusFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 3, false)
	requests[0].Status = models.RequestStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)

	// Get the Accepted RequestHistory added
	requests[1].Status = models.RequestStatusAccepted
	requests[1].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Update(&requests))

	requests[1].Status = models.RequestStatusDelivered

	requests[2].Status = models.RequestStatusCompleted
	requests[2].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Update(&requests))

	as.DB.Save(&requests[0])

	return UpdateRequestStatusFixtures{
		Requests: requests,
		Users:    users,
	}
}
