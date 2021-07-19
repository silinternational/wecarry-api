package actions

import (
	"time"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type threadTestFixtures struct {
	models.Organization
	models.Users
	models.Requests
	models.Threads
	models.Messages
}

func createFixturesForThreads(as *ActionSuite) threadTestFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 3)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 2, false, users[0].ID)

	threads := models.Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[1].ID},
	}
	for i := range threads {
		createFixture(as, &threads[i])
	}

	threadParticipants := models.ThreadParticipants{
		{ThreadID: threads[0].ID, UserID: requests[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(as, &threadParticipants[i])
	}

	messages := models.Messages{
		{
			ThreadID: threads[0].ID,
			SentByID: users[1].ID,
			Content:  "Message from " + users[1].Nickname,
		},
		{
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Reply from " + users[0].Nickname,
		},
	}
	for i := range messages {
		messages[i].UUID = domain.GetUUID()
		createFixture(as, &messages[i])
	}

	// Make sure the unread Message count is doing something
	yesterday := time.Now().Add(-time.Hour * 48)
	threadParticipants[0].UpdateLastViewedAt(as.DB, yesterday)

	return threadTestFixtures{
		Organization: org,
		Users:        users,
		Requests:     requests,
		Threads:      threads,
		Messages:     messages,
	}
}
