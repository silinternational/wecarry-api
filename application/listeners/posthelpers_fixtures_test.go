package listeners

import (
	"testing"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type orgUserPostFixtures struct {
	orgs  models.Organizations
	users models.Users
	posts models.Requests
}

func CreateFixtures_GetPostUsers(ms *ModelSuite, t *testing.T) orgUserPostFixtures {
	userFixtures := test.CreateUserFixtures(ms.DB, 3)
	org := userFixtures.Organization
	users := userFixtures.Users

	_, err := users[1].UpdateStandardPreferences(models.StandardPreferences{Language: domain.UserPreferenceLanguageFrench})

	ms.NoError(err, "could not create language preference for user "+users[1].Nickname)

	posts := test.CreateRequestFixtures(ms.DB, 2, false)
	posts[0].Status = models.RequestStatusAccepted
	posts[0].ProviderID = nulls.NewInt(users[1].ID)
	ms.NoError(ms.DB.Save(&posts[0]))

	return orgUserPostFixtures{
		orgs:  models.Organizations{org},
		users: users,
		posts: posts,
	}
}

func CreateFixtures_RequestStatusUpdatedNotifications(ms *ModelSuite, t *testing.T) orgUserPostFixtures {
	userFixtures := test.CreateUserFixtures(ms.DB, 3)
	org := userFixtures.Organization
	users := userFixtures.Users

	tU := models.User{}
	ms.NoError(tU.FindByID(users[1].ID))

	posts := test.CreateRequestFixtures(ms.DB, 2, false)
	posts[0].Status = models.RequestStatusAccepted
	posts[0].ProviderID = nulls.NewInt(users[1].ID)
	ms.NoError(posts[0].Update())

	return orgUserPostFixtures{
		orgs:  models.Organizations{org},
		users: users,
		posts: posts,
	}
}

func CreateFixtures_sendNotificationRequestFromStatus(ms *ModelSuite, t *testing.T) orgUserPostFixtures {

	unique := domain.GetUUID().String()

	// Load Organization test fixtures
	org := models.Organization{
		Name:       "ACME-" + unique,
		UUID:       domain.GetUUID(),
		AuthType:   models.AuthTypeSaml,
		AuthConfig: "{}",
	}

	if err := ms.DB.Create(&org); err != nil {
		t.Errorf("error creating org %+v ...\n %v \n", org, err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1-" + unique + "@example.com",
			FirstName: "Existing1",
			LastName:  "User",
			Nickname:  "Existing1 User " + unique,
		},
		{
			Email:     "user2-" + unique + "@example.com",
			FirstName: "Existing2",
			LastName:  "User",
			Nickname:  "Existing2 User " + unique,
		},
		{
			Email:     "not_participating-" + unique + "@example.com",
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  "Not Participating " + unique,
		},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		if err := ms.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user %v ... %v", users[i], err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrg := models.UserOrganization{

		OrganizationID: org.ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	}
	if err := ms.DB.Create(&userOrg); err != nil {
		t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrg)
		t.FailNow()
	}

	if err := models.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []models.Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	// Load Request test fixtures
	posts := []models.Request{
		{
			CreatedByID:    users[0].ID,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "First Request",
			Size:           models.RequestSizeMedium,
			Status:         models.RequestStatusOpen,
			ProviderID:     nulls.NewInt(users[1].ID),
			DestinationID:  locations[0].ID,
		},
		{
			CreatedByID:    users[0].ID,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "Second Request",
			Size:           models.RequestSizeMedium,
			Status:         models.RequestStatusOpen,
			DestinationID:  locations[1].ID,
		},
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := models.DB.Load(&posts[i], "CreatedBy", "Provider", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}

	return orgUserPostFixtures{
		orgs:  models.Organizations{org},
		users: users,
		posts: posts,
	}
}

func createFixturesForTestSendNewPostNotifications(ms *ModelSuite) orgUserPostFixtures {
	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	posts := test.CreateRequestFixtures(ms.DB, 1, false)
	postOrigin, err := posts[0].GetOrigin()
	ms.NoError(err)

	users := test.CreateUserFixtures(ms.DB, 3).Users
	for i := range users {
		ms.NoError(users[i].SetLocation(*postOrigin))
	}

	return orgUserPostFixtures{
		users: users,
		posts: posts,
	}
}

func createFixture(ms *ModelSuite, f interface{}) {
	err := ms.DB.Create(f)
	if err != nil {
		ms.T().Errorf("error creating %T fixture, %s", f, err)
		ms.T().FailNow()
	}
}
