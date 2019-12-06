package models

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
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

	unique := domain.GetUUID().String()

	// Load Organization test fixtures
	orgs := []Organization{
		{Name: fmt.Sprintf("Starfleet Academy-%s", unique)},
		{Name: fmt.Sprintf("ACME-%s", unique)},
	}

	for i := range orgs {
		orgs[i].AuthType = AuthTypeSaml
		orgs[i].AuthConfig = "{}"
		orgs[i].UUID = domain.GetUUID()
		createFixture(ms, &orgs[i])
	}

	// Load User test fixtures
	users := Users{
		{
			Email:     fmt.Sprintf("user1-%s@example.com", unique),
			FirstName: "Existing",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Existing User %s", unique),
		},
		{
			Email:     fmt.Sprintf("user2-%s@example.com", unique),
			FirstName: "Another",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Another User %s", unique),
		},
		{
			Email:     fmt.Sprintf("not_participating-%s@example.com", unique),
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  fmt.Sprintf("Not Participating %s", unique),
		},
	}

	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	// Load UserOrganization test fixtures
	userOrgs := UserOrganizations{{OrganizationID: orgs[0].ID}, {OrganizationID: orgs[1].ID}}

	for i := range userOrgs {
		userOrgs[i].UserID = users[0].ID
		userOrgs[i].AuthID = users[0].Email
		userOrgs[i].AuthEmail = users[0].Email

		createFixture(ms, &userOrgs[i])
	}
	return orgs, users, userOrgs
}

func CreateUserFixtures_CanEditAllPosts(ms *ModelSuite) UserFixtures {
	org := Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, &org)

	unique := org.UUID.String()
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
		users[i].UUID = domain.GetUUID()

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

func createFixturesForUserFind(ms *ModelSuite) UserFixtures {
	org := &Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, org)

	unique := domain.GetUUID().String()
	users := Users{
		{Email: unique + "user1@example.com", Nickname: unique + "User1"},
		{Email: unique + "user2@example.com", Nickname: unique + "User2"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	return UserFixtures{
		Users: users,
	}
}

func CreateFixturesForUserGetPosts(ms *ModelSuite) UserFixtures {
	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	unique := org.UUID.String()
	users := Users{
		{Email: unique + "user0@example.com", Nickname: unique + "User 0"},
		{Email: unique + "user1@example.com", Nickname: unique + "User 1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
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
		posts[i].UUID = domain.GetUUID()
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return UserFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForTestUserGetPhoto(ms *ModelSuite) UserFixtures {
	ms.NoError(aws.CreateS3Bucket())

	fileFixtures := make([]File, 2)
	for i := range fileFixtures {
		var f File
		err := f.Store(fmt.Sprintf("photo%d.gif", i), []byte("GIF89a"))
		ms.NoError(err)
		fileFixtures[i] = f
	}

	var photoFixture File
	const filename = "photo.gif"
	err := photoFixture.Store(filename, []byte("GIF89a"))
	ms.NoError(err, "failed to create file fixture")

	unique := domain.GetUUID()
	users := Users{
		{},
		{AuthPhotoURL: nulls.NewString("http://www.example.com")},
		{PhotoFileID: nulls.NewInt(fileFixtures[0].ID)},
		{AuthPhotoURL: nulls.NewString("http://www.example.com"), PhotoFileID: nulls.NewInt(fileFixtures[1].ID)},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		users[i].Email = fmt.Sprintf("%s_user%d@example.com", unique, i)
		users[i].Nickname = fmt.Sprintf("%s_User%d", unique, i)
		createFixture(ms, &users[i])

		// ensure the relation is loaded in order to compare filenames
		ms.NoError(DB.Load(&users[i], "PhotoFile"))
	}

	return UserFixtures{
		Users: users,
	}
}

func createFixturesForTestUserSave(ms *ModelSuite) UserFixtures {
	unique := domain.GetUUID()
	users := make(Users, 5)
	for i := range users {
		users[i] = User{
			Email:     fmt.Sprintf("%s_user%d@example.com", unique, i),
			Nickname:  fmt.Sprintf("%s_User%d", unique, i),
			FirstName: fmt.Sprintf("First"),
			LastName:  fmt.Sprintf("Last"),
		}
	}
	users[2].UUID = domain.GetUUID()
	createFixture(ms, &users[3])
	users[3].FirstName = "New"
	users[4].FirstName = ""

	return UserFixtures{
		Users: users,
	}
}

func CreateUserFixturesForNicknames(ms *ModelSuite, t *testing.T) User {
	prefix := allPrefixes()[0]

	// Load User test fixtures
	user := User{
		Email:     fmt.Sprintf("user1-%s@example.com", t.Name()),
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  prefix + "ExistingU",
		UUID:      domain.GetUUID(),
	}

	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("could not create test user %v ... %v", user, err)
		t.FailNow()
	}

	return user
}

func CreateUserFixtures_UnreadMessageCount(ms *ModelSuite, t *testing.T) UserMessageFixtures {

	unique := domain.GetUUID().String()

	// Load Organization test fixtures
	org := Organization{
		Name:       fmt.Sprintf("ACME-%s", unique),
		UUID:       domain.GetUUID(),
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
		},
		{
			Email:     fmt.Sprintf("user2-%s@example.com", unique),
			FirstName: "Lazy",
			LastName:  "User",
			Nickname:  fmt.Sprintf("Lazy User %s", unique),
		},
	}

	for i := range users {
		users[i].UUID = domain.GetUUID()
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
			CreatedByID:   users[0].ID,
			Title:         "Open Request 0",
			ProviderID:    nulls.NewInt(users[1].ID),
			DestinationID: locations[0].ID,
		},
		{
			CreatedByID:   users[1].ID,
			Title:         "Committed Request 1",
			ProviderID:    nulls.NewInt(users[0].ID),
			DestinationID: locations[1].ID,
		},
	}

	for i := range posts {
		posts[i].OrganizationID = org.ID
		posts[i].Type = PostTypeRequest
		posts[i].Size = PostSizeMedium
		posts[i].Status = PostStatusOpen
		posts[i].UUID = domain.GetUUID()
		createFixture(ms, &posts[i])
	}

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
	unique := domain.GetUUID().String()

	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "_user0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "_user1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	location := Location{}
	createFixture(ms, &location)

	posts := Posts{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	createFixture(ms, &posts[0])

	threads := []Thread{{PostID: posts[0].ID}, {PostID: posts[0].ID}}
	for i := range threads {
		threads[i].UUID = domain.GetUUID()
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
	org := Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, &org)

	nicknames := []string{"alice", "bob"}
	unique := org.UUID.String()
	users := make(Users, len(nicknames))
	userLocations := make(Locations, len(users))
	userOrgFixtures := make(UserOrganizations, len(users))
	for i := range users {
		userLocations[i].Country = "US"
		createFixture(ms, &userLocations[i])

		users[i] = User{
			Email:      "user" + strconv.Itoa(i) + unique + "@example.com",
			Nickname:   nicknames[i] + unique,
			UUID:       domain.GetUUID(),
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
		posts[i].UUID = domain.GetUUID()
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
	unique := domain.GetUUID().String()
	users := make(Users, len(nicknames))
	for i := range users {
		users[i] = User{
			Email:    "user" + strconv.Itoa(i) + unique + "@example.com",
			Nickname: nicknames[i] + unique,
			UUID:     domain.GetUUID(),
		}

		createFixture(ms, &users[i])
	}

	// Load UserPreferences test fixtures
	userPreferences := UserPreferences{
		{
			UserID: users[0].ID,
			Key:    domain.UserPreferenceKeyLanguage,
			Value:  domain.UserPreferenceLanguageEnglish,
		},
		{
			UserID: users[0].ID,
			Key:    domain.UserPreferenceKeyWeightUnit,
			Value:  domain.UserPreferenceWeightUnitKGs,
		},
	}

	for i := range userPreferences {
		userPreferences[i].UUID = domain.GetUUID()
		createFixture(ms, &userPreferences[i])
	}

	return UserFixtures{Users: users, UserPreferences: userPreferences}
}
