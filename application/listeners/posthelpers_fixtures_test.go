package listeners

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	m "github.com/silinternational/wecarry-api/models"
)

type orgUserPostFixtures struct {
	orgs  m.Organizations
	users m.Users
	posts m.Posts
}

func CreateFixtures_GetPostRecipients(ms *ModelSuite, t *testing.T) orgUserPostFixtures {

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	org := m.Organization{
		Name:       "ACME-" + unique,
		Uuid:       domain.GetUuid(),
		AuthType:   m.AuthTypeSaml,
		AuthConfig: "{}",
	}

	if err := ms.DB.Create(&org); err != nil {
		t.Errorf("error creating org %+v ...\n %v \n", org, err)
		t.FailNow()
	}

	// Load User test fixtures
	users := m.Users{
		{
			Email:     "user1-" + unique + "@example.com",
			FirstName: "Existing1",
			LastName:  "User",
			Nickname:  "Existing1 User " + unique,
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "user2-" + unique + "@example.com",
			FirstName: "Existing2",
			LastName:  "User",
			Nickname:  "Existing2 User " + unique,
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "not_participating-" + unique + "@example.com",
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  "Not Participating " + unique,
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
	userOrg := m.UserOrganization{
		OrganizationID: org.ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	}
	if err := ms.DB.Create(&userOrg); err != nil {
		t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrg)
		t.FailNow()
	}

	if err := m.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []m.Location{{}, {}}
	for i := range locations {
		createFixture(ms, &(locations[i]))
	}

	// Load Post test fixtures
	posts := []m.Post{
		{
			CreatedByID:    users[0].ID,
			Type:           m.PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "First Request",
			Size:           m.PostSizeMedium,
			Status:         m.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
			ReceiverID:     nulls.NewInt(users[0].ID),
			DestinationID:  locations[0].ID,
		},
		{
			CreatedByID:    users[0].ID,
			Type:           m.PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "Second Request",
			Size:           m.PostSizeMedium,
			Status:         m.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[0].ID),
			DestinationID:  locations[1].ID,
		},
	}
	for i := range posts {
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := m.DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}

	return orgUserPostFixtures{
		orgs:  m.Organizations{org},
		users: users,
		posts: posts,
	}
}

func CreateFixtures_RequestStatusUpdatedNotifications(ms *ModelSuite, t *testing.T) orgUserPostFixtures {

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	org := m.Organization{
		Name:       "ACME-" + unique,
		Uuid:       domain.GetUuid(),
		AuthType:   m.AuthTypeSaml,
		AuthConfig: "{}",
	}

	if err := ms.DB.Create(&org); err != nil {
		t.Errorf("error creating org %+v ...\n %v \n", org, err)
		t.FailNow()
	}

	// Load User test fixtures
	users := m.Users{
		{
			Email:     "user1-" + unique + "@example.com",
			FirstName: "Existing1",
			LastName:  "User",
			Nickname:  "Existing1 User " + unique,
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "user2-" + unique + "@example.com",
			FirstName: "Existing2",
			LastName:  "User",
			Nickname:  "Existing2 User " + unique,
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "not_participating-" + unique + "@example.com",
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  "Not Participating " + unique,
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
	userOrg := m.UserOrganization{

		OrganizationID: org.ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	}
	if err := ms.DB.Create(&userOrg); err != nil {
		t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrg)
		t.FailNow()
	}

	if err := m.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []m.Location{{}, {}}
	for i := range locations {
		createFixture(ms, &(locations[i]))
	}

	// Load Post test fixtures
	posts := []m.Post{
		{
			CreatedByID:    users[0].ID,
			Type:           m.PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "First Request",
			Size:           m.PostSizeMedium,
			Status:         m.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
			ReceiverID:     nulls.NewInt(users[0].ID),
			DestinationID:  locations[0].ID,
		},
		{
			CreatedByID:    users[0].ID,
			Type:           m.PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "Second Request",
			Size:           m.PostSizeMedium,
			Status:         m.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[0].ID),
			DestinationID:  locations[1].ID,
		},
	}
	for i := range posts {
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := m.DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}

	return orgUserPostFixtures{
		orgs:  m.Organizations{org},
		users: users,
		posts: posts,
	}
}

func CreateFixtures_sendNotificationRequestFromStatus(ms *ModelSuite, t *testing.T) orgUserPostFixtures {

	unique := domain.GetUuid().String()

	// Load Organization test fixtures
	org := m.Organization{
		Name:       "ACME-" + unique,
		Uuid:       domain.GetUuid(),
		AuthType:   m.AuthTypeSaml,
		AuthConfig: "{}",
	}

	if err := ms.DB.Create(&org); err != nil {
		t.Errorf("error creating org %+v ...\n %v \n", org, err)
		t.FailNow()
	}

	// Load User test fixtures
	users := m.Users{
		{
			Email:     "user1-" + unique + "@example.com",
			FirstName: "Existing1",
			LastName:  "User",
			Nickname:  "Existing1 User " + unique,
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "user2-" + unique + "@example.com",
			FirstName: "Existing2",
			LastName:  "User",
			Nickname:  "Existing2 User " + unique,
			Uuid:      domain.GetUuid(),
		},
		{
			Email:     "not_participating-" + unique + "@example.com",
			FirstName: "Not",
			LastName:  "Participating",
			Nickname:  "Not Participating " + unique,
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
	userOrg := m.UserOrganization{

		OrganizationID: org.ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	}
	if err := ms.DB.Create(&userOrg); err != nil {
		t.Errorf("could not create test user org ... %v. uo = %+v", err, userOrg)
		t.FailNow()
	}

	if err := m.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []m.Location{{}, {}}
	for i := range locations {
		createFixture(ms, &(locations[i]))
	}

	// Load Post test fixtures
	posts := []m.Post{
		{
			CreatedByID:    users[0].ID,
			Type:           m.PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "First Request",
			Size:           m.PostSizeMedium,
			Status:         m.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
			ReceiverID:     nulls.NewInt(users[0].ID),
			DestinationID:  locations[0].ID,
		},
		{
			CreatedByID:    users[0].ID,
			Type:           m.PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "Second Request",
			Size:           m.PostSizeMedium,
			Status:         m.PostStatusOpen,
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[0].ID),
			DestinationID:  locations[1].ID,
		},
	}
	for i := range posts {
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := m.DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}

	return orgUserPostFixtures{
		orgs:  m.Organizations{org},
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
