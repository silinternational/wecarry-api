package models

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type UserMessageFixtures struct {
	Users
	Posts
	Threads
	ThreadParticipants
	Messages
}

func CreateUserFixtures(ms *ModelSuite, t *testing.T) ([]Organization, Users, UserOrganizations) {

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	orgs := []Organization{
		{
			Name:       fmt.Sprintf("ACME-%s", unique),
			Uuid:       domain.GetUuid(),
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
		},
		{
			Name:       fmt.Sprintf("Starfleet Academy-%s", unique),
			Uuid:       domain.GetUuid(),
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
		},
	}
	for i := range orgs {
		if err := ms.DB.Create(&orgs[i]); err != nil {
			t.Errorf("error creating org %+v ...\n %v \n", orgs[i], err)
			t.FailNow()
		}
	}

	// Load User test fixtures
	users := Users{
		{
			Email:     fmt.Sprintf("user1-%s@example.com", unique),
			FirstName: "Existing",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Existing User %s", unique),
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     fmt.Sprintf("user2-%s@example.com", unique),
			FirstName: "Another",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Another User %s", unique),
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     fmt.Sprintf("not_participating-%s@example.com", unique),
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  fmt.Sprintf("Not Participating %s", unique),
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
			OrganizationID: orgs[0].ID,
			UserID:         users[0].ID,
			AuthID:         users[0].Email,
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: orgs[1].ID,
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

	return orgs, users, userOrgs
}

func CreateUserFixtures_CanEditAllPosts(ms *ModelSuite) UserFixtures {
	org := Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, &org)

	unique := org.Uuid.String()
	users := Users{
		{AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin)},
		{AdminRole: nulls.NewString(domain.AdminRoleSalesAdmin)},
		{AdminRole: nulls.String{}},
		{AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin)},
		{AdminRole: nulls.String{}},
		{AdminRole: nulls.NewString(domain.AdminRoleSalesAdmin)},
	}
	for i := range users {
		users[i].Email = "user" + strconv.Itoa(i) + unique + "example.com"
		users[i].Nickname = users[i].Email
		users[i].Uuid = domain.GetUuid()

		createFixture(ms, &users[i])
	}

	userOrgFixtures := []UserOrganization{
		{Role: UserOrganizationRoleAdmin},
		{Role: UserOrganizationRoleAdmin},
		{Role: UserOrganizationRoleAdmin},
		{Role: UserOrganizationRoleUser},
		{Role: UserOrganizationRoleUser},
		{Role: UserOrganizationRoleUser},
	}
	for i := range userOrgFixtures {
		userOrgFixtures[i].OrganizationID = org.ID
		userOrgFixtures[i].UserID = users[i].ID
		userOrgFixtures[i].AuthID = users[i].Email
		userOrgFixtures[i].AuthEmail = users[i].Email

		createFixture(ms, &userOrgFixtures[i])
	}

	return UserFixtures{
		Users: users,
	}
}

func CreateUserFixtures_UnreadMessageCount(ms *ModelSuite, t *testing.T) UserMessageFixtures {

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
			FirstName: "Lazy",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Lazy User %s", unique),
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
			UserID:         users[1].ID,
			AuthID:         users[1].Email,
			AuthEmail:      users[1].Email,
		},
	}
	for i := range userOrgs {
		if err := ms.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrgs[i])
			t.FailNow()
		}
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

	return UserMessageFixtures{
		Users:              users,
		Threads:            threads,
		Messages:           messages,
		ThreadParticipants: threadParticipants,
	}
}

type UserFixtures struct {
	Users
	Threads
}

func CreateUserFixtures_GetThreads(ms *ModelSuite) UserFixtures {
	unique := domain.GetUuid().String()

	org := Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(ms, &org)

	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "_user0", Uuid: domain.GetUuid()},
		{Email: unique + "_user1@example.com", Nickname: unique + "_user1", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{CreatedByID: users[0].ID, OrganizationID: org.ID, Uuid: domain.GetUuid()},
	}
	for i := range posts {
		createFixture(ms, &posts[i])
	}

	threads := []Thread{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(ms, &threads[i])
	}

	threadParticipants := []ThreadParticipant{
		{ThreadID: threads[0].ID, UserID: users[0].ID},
		{ThreadID: threads[1].ID, UserID: users[0].ID},
	}
	for i := range threadParticipants {
		createFixture(ms, &threadParticipants[i])
	}

	return UserFixtures{
		Users:   users,
		Threads: threads,
	}
}
