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
	models.Posts
	models.Users
	models.Files
	models.Locations
}

type CreatePostFixtures struct {
	models.Users
	models.Organization
	models.File
	models.Meetings
}

type UpdateRequestStatusFixtures struct {
	models.Posts
	models.Users
}

func createFixturesForRequestQuery(as *ActionSuite) RequestQueryFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreatePostFixtures(as.DB, 3, true)
	requests[0].Status = models.PostStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Save(&requests[0]))
	as.NoError(requests[0].SetDestination(models.Location{Description: "Australia", Country: "AU"}))
	as.NoError(requests[1].SetOrigin(models.Location{Description: "Australia", Country: "AU"}))

	requests[2].Status = models.PostStatusCompleted
	requests[2].CompletedOn = nulls.NewTime(time.Now())
	requests[2].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Save(&requests[2]))

	threads := []models.Thread{
		{UUID: domain.GetUUID(), PostID: requests[0].ID},
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

	var fileFixture models.File
	as.Nil(fileFixture.Store("dummy.pdf", []byte("%PDF-")), "failed to create file fixture")

	if _, err := requests[0].AttachFile(fileFixture.UUID.String()); err != nil {
		t.Errorf("failed to attach file to request, %s", err)
		t.FailNow()
	}

	return RequestQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        requests,
		Threads:      threads,
	}
}

func createFixturesForSearchRequestsQuery(as *ActionSuite) RequestQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreatePostFixtures(as.DB, 2, false)
	requests[0].Title = "A Match"
	as.NoError(as.DB.Save(&requests[0]))

	return RequestQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        requests,
	}
}

func createFixturesForUpdateRequest(as *ActionSuite) UpdateRequestFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreatePostFixtures(as.DB, 1, false)
	requests[0].OriginID = nulls.Int{}
	as.NoError(as.DB.Save(&requests[0]))

	var fileFixture models.File
	as.Nil(fileFixture.Store("new_photo.webp", []byte("RIFFxxxxWEBPVP")), "failed to create file fixture")

	return UpdateRequestFixtures{
		Posts: requests,
		Users: users,
		Files: models.Files{fileFixture},
	}
}

func createFixturesForCreateRequest(as *ActionSuite) CreatePostFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, 1)
	org := userFixtures.Organization

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var fileFixture models.File
	as.Nil(fileFixture.Store("photo.gif", []byte("GIF89a")), "failed to create file fixture")

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

	return CreatePostFixtures{
		Users:        userFixtures.Users,
		Organization: org,
		File:         fileFixture,
		Meetings:     meetings,
	}
}

func createFixturesForUpdateRequestStatus(as *ActionSuite) UpdateRequestStatusFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreatePostFixtures(as.DB, 1, false)

	return UpdateRequestStatusFixtures{
		Posts: requests,
		Users: users,
	}
}

func createFixturesForMarkRequestAsDelivered(as *ActionSuite) UpdateRequestStatusFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreatePostFixtures(as.DB, 2, false)
	requests[0].Status = models.PostStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)

	requests[1].Status = models.PostStatusCompleted
	requests[1].ProviderID = nulls.NewInt(users[1].ID)

	as.NoError(as.DB.Update(&requests))

	as.DB.Save(&requests[0])

	return UpdateRequestStatusFixtures{
		Posts: requests,
		Users: users,
	}
}

func createFixturesForMarkRequestAsReceived(as *ActionSuite) UpdateRequestStatusFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	requests := test.CreatePostFixtures(as.DB, 3, false)
	requests[0].Status = models.PostStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)

	// Get the Accepted RequestHistory added
	requests[1].Status = models.PostStatusAccepted
	requests[1].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Update(&requests))

	requests[1].Status = models.PostStatusDelivered

	requests[2].Status = models.PostStatusCompleted
	requests[2].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Update(&requests))

	as.DB.Save(&requests[0])

	return UpdateRequestStatusFixtures{
		Posts: requests,
		Users: users,
	}
}
