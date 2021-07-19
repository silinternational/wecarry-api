package actions

import (
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type messageQueryFixtures struct {
	models.Organization
	models.Users
	models.Requests
	models.Threads
	models.Messages
}

func createFixtures_MessageQuery(as *ActionSuite) messageQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	requests := test.CreateRequestFixtures(as.DB, 1, false, users[0].ID)

	threads := models.Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
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
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: users[1].ID,
			Content:  "Message from " + users[1].Nickname,
		},
		{
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Reply from " + users[0].Nickname,
		},
	}
	for i := range messages {
		createFixture(as, &messages[i])
	}

	return messageQueryFixtures{
		Organization: org,
		Users:        users,
		Requests:     requests,
		Threads:      threads,
		Messages:     messages,
	}
}
