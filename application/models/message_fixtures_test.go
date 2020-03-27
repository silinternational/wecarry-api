package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
)

type MessageFixtures struct {
	Users
	Messages
	Threads
	ThreadParticipants
	Requests
}

func Fixtures_Message_GetSender(ms *ModelSuite, t *testing.T) MessageFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	createFixture(ms, &requests[0])

	threads := Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	createFixture(ms, &threads[0])

	messages := Messages{
		{
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
	}
	createFixture(ms, &messages[0])

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}

}

// Request 0: Org A (no threads)
// Request 1: Org B
//     Thread 0: User 0
//     Thread 1: User 0, User 2
//     Thread 2: (no thread participants)
// Request 2: Org A (no threads)
func Fixtures_Message_Create(ms *ModelSuite, t *testing.T) MessageFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users
	requests := createRequestFixtures(ms.DB, 3, false)

	org := createOrganizationFixtures(ms.DB, 1)[0]
	requests[0].OrganizationID = org.ID
	ms.NoError(requests[0].Update())

	tf := CreateThreadFixtures(ms, requests[1])

	return MessageFixtures{
		Users:    append(users, tf.Users...),
		Threads:  tf.Threads,
		Requests: requests,
	}
}

func Fixtures_Message_FindByID(ms *ModelSuite, t *testing.T) MessageFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	createFixture(ms, &requests[0])

	threads := Threads{
		{UUID: domain.GetUUID(), RequestID: requests[0].ID},
	}
	createFixture(ms, &threads[0])

	messages := Messages{
		{
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
	}
	createFixture(ms, &messages[0])

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}
}

// Fixtures_Message_Find is used by TestMessage_FindByUUID and TestMessage_FindByUserAndUUID
func Fixtures_Message_Find(ms *ModelSuite) MessageFixtures {
	uf := createUserFixtures(ms.DB, 5)
	org := uf.Organization
	users := uf.Users

	roles := []UserAdminRole{
		UserAdminRoleUser,
		UserAdminRoleUser,
		UserAdminRoleAdmin,
		UserAdminRoleSalesAdmin,
		UserAdminRoleSuperAdmin,
	}
	for i := range users {
		users[i].AdminRole = roles[i]
		ms.NoError(ms.DB.Save(&users[i]))
	}

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	createFixture(ms, &requests[0])

	threads := make(Threads, 2)
	for i := range threads {
		threads[i].UUID = domain.GetUUID()
		threads[i].RequestID = requests[0].ID
		createFixture(ms, &threads[i])
	}

	threadParticipants := ThreadParticipants{
		{ThreadID: threads[0].ID, UserID: users[1].ID},
		{ThreadID: threads[0].ID, UserID: users[2].ID},
		{ThreadID: threads[0].ID, UserID: users[3].ID},
		{ThreadID: threads[0].ID, UserID: users[4].ID},
	}
	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	messages := Messages{
		{ThreadID: threads[0].ID, Content: "Love must be sincere. Hate what is evil; cling to what is good."},
		{ThreadID: threads[1].ID, Content: "Love your neighbor as yourself."},
	}
	for i := range messages {
		messages[i].UUID = domain.GetUUID()
		messages[i].SentByID = users[0].ID
		createFixture(ms, &messages[i])
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
	}
}

func CreateMessageFixtures_AfterCreate(ms *ModelSuite, t *testing.T) MessageFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	// Each user has a request and is a provider on the other user's request
	requests := createRequestFixtures(ms.DB, 2, false)
	requests[0].Status = RequestStatusAccepted
	requests[0].ProviderID = nulls.NewInt(users[1].ID)
	requests[1].Status = RequestStatusAccepted
	requests[1].CreatedByID = users[1].ID
	requests[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&requests))

	tenMinutesAgo := time.Now().Add(-10 * time.Minute)
	threads := make([]Thread, 2)
	for i, thread := range threads {
		thread.UUID = domain.GetUUID()
		err := ms.DB.RawQuery(`INSERT INTO threads (created_at, updated_at, uuid, request_id) VALUES (?, ?, ?, ?)`,
			tenMinutesAgo, tenMinutesAgo, thread.UUID, requests[i].ID).Exec()
		ms.NoError(err, "error loading threads")

		// get the record back from the database
		err = ms.DB.Where("uuid = ?", thread.UUID.String()).First(&threads[i])
		ms.NoError(err, "error finding thread fixture %s, %s", thread.UUID.String(), err)
	}
	tNow := time.Now().Round(time.Duration(time.Second))
	oldTime := tNow.Add(-time.Duration(time.Hour))
	oldOldTime := oldTime.Add(-time.Duration(time.Hour))

	// One thread per request with 2 users per thread
	threadParticipants := []ThreadParticipant{
		{
			ThreadID:     threads[0].ID,
			UserID:       requests[0].CreatedByID,
			LastViewedAt: tNow, // This will get overridden and then reset again
		},
		{
			ThreadID:     threads[0].ID,
			UserID:       requests[0].ProviderID.Int,
			LastViewedAt: oldTime,
		},
		{
			ThreadID:     threads[1].ID,
			UserID:       requests[1].CreatedByID,
			LastViewedAt: oldTime,
		},
		{
			ThreadID:     threads[1].ID,
			UserID:       requests[1].ProviderID.Int,
			LastViewedAt: tNow,
		},
	}

	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	// I can't seem to give them custom times
	messages := Messages{
		{
			ThreadID:  threads[0].ID,           // user 0's request
			SentByID:  requests[0].CreatedByID, // user 0 (Eager)
			Content:   "I can being chocolate if you bring PB",
			CreatedAt: oldOldTime,
		},
		{
			ThreadID:  threads[0].ID,              // user 0's request
			SentByID:  requests[0].ProviderID.Int, // user 1 (Lazy)
			Content:   "Great",
			CreatedAt: oldTime,
		},
		{
			ThreadID:  threads[0].ID,           // user 0's request
			SentByID:  requests[0].CreatedByID, // user 0 (Eager)
			Content:   "Can you get it here by next week?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
		{
			ThreadID:  threads[1].ID,           // user 1's request
			SentByID:  requests[1].CreatedByID, // user 1 (Lazy)
			Content:   "I can being PB if you bring chocolate",
			CreatedAt: oldTime,
		},
		{
			ThreadID:  threads[1].ID,              // user 1's request
			SentByID:  requests[1].ProviderID.Int, // user 0 (Eager)
			Content:   "Did you see my other message?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
		{
			ThreadID:  threads[1].ID,              // user 1's request
			SentByID:  requests[1].ProviderID.Int, // user 0 (Eager)
			Content:   "Anyone Home?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
	}

	for _, m := range messages {
		m.UUID = domain.GetUUID()
		if err := ms.DB.RawQuery(`INSERT INTO messages (content, created_at, sent_by_id, thread_id, updated_at, uuid)`+
			`VALUES (?, ?, ?, ?, ?, ?)`,
			m.Content, m.CreatedAt, m.SentByID, m.ThreadID, m.CreatedAt, m.UUID).Exec(); err != nil {
			t.Errorf("error loading messages ... %v", err)
			t.FailNow()
		}
	}

	return MessageFixtures{
		Users:              users,
		Threads:            threads,
		Messages:           messages,
		ThreadParticipants: threadParticipants,
	}
}
