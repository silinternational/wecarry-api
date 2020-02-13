package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type ThreadFixtures struct {
	Users
	Posts
	Threads
	ThreadParticipants
	Messages
}

func CreateThreadFixtures(ms *ModelSuite, post Post) ThreadFixtures {
	// need another User for these fixtures, to act as the Provider and 2nd Thread Participant
	uf := createUserFixtures(ms.DB, 1)

	post.Status = PostStatusAccepted
	post.ProviderID = nulls.NewInt(uf.Users[0].ID)
	ms.NoError(ms.DB.Save(&post))

	// Load Thread test fixtures
	threads := make(Threads, 3)
	for i := range threads {
		threads[i].UUID = domain.GetUUID()
		threads[i].PostID = post.ID
		createFixture(ms, &threads[i])
	}

	// Load Thread Participants test fixtures
	threadParticipants := []ThreadParticipant{
		{
			ThreadID:     threads[0].ID,
			UserID:       post.CreatedByID,
			LastViewedAt: time.Now(),
		},
		{
			ThreadID: threads[1].ID,
			UserID:   uf.Users[0].ID,
		},
		{
			ThreadID: threads[1].ID,
			UserID:   post.CreatedByID,
		},
	}
	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	// Load Message test fixtures
	messages := Messages{
		{
			ThreadID: threads[0].ID,
			SentByID: post.CreatedByID,
			Content:  "I can being chocolate if you bring PB",
		},
		{
			ThreadID: threads[1].ID,
			SentByID: uf.Users[0].ID,
			Content:  "I can being PB if you bring chocolate",
		},
		{
			ThreadID: threads[1].ID,
			SentByID: post.CreatedByID,
			Content:  "Great!",
		},
	}

	for i := range messages {
		messages[i].UUID = domain.GetUUID()
		createFixture(ms, &messages[i])
	}

	return ThreadFixtures{
		Threads:            threads,
		Messages:           messages,
		ThreadParticipants: threadParticipants,
		Users:              uf.Users,
	}
}

func CreateThreadFixtures_UnreadMessageCount(ms *ModelSuite, t *testing.T) ThreadFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	// Each user has a request and is a provider on the other user's post
	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[0].Status = PostStatusAccepted
	posts[0].ProviderID = nulls.NewInt(users[1].ID)
	posts[1].Status = PostStatusAccepted
	posts[1].CreatedByID = users[1].ID
	posts[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&posts))

	threads := []Thread{{PostID: posts[0].ID}, {PostID: posts[1].ID}}

	for i := range threads {
		threads[i].UUID = domain.GetUUID()
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
			ThreadID:  threads[0].ID,        // user 0's post
			SentByID:  posts[0].CreatedByID, // user 0 (Eager)
			Content:   "I can being chocolate if you bring PB",
			CreatedAt: oldOldTime,
		},
		{
			ThreadID:  threads[0].ID,           // user 0's post
			SentByID:  posts[0].ProviderID.Int, // user 1 (Lazy)
			Content:   "Great",
			CreatedAt: oldTime,
		},
		{
			ThreadID:  threads[0].ID,        // user 0's post
			SentByID:  posts[0].CreatedByID, // user 0 (Eager)
			Content:   "Can you get it here by next week?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
		{
			ThreadID:  threads[1].ID,        // user 1's post
			SentByID:  posts[1].CreatedByID, // user 1 (Lazy)
			Content:   "I can being PB if you bring chocolate",
			CreatedAt: oldTime,
		},
		{
			ThreadID:  threads[1].ID,           // user 1's post
			SentByID:  posts[1].ProviderID.Int, // user 0 (Eager)
			Content:   "Did you see my other message?",
			CreatedAt: tNow, // Lazy User doesn't see this one
		},
		{
			ThreadID:  threads[1].ID,           // user 1's post
			SentByID:  posts[1].ProviderID.Int, // user 0 (Eager)
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

	return ThreadFixtures{
		Users:              users,
		Threads:            threads,
		Messages:           messages,
		ThreadParticipants: threadParticipants,
	}
}
