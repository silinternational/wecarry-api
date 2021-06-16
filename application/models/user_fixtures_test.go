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
	Requests
	Threads
	ThreadParticipants
	Messages
}

// createFixturesForUserGetOrganizations is used for TestUser_GetOrgIDs and TestUser_GetOrganizations
func createFixturesForUserGetOrganizations(ms *ModelSuite) ([]Organization, Users) {
	unique := domain.GetUUID()
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

	users := createUserFixtures(ms.DB, 1).Users

	// user is already in org 0, but need user to also be in org 1
	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	return orgs, users
}

func CreateUserFixtures_CanEditAllRequests(ms *ModelSuite) UserRequestFixtures {
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

	return UserRequestFixtures{
		Users: users,
	}
}

func CreateFixturesForUserGetRequests(ms *ModelSuite) UserRequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	requests := createRequestFixtures(ms.DB, 4, false)
	userID := users[1].UUID.String()
	requests[0].SetProviderWithStatus(Ctx(), RequestStatusAccepted, &userID)
	requests[1].SetProviderWithStatus(Ctx(), RequestStatusAccepted, &userID)
	requests[2].Status = RequestStatusAccepted
	requests[3].Status = RequestStatusAccepted
	ms.NoError(ms.DB.Save(&requests))

	return UserRequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

//        Org0          Org1            Org2
//        | |           | | |          | |    \
//        | +-------+---+ | +----+-----+ +    |
//        |         |            |       |    |
//       User0    User1         Trust   User2  User3 (SuperAdmin)
//
// Org0: Request0 (SAME)    <cannot be seen by User2>
// Org1: Request1 (SAME)    <cannot be seen by User0 and User2>
// Org2: Request2 (ALL)     <seen by all>,
//		 Request3 (TRUSTED) <cannot be seen by User0>,
//		 Request4 (SAME)    <cannot be seen by User0 and User1>
//
func CreateFixturesForUserCanViewRequest(ms *ModelSuite) UserRequestFixtures {
	orgs := createOrganizationFixtures(ms.DB, 3)

	trusts := OrganizationTrusts{
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[2].ID},
		{PrimaryID: orgs[2].ID, SecondaryID: orgs[1].ID},
	}
	createFixture(ms, &trusts)

	users := createUserFixtures(ms.DB, 4).Users
	users[3].AdminRole = UserAdminRoleSuperAdmin
	ms.NoError(ms.DB.Save(&users[3]))

	// Give User1 a second Org
	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[1].ID,
		AuthID:         users[1].Email,
		AuthEmail:      users[1].Email,
	})

	// Switch User2's org to Org2
	uo, err := users[2].FindUserOrganization(ms.DB, orgs[0])
	ms.NoError(err)
	uo.OrganizationID = orgs[2].ID
	ms.NoError(DB.UpdateColumns(&uo, "organization_id"))

	// Switch User3's org to Org2
	uo, err = users[3].FindUserOrganization(ms.DB, orgs[0])
	ms.NoError(err)
	uo.OrganizationID = orgs[2].ID
	ms.NoError(DB.UpdateColumns(&uo, "organization_id"))

	requests := createRequestFixtures(ms.DB, 5, false)
	requests[1].OrganizationID = orgs[1].ID
	requests[1].CreatedByID = users[1].ID

	requests[2].OrganizationID = orgs[2].ID
	requests[2].CreatedByID = users[2].ID
	requests[2].Visibility = RequestVisibilityAll

	requests[3].OrganizationID = orgs[2].ID
	requests[3].CreatedByID = users[2].ID
	requests[3].Visibility = RequestVisibilityTrusted

	requests[4].OrganizationID = orgs[2].ID
	requests[4].CreatedByID = users[2].ID
	ms.NoError(ms.DB.Save(&requests))

	return UserRequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func createFixturesForTestUserGetPhoto(ms *ModelSuite) UserRequestFixtures {
	ms.NoError(aws.CreateS3Bucket())

	fileFixtures := createFileFixtures(ms.DB, 2)

	unique := domain.GetUUID()
	users := Users{
		{},
		{AuthPhotoURL: nulls.NewString("http://www.example.com")},
		{FileID: nulls.NewInt(fileFixtures[0].ID)},
		{AuthPhotoURL: nulls.NewString("http://www.example.com"), FileID: nulls.NewInt(fileFixtures[1].ID)},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		users[i].Email = fmt.Sprintf("%s_user%d@example.com", unique, i)
		users[i].Nickname = fmt.Sprintf("%s_User%d", unique, i)
		createFixture(ms, &users[i])

		// ensure the relation is loaded in order to compare filenames
		ms.NoError(DB.Load(&users[i], "PhotoFile"))
	}

	return UserRequestFixtures{
		Users: users,
	}
}

func createFixturesForTestUserSave(ms *ModelSuite) UserRequestFixtures {
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

	return UserRequestFixtures{
		Users: users,
	}
}

func CreateUserFixturesForNicknames(ms *ModelSuite, t *testing.T, prefix string) User {
	// Load User test fixtures
	user := User{
		Email:     fmt.Sprintf("user1-%s@example.com", t.Name()),
		FirstName: "Existing",
		LastName:  "User",
		Nickname:  prefix + " Existing U",
		UUID:      domain.GetUUID(),
	}

	if err := ms.DB.Create(&user); err != nil {
		t.Errorf("could not create test user %v ... %v", user, err)
		t.FailNow()
	}

	return user
}

func CreateUserFixtures_UnreadMessageCount(ms *ModelSuite, t *testing.T) UserMessageFixtures {
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

	threads := []Thread{{RequestID: requests[0].ID}, {RequestID: requests[1].ID}}

	for i := range threads {
		threads[i].UUID = domain.GetUUID()
		createFixture(ms, &threads[i])
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

	return UserMessageFixtures{
		Users:              users,
		Threads:            threads,
		Messages:           messages,
		ThreadParticipants: threadParticipants,
	}
}

type UserRequestFixtures struct {
	Users
	UserPreferences
	Requests
	Threads
}

func CreateUserFixtures_GetThreads(ms *ModelSuite) UserRequestFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	location := Location{}
	createFixture(ms, &location)

	requests := Requests{
		{UUID: domain.GetUUID(), CreatedByID: users[0].ID, OrganizationID: org.ID, DestinationID: location.ID},
	}
	createFixture(ms, &requests[0])

	threads := []Thread{{RequestID: requests[0].ID}, {RequestID: requests[0].ID}}
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

	return UserRequestFixtures{
		Users:   users,
		Threads: threads,
	}
}

func CreateFixturesForUserWantsRequestNotification(ms *ModelSuite) UserRequestFixtures {
	uf := createUserFixtures(ms.DB, 3)
	users := uf.Users

	for i := range users {
		ms.NoError(ms.DB.Load(&users[i], "Location"))
		users[i].Location.Country = "US"
		ms.NoError(ms.DB.Save(&users[i].Location))
	}

	requests := createRequestFixtures(ms.DB, 3, false)
	requestOneLocation, err := requests[1].GetOrigin(ms.DB)
	ms.NoError(err)
	ms.NoError(users[1].SetLocation(Ctx(), *requestOneLocation))

	// make a copy of the request destination and assign it to a watch
	watchLocation, err := requests[2].GetDestination(ms.DB)
	ms.NoError(err)
	createFixture(ms, watchLocation)
	watch := Watch{
		UUID:          domain.GetUUID(),
		OwnerID:       users[2].ID,
		DestinationID: nulls.NewInt(watchLocation.ID),
	}
	createFixture(ms, &watch)

	return UserRequestFixtures{
		Users:    users,
		Requests: requests,
	}
}

func CreateUserFixtures_TestGetPreference(ms *ModelSuite) UserRequestFixtures {
	users := createUserFixtures(ms.DB, 2).Users

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

	return UserRequestFixtures{Users: users, UserPreferences: userPreferences}
}

func CreateUserFixtures_TestGetLanguagePreference(ms *ModelSuite) Users {
	users := createUserFixtures(ms.DB, 3).Users

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
		{
			UserID: users[2].ID,
			Key:    domain.UserPreferenceKeyLanguage,
			Value:  domain.UserPreferenceLanguageFrench,
		},
	}

	for i := range userPreferences {
		userPreferences[i].UUID = domain.GetUUID()
		createFixture(ms, &userPreferences[i])
	}

	return users
}
