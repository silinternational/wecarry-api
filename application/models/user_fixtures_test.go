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
			Name:       fmt.Sprintf("Starfleet Academy-%s", unique),
			Uuid:       domain.GetUuid(),
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
		},
		{
			Name:       fmt.Sprintf("ACME-%s", unique),
			Uuid:       domain.GetUuid(),
			AuthType:   AuthTypeSaml,
			AuthConfig: "{}",
		},
	}

	for i := range orgs {
		createFixture(ms, &orgs[i])
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
		createFixture(ms, &users[i])
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
		createFixture(ms, &userOrgs[i])
	}
	return orgs, users, userOrgs
}

func CreateUserFixtures_CanEditAllPosts(ms *ModelSuite) UserFixtures {
	org := Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, &org)

	unique := org.Uuid.String()
	users := Users{
		{AdminRole: UserAdminRoleSuperAdmin},
		{AdminRole: UserAdminRoleSalesAdmin},
		{AdminRole: UserAdminRoleUser},
		{AdminRole: UserAdminRoleSuperAdmin},
		{AdminRole: UserAdminRoleUser},
		{AdminRole: UserAdminRoleSalesAdmin},
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

func CreateFixturesForUserGetPosts(ms *ModelSuite) UserFixtures {
	org := Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(ms, &org)

	unique := org.Uuid.String()
	users := Users{
		{Email: unique + "user0@example.com", Nickname: unique + "User 0", Uuid: domain.GetUuid()},
		{Email: unique + "user1@example.com", Nickname: unique + "User 1", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	const numberOfPosts = 4
	locations := make([]Location, numberOfPosts)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := []Post{
		{ProviderID: nulls.NewInt(users[1].ID)},
		{ProviderID: nulls.NewInt(users[1].ID)},
		{ReceiverID: nulls.NewInt(users[1].ID)},
		{ReceiverID: nulls.NewInt(users[1].ID)},
	}
	for i := range posts {
		posts[i].CreatedByID = users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].Uuid = domain.GetUuid()
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return UserFixtures{
		Users: users,
		Posts: posts,
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

	locations := []Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
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
			DestinationID:  locations[0].ID,
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
			DestinationID:  locations[1].ID,
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
	UserPreferences
	Posts
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

	location := Location{}
	createFixture(ms, &location)

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
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

func CreateFixturesForUserWantsPostNotification(ms *ModelSuite) UserFixtures {
	org := Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, &org)

	nicknames := []string{"alice", "bob"}
	unique := org.Uuid.String()
	users := make(Users, len(nicknames))
	userLocations := make(Locations, len(users))
	userOrgFixtures := make(UserOrganizations, len(users))
	for i := range users {
		userLocations[i].Country = "US"
		createFixture(ms, &userLocations[i])

		users[i] = User{
			Email:      "user" + strconv.Itoa(i) + unique + "@example.com",
			Nickname:   nicknames[i] + unique,
			Uuid:       domain.GetUuid(),
			LocationID: nulls.NewInt(userLocations[i].ID),
		}
		createFixture(ms, &users[i])

		userOrgFixtures[i] = UserOrganization{
			OrganizationID: org.ID,
			UserID:         users[i].ID,
			AuthID:         users[i].Email,
			AuthEmail:      users[i].Email,
		}
		createFixture(ms, &userOrgFixtures[i])
	}

	postLocations := Locations{
		{ // Post 0 Destination
			Description: "close",
			Country:     "US",
		},
		{ // Post 1 Destination
			Description: "far away",
			Country:     "KR",
		},
		{ // Post 2 Destination
			Description: "far away",
			Country:     "KR",
		},
		{ // Post 3 Destination
			Description: "close",
			Country:     "US",
		},
		{ // Post 4 Destination
			Description: "close",
			Country:     "US",
		},
		{ // Post 0 Origin
			Description: "far away",
			Country:     "KR",
		},
		{ // Post 2 Origin
			Description: "close",
			Country:     "US",
		},
	}
	for i := range postLocations {
		createFixture(ms, &postLocations[i])
	}

	posts := Posts{
		{
			Type:     PostTypeRequest,
			OriginID: nulls.NewInt(postLocations[5].ID),
		},
		{
			Type:     PostTypeOffer,
			OriginID: nulls.Int{},
		},
		{
			Type:     PostTypeRequest,
			OriginID: nulls.NewInt(postLocations[6].ID),
		},
		{
			Type:     PostTypeOffer,
			OriginID: nulls.Int{},
		},
		{
			Type:     PostTypeRequest,
			OriginID: nulls.Int{},
		},
	}
	for i := range posts {
		posts[i].Uuid = domain.GetUuid()
		posts[i].CreatedByID = users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = postLocations[i].ID
		createFixture(ms, &posts[i])
	}

	return UserFixtures{
		Users: users,
		Posts: posts,
	}
}

func CreateUserFixtures_TestGetPreference(ms *ModelSuite) UserFixtures {
	nicknames := []string{"alice", "bob"}
	unique := domain.GetUuid().String()
	users := make(Users, len(nicknames))
	for i := range users {
		users[i] = User{
			Email:    "user" + strconv.Itoa(i) + unique + "@example.com",
			Nickname: nicknames[i] + unique,
			Uuid:     domain.GetUuid(),
		}

		createFixture(ms, &users[i])
	}

	// Load UserPreferences test fixtures
	userPreferences := UserPreferences{
		{
			Uuid:   domain.GetUuid(),
			UserID: users[0].ID,
			Key:    "User0Key1",
			Value:  "User0Val1",
		},
		{
			Uuid:   domain.GetUuid(),
			UserID: users[0].ID,
			Key:    "User0Key2",
			Value:  "User0Val2",
		},
	}

	for i := range userPreferences {
		createFixture(ms, &userPreferences[i])
	}

	return UserFixtures{Users: users, UserPreferences: userPreferences}
}
