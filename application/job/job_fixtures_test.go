package job

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type MessageFixtures struct {
	models.Users
	models.Messages
	models.Threads
}
type RequestFixtures struct {
	models.Users
	models.Requests
}

func createFixture(js *JobSuite, f interface{}) {
	err := js.DB.Create(f)
	if err != nil {
		js.T().Errorf("error creating %T fixture, %s", f, err)
		js.T().FailNow()
	}
}

func CreateFixtures_TestOutdatedRequestHandler(js *JobSuite) RequestFixtures {
	users := test.CreateUserFixtures(js.DB, 2).Users
	requests := test.CreateRequestFixtures(js.DB, 2, false, users[0].ID, users[1].ID)
	db := js.DB

	requests[0].Status = models.RequestStatusAccepted
	js.NoError(requests[0].Update(db), "error modifying request for test")

	pastDate := nulls.NewTime(time.Now().Add(-1 * domain.DurationDay))
	requests[1].NeededBefore = pastDate
	js.NoError(js.DB.Update(&requests[1]), "error modifying request for test")

	return RequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func CreateFixtures_TestNewThreadMessageHandler(js *JobSuite) MessageFixtures {
	uf := test.CreateUserFixtures(js.DB, 7)
	org := uf.Organization
	users := uf.Users

	location := models.Location{}
	createFixture(js, &location)

	requests := models.Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	for i := range requests {
		createFixture(js, &requests[i])
	}

	threads := models.Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	for i := range threads {
		createFixture(js, &threads[i])
	}

	threadParticipants := models.ThreadParticipants{
		// Thread 0; Recipient = User1; viewed before last notification
		{ThreadID: threads[0].ID, UserID: users[0].ID},
		{
			ThreadID:       threads[0].ID,
			UserID:         users[1].ID,
			LastNotifiedAt: time.Now().Add(-2 * time.Minute),
			LastViewedAt:   time.Now().Add(-4 * time.Minute),
		},

		// Thread 1; Recipient = User2; viewed before last notification
		{ThreadID: threads[1].ID, UserID: users[0].ID},
		{
			ThreadID:       threads[1].ID,
			UserID:         users[2].ID,
			LastNotifiedAt: time.Now().Add(-2 * time.Minute),
			LastViewedAt:   time.Now().Add(-4 * time.Minute),
		},

		// Thread 2; Recipient = User3; viewed before last notification
		{ThreadID: threads[2].ID, UserID: users[0].ID},
		{
			ThreadID:       threads[2].ID,
			UserID:         users[3].ID,
			LastNotifiedAt: time.Now().Add(-2 * time.Minute),
			LastViewedAt:   time.Now().Add(-4 * time.Minute),
		},

		// Thread 3; Recipient = User4; viewed after last notification
		{ThreadID: threads[3].ID, UserID: users[0].ID},
		{
			ThreadID:       threads[3].ID,
			UserID:         users[4].ID,
			LastNotifiedAt: time.Now().Add(-4 * time.Minute),
			LastViewedAt:   time.Now().Add(-2 * time.Minute),
		},

		// Thread 4; Recipient = User5; viewed after last notification
		{ThreadID: threads[4].ID, UserID: users[0].ID},
		{
			ThreadID:       threads[4].ID,
			UserID:         users[5].ID,
			LastNotifiedAt: time.Now().Add(-4 * time.Minute),
			LastViewedAt:   time.Now().Add(-2 * time.Minute),
		},

		// Thread 5; Recipient = User6; viewed after last notification
		{ThreadID: threads[5].ID, UserID: users[0].ID},
		{
			ThreadID:       threads[5].ID,
			UserID:         users[6].ID,
			LastNotifiedAt: time.Now().Add(-4 * time.Minute),
			LastViewedAt:   time.Now().Add(-2 * time.Minute),
		},
	}
	for i := range threadParticipants {
		createFixture(js, &threadParticipants[i])
	}

	messages := models.Messages{
		{
			UUID:      domain.GetUUID(),
			ThreadID:  threads[0].ID,
			SentByID:  users[0].ID,
			UpdatedAt: time.Now().Add(-1 * time.Minute),
			Content:   "New message, last_viewed_at < last_notified_at < message updated_at",
		},
		{
			UUID:      domain.GetUUID(),
			ThreadID:  threads[1].ID,
			SentByID:  users[0].ID,
			UpdatedAt: time.Now().Add(-3 * time.Minute),
			Content:   "New message, last_viewed_at < message updated_at < last_notified_at",
		},
		{
			UUID:      domain.GetUUID(),
			ThreadID:  threads[2].ID,
			SentByID:  users[0].ID,
			UpdatedAt: time.Now().Add(-5 * time.Minute),
			Content:   "New message, message updated_at < last_viewed_at < last_notified_at",
		},
		{
			UUID:      domain.GetUUID(),
			ThreadID:  threads[3].ID,
			SentByID:  users[0].ID,
			UpdatedAt: time.Now().Add(-1 * time.Minute),
			Content:   "New message, last_notified_at < last_viewed_at < message updated_at",
		},
		{
			UUID:      domain.GetUUID(),
			ThreadID:  threads[4].ID,
			SentByID:  users[0].ID,
			UpdatedAt: time.Now().Add(-3 * time.Minute),
			Content:   "New message, last_notified_at < message updated_at < last_viewed_at",
		},
		{
			UUID:      domain.GetUUID(),
			ThreadID:  threads[5].ID,
			SentByID:  users[0].ID,
			UpdatedAt: time.Now().Add(-5 * time.Minute),
			Content:   "New message, message updated_at < last_notified_at < last_viewed_at",
		},
	}
	for i, m := range messages {
		// manually create records to bypass automatic updated_at code
		err := js.DB.RawQuery(`INSERT INTO messages (content, created_at, sent_by_id, thread_id, updated_at, uuid)
			VALUES (?, ?, ?, ?, ?, ?)`, m.Content, m.UpdatedAt, m.SentByID, m.ThreadID, m.UpdatedAt, m.UUID).Exec()
		if err != nil {
			js.T().Errorf("error creating message fixture, %s", err)
			js.T().FailNow()
		}

		// get the new message ID
		var m2 models.Message
		err = js.DB.Where("uuid = ?", m.UUID.String()).First(&m2)
		if err != nil {
			js.T().Errorf("error finding message fixture %s, %s", m.UUID.String(), err)
			js.T().FailNow()
		}
		messages[i].ID = m2.ID
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}
}
