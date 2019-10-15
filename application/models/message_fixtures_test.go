package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type MessageFixtures struct {
	Users    Users
	Messages Messages
	Threads  Threads
}

func Fixtures_GetSender(ms *ModelSuite, t *testing.T) MessageFixtures {
	// Load Org test fixtures
	org := &Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := ms.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for i := range users {
		if err := ms.DB.Create(&users[i]); err != nil {
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

	for i := range userOrgs {
		if err := ms.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	for _, loc := range fixtureLocations {
		err := models.DB.Create(loc)
		if err != nil {
			err = fmt.Errorf("error loading post fixture ... %+v\n %v", loc, err.Error())
			return err
		}
	}

	// Load Post test fixtures
	posts := Posts{
		{
			Uuid:           domain.GetUuid(),
			Type:           PostTypeRequest,
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Status:         PostStatusOpen,
			Title:          "I need PB",
			Size:           PostSizeMedium,
			ReceiverID:     nulls.NewInt(users[0].ID),
			NeededAfter:    time.Date(2019, time.July, 19, 0, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, time.August, 3, 0, 0, 0, 0, time.UTC),
			Category:       "Unknown",
			Description:    nulls.NewString("Missing my PB & J"),
		},
		{
			Uuid:           domain.GetUuid(),
			Type:           PostTypeRequest,
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			Status:         PostStatusOpen,
			Title:          "Please bring chocolate",
			Size:           PostSizeSmall,
			ReceiverID:     nulls.NewInt(users[1].ID),
			NeededAfter:    time.Date(2019, time.July, 19, 1, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, time.August, 3, 1, 0, 0, 0, time.UTC),
			Category:       "Unknown",
			Description:    nulls.NewString("2-3 bars"),
		},
	}

	for i := range posts {
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
	}

	// Load Thread test fixtures
	threads := Threads{
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
		if err := ms.DB.Create(&threads[i]); err != nil {
			t.Errorf("could not create test thread ... %v", err)
			t.FailNow()
		}
	}

	// Load Message test fixtures
	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[1].ID,
			SentByID: users[1].ID,
			Content:  "I can being PB if you bring chocolate",
		},
	}

	for i := range messages {
		if err := ms.DB.Create(&messages[i]); err != nil {
			t.Errorf("could not create test message ... %v", err)
			t.FailNow()
		}
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}

}
