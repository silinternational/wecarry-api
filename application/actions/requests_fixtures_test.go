package actions

import (
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type UpdateRequestStatusFixtures struct {
	models.Requests
	models.Users
}

func createFixturesForRequests(as *ActionSuite) RequestFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	org2 := models.Organization{Name: "org2", AuthType: AuthTypeGoogle, AuthConfig: "{}"}
	as.NoError(org2.Save(as.DB))

	requests := test.CreateRequestFixtures(as.DB, 4, true, users[0].ID)
	requests[0].Status = models.RequestStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)

	requests[2].Status = models.RequestStatusCompleted
	requests[2].CompletedOn = nulls.NewTime(time.Now())
	requests[2].ProviderID = nulls.NewInt(users[1].ID)

	requests[3].OrganizationID = org2.ID

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

	meetings := test.CreateMeetingFixtures(as.DB, 1, users[0])

	return RequestFixtures{
		Organization: org,
		Users:        users,
		Requests:     requests,
		Threads:      threads,
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
