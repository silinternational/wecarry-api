package models

import (
	"fmt"
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
}

func Fixtures_Message_GetSender(ms *ModelSuite, t *testing.T) MessageFixtures {
	org := &Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, org)

	unique := domain.GetUuid().String()
	users := Users{
		{Email: unique + "user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
		{Email: unique + "user2@example.com", Nickname: unique + "User2", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: org.ID},
	}
	for i := range posts {
		createFixture(ms, &posts[i])
	}

	threads := Threads{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
	}
	for i := range messages {
		createFixture(ms, &messages[i])
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}

}

func Fixtures_Message_Create(ms *ModelSuite, t *testing.T) MessageFixtures {
	org := &Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, org)

	unique := domain.GetUuid().String()
	users := Users{
		{Email: unique + "user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
		{Email: unique + "user2@example.com", Nickname: unique + "User2", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: org.ID},
	}
	for i := range posts {
		createFixture(ms, &posts[i])
	}

	threads := Threads{
		{
			CreatedAt: time.Now().Add(-10 * time.Minute),
			UpdatedAt: time.Now().Add(-10 * time.Minute),
			Uuid:      domain.GetUuid(),
			PostID:    posts[0].ID,
		},
	}
	for i, thread := range threads {
		if err := ms.DB.RawQuery(`INSERT INTO threads (created_at, updated_at, uuid, post_id) VALUES (?, ?, ?, ?)`,
			thread.CreatedAt, thread.UpdatedAt, thread.Uuid, thread.PostID).Exec(); err != nil {
			t.Errorf("error loading threads, %v", err)
			t.FailNow()
		}

		// get the new thread ID
		err := ms.DB.Where("uuid = ?", thread.Uuid.String()).First(&threads[i])
		if err != nil {
			ms.T().Errorf("error finding thread fixture %s, %s", thread.Uuid.String(), err)
			ms.T().FailNow()
		}
	}

	return MessageFixtures{
		Users:   users,
		Threads: threads,
	}
}

func Fixtures_Message_FindByID(ms *ModelSuite, t *testing.T) MessageFixtures {
	org := &Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, org)

	unique := domain.GetUuid().String()
	users := Users{
		{Email: unique + "user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
		{Email: unique + "user2@example.com", Nickname: unique + "User2", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: org.ID},
	}
	for i := range posts {
		createFixture(ms, &posts[i])
	}

	threads := Threads{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
	}
	for i := range messages {
		createFixture(ms, &messages[i])
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}
}

func Fixtures_Message_FindByUUID(ms *ModelSuite) MessageFixtures {
	org := &Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, org)

	unique := domain.GetUuid().String()
	users := Users{
		{Email: unique + "user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: org.ID},
	}
	for i := range posts {
		createFixture(ms, &posts[i])
	}

	threads := Threads{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Love must be sincere. Hate what is evil; cling to what is good.",
		},
	}
	for i := range messages {
		createFixture(ms, &messages[i])
	}

	return MessageFixtures{
		Messages: messages,
	}
}

func CreateMessageFixtures_AfterCreate(ms *ModelSuite, t *testing.T) MessageFixtures {

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	org := Organization{
		Name:       fmt.Sprintf("ACME-%s", unique),
		Uuid:       domain.GetUuid(),
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
	}

	createFixture(ms, &org)

	// Load User test fixtures
	users := Users{
		{
			Email:     fmt.Sprintf("user1-%s@example.com", unique),
			FirstName: "Eager",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Eager User %s", unique),
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     fmt.Sprintf("user2-%s@example.com", unique),
			FirstName: "Lazy",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Lazy User %s", unique),
			Uuid:      domain.GetUuid(),
		},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         users[0].Email,
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         users[1].Email,
			AuthEmail:      users[1].Email,
		},
	}
	for i := range userOrgs {
		createFixture(ms, &userOrgs[i])
	}

	// Each user has a request and is a provider on the other user's post
	posts := Posts{
		{
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Open Request 0",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Committed Request 1",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[0].ID),
		},
	}

	for i := range posts {
		createFixture(ms, &posts[i])
	}

	threads := []Thread{
		{
			Uuid:   domain.GetUuid(),
			PostID: posts[0].ID,
		},
		{
			Uuid:   domain.GetUuid(),
			PostID: posts[1].ID,
		},
	}

	for i := range threads {
		createFixture(ms, &threads[i])
	}

	tNow := time.Now().Round(time.Duration(time.Second))
	oldTime := tNow.Add(-time.Duration(time.Hour))
	oldOldTime := oldTime.Add(-time.Duration(time.Hour))

	// One thread per post with 2 users per thread
	threadParticipants := []ThreadParticipant{
		{
			ThreadID:     threads[0].ID,
			UserID:       posts[0].CreatedByID,
			LastViewedAt: tNow, // This will get overridden and then reset again
		},
		{
			ThreadID:     threads[0].ID,
			UserID:       posts[0].ProviderID.Int,
			LastViewedAt: oldTime,
		},
		{
			ThreadID:     threads[1].ID,
			UserID:       posts[1].CreatedByID,
			LastViewedAt: oldTime,
		},
		{
			ThreadID:     threads[1].ID,
			UserID:       posts[1].ProviderID.Int,
			LastViewedAt: tNow,
		},
	}

	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	// I can't seem to give them custom times
	messages := Messages{
		{
			Uuid:      domain.GetUuid(),
			ThreadID:  threads[0].ID,        // user 0's post
			SentByID:  posts[0].CreatedByID, // user 0 (Eager)
			Content:   "I can being chocolate if you bring PB",
			CreatedAt: oldOldTime,
		},
		{
			Uuid:      domain.GetUuid(),
			ThreadID:  threads[0].ID,           // user 0's post
			SentByID:  posts[0].ProviderID.Int, // user 1 (Lazy)
			Content:   "Great",
			CreatedAt: oldTime,
		},
		{
			Uuid:      domain.GetUuid(),
			ThreadID:  threads[0].ID,        // user 0's post
			SentByID:  posts[0].CreatedByID, // user 0 (Eager)
			Content:   "Can you get it here by next week?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
		{
			Uuid:      domain.GetUuid(),
			ThreadID:  threads[1].ID,        // user 1's post
			SentByID:  posts[1].CreatedByID, // user 1 (Lazy)
			Content:   "I can being PB if you bring chocolate",
			CreatedAt: oldTime,
		},
		{
			Uuid:      domain.GetUuid(),
			ThreadID:  threads[1].ID,           // user 1's post
			SentByID:  posts[1].ProviderID.Int, // user 0 (Eager)
			Content:   "Did you see my other message?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
		{
			Uuid:      domain.GetUuid(),
			ThreadID:  threads[1].ID,           // user 1's post
			SentByID:  posts[1].ProviderID.Int, // user 0 (Eager)
			Content:   "Anyone Home?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
	}

	for _, m := range messages {
		if err := ms.DB.RawQuery(`INSERT INTO messages (content, created_at, sent_by_id, thread_id, updated_at, uuid)`+
			`VALUES (?, ?, ?, ?, ?, ?)`,
			m.Content, m.CreatedAt, m.SentByID, m.ThreadID, m.CreatedAt, m.Uuid).Exec(); err != nil {
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
