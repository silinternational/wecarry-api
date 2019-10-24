package models

import (
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
	"testing"
	"time"
)

type ThreadFixtures struct {
	Users
	Posts
	Threads
	ThreadParticipants
	Messages
}

func CreateThreadFixtures(ms *ModelSuite, t *testing.T, post Post) ThreadFixtures {
	// Load Thread test fixtures
	threads := []Thread{
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
	}
	for i := range threads {
		if err := ms.DB.Create(&threads[i]); err != nil {
			t.Errorf("could not create test threads ... %v", err)
			t.FailNow()
		}
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
			UserID:   post.ProviderID.Int,
		},
		{
			ThreadID: threads[1].ID,
			UserID:   post.CreatedByID,
		},
	}
	for i := range threadParticipants {
		if err := ms.DB.Create(&threadParticipants[i]); err != nil {
			t.Errorf("could not create test thread participants ... %v", err)
			t.FailNow()
		}
	}

	// Load Message test fixtures
	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: post.CreatedByID,
			Content:  "I can being chocolate if you bring PB",
		},
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[1].ID,
			SentByID: post.ProviderID.Int,
			Content:  "I can being PB if you bring chocolate",
		},
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[1].ID,
			SentByID: post.CreatedByID,
			Content:  "Great!",
		},
	}

	for _, message := range messages {
		if err := ms.DB.Create(&message); err != nil {
			t.Errorf("could not create test message ... %v", err)
			t.FailNow()
		}
	}

	return ThreadFixtures{Threads: threads, Messages: messages, ThreadParticipants: threadParticipants}
}

func CreateThreadFixtures_UnreadMessageCount(ms *ModelSuite, t *testing.T, post Post) ThreadFixtures {

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	org := Organization{
		Name:       fmt.Sprintf("ACME-%s", unique),
		Uuid:       domain.GetUuid(),
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
	}

	if err := ms.DB.Create(&org); err != nil {
		t.Errorf("error creating org %+v ...\n %v \n", org, err)
		t.FailNow()
	}

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
			FirstName: "Sleepy",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Sleepy User %s", unique),
			Uuid:      domain.GetUuid(),
		},
	}
	for i := range users {
		if err := ms.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user %v ... %v", users[i], err)
			t.FailNow()
		}
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
			UserID:         users[0].ID,
			AuthID:         users[0].Email,
			AuthEmail:      users[0].Email,
		},
	}
	for i := range userOrgs {
		if err := ms.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrgs[i])
			t.FailNow()
		}
	}

	posts := []Post{
		{
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Open Request 0",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			Type:           PostTypeRequest,
			Title:          "Committed Request 1",
			Size:           PostSizeMedium,
			Status:         PostStatusOpen,
			Uuid:           domain.GetUuid(),
		},
	}

	if err := CreatePosts(posts); err != nil {
		t.Errorf("could not create test post ... %v", err)
		t.FailNow()
	}

	return ThreadFixtures{}
}
