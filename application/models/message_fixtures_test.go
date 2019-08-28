package models

import (
	buffalo_models "github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"testing"
	"time"
)

type MessageFixtures struct {
	Users    Users
	Messages Messages
}

func Fixtures_GetSender(t *testing.T) MessageFixtures {
	// Load Org test fixtures
	org := &Organization{
		ID:         1,
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := buffalo_models.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := Users{
		{
			ID:        1,
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
		},
		{
			ID:        2,
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for _, user := range users {
		if err := DB.Create(&user); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         "auth_user1",
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         "auth_user2",
			AuthEmail:      users[1].Email,
		},
	}

	for _, uOrg := range userOrgs {
		if err := DB.Create(&uOrg); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	// Load Post test fixtures

	posts := Posts{
		{
			ID:             1,
			Uuid:           domain.GetUuid(),
			Type:           PostTypeRequest,
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Status:         PostStatusUnfulfilled,
			Title:          "I need PB",
			Destination:    nulls.NewString("Madrid, Spain"),
			Size:           PostSizeMedium,
			ReceiverID:     nulls.NewInt(users[0].ID),
			NeededAfter:    time.Date(2019, time.July, 19, 0, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, time.August, 3, 0, 0, 0, 0, time.UTC),
			Category:       "Unknown",
			Description:    nulls.NewString("Missing my PB & J"),
		},
		{
			ID:             2,
			Uuid:           domain.GetUuid(),
			Type:           PostTypeRequest,
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			Status:         PostStatusUnfulfilled,
			Title:          "Please bring chocolate",
			Destination:    nulls.NewString("Nairobi, Kenya"),
			Size:           PostSizeSmall,
			ReceiverID:     nulls.NewInt(users[1].ID),
			NeededAfter:    time.Date(2019, time.July, 19, 1, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, time.August, 3, 1, 0, 0, 0, time.UTC),
			Category:       "Unknown",
			Description:    nulls.NewString("2-3 bars"),
		},
	}

	for _, post := range posts {
		if err := DB.Create(&post); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
	}

	// Load Thread test fixtures
	threads := Threads{
		{
			ID:     1,
			Uuid:   domain.GetUuid(),
			PostID: posts[0].ID,
		},
		{
			ID:     2,
			Uuid:   domain.GetUuid(),
			PostID: posts[1].ID,
		},
	}

	for _, thread := range threads {
		if err := DB.Create(&thread); err != nil {
			t.Errorf("could not create test thread ... %v", err)
			t.FailNow()
		}
	}

	// Load Message test fixtures
	messages := Messages{
		{
			ID:       1,
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
		{
			ID:       2,
			Uuid:     domain.GetUuid(),
			ThreadID: threads[1].ID,
			SentByID: users[1].ID,
			Content:  "I can being PB if you bring chocolate",
		},
	}

	for _, message := range messages {
		if err := DB.Create(&message); err != nil {
			t.Errorf("could not create test message ... %v", err)
			t.FailNow()
		}
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
	}

}
