package actions

import (
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"

	"testing"

	"github.com/gobuffalo/nulls"
)

type UserOrgFixtures struct {
	users    models.Users
	orgs     models.Organizations
	userOrgs models.UserOrganizations
}

type meetingFixtures struct {
	models.Users
	models.Meetings
	models.MeetingInvites
	models.File
}

func createFixturesForAuthInvite(as *ActionSuite) meetingFixtures {
	uf := test.CreateUserFixtures(as.DB, 2)
	user := uf.Users[0]
	locations := test.CreateLocationFixtures(as.DB, 2)

	err := aws.CreateS3Bucket()
	as.NoError(err, "failed to create S3 bucket, %s", err)

	var fileFixture models.File
	fErr := fileFixture.Store("new_photo.webp", []byte("RIFFxxxxWEBPVP"))
	as.Nil(fErr, "failed to create ImageFile fixture")

	meetings := make(models.Meetings, 2)
	meetings[1].ImageFileID = nulls.NewInt(fileFixture.ID)

	for i := range meetings {
		meetings[i].CreatedByID = user.ID
		meetings[i].Name = fmt.Sprintf("Meeting%v", i)
		meetings[i].StartDate = time.Now().Add(domain.DurationWeek * time.Duration(i+1))
		meetings[i].EndDate = time.Now().Add(domain.DurationWeek * time.Duration(i+3))
		meetings[i].UUID = domain.GetUUID()
		meetings[i].InviteCode = nulls.NewUUID(domain.GetUUID())
		meetings[i].LocationID = locations[i].ID
		createFixture(as, &meetings[i])
	}

	return meetingFixtures{
		Meetings: meetings,
		File:     fileFixture,
	}
}

func Fixtures_GetOrgAndUserOrgs(as *ActionSuite, t *testing.T) UserOrgFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   models.AuthTypeSaml,
		AuthConfig: "{}",
		UUID:       domain.GetUUID(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	orgDomain := models.OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "example.com",
	}

	if err := as.DB.Create(&orgDomain); err != nil {
		t.Errorf("could not create test org domain ... %v", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			AdminRole: models.UserAdminRoleSuperAdmin,
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
		},
	}

	for i := range users {
		users[i].UUID = domain.GetUUID()
		if err := as.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
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
		if err := as.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	return UserOrgFixtures{
		orgs:     models.Organizations{*org},
		userOrgs: userOrgs,
	}
}

func Fixtures_CreateAuthUser(as *ActionSuite, t *testing.T) UserOrgFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   models.AuthTypeSaml,
		AuthConfig: "{}",
		UUID:       domain.GetUUID(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	orgDomain := models.OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "example.com",
	}

	if err := as.DB.Create(&orgDomain); err != nil {
		t.Errorf("could not create test org domain ... %v", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			AdminRole: models.UserAdminRoleSuperAdmin,
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
		},
	}

	for i := range users {
		users[i].UUID = domain.GetUUID()
		if err := as.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
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
		if err := as.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	return UserOrgFixtures{
		users: users,
		orgs:  models.Organizations{*org},
	}
}

func createFixturesForEnsureMeetingParticipant(as *ActionSuite) meetingFixtures {
	uf := test.CreateUserFixtures(as.DB, 2)
	users := uf.Users
	locations := test.CreateLocationFixtures(as.DB, 2)

	err := aws.CreateS3Bucket()
	as.NoError(err, "failed to create S3 bucket, %s", err)

	var fileFixture models.File
	fErr := fileFixture.Store("new_photo.webp", []byte("RIFFxxxxWEBPVP"))
	as.Nil(fErr, "failed to create ImageFile fixture")

	meetings := make(models.Meetings, 2)
	meetings[1].ImageFileID = nulls.NewInt(fileFixture.ID)

	for i := range meetings {
		meetings[i].CreatedByID = users[0].ID
		meetings[i].Name = fmt.Sprintf("Meeting%v", i)
		meetings[i].StartDate = time.Now().Add(domain.DurationWeek * time.Duration(i+1))
		meetings[i].EndDate = time.Now().Add(domain.DurationWeek * time.Duration(i+3))
		meetings[i].UUID = domain.GetUUID()
		meetings[i].InviteCode = nulls.NewUUID(domain.GetUUID())
		meetings[i].LocationID = locations[i].ID
	}

	createFixture(as, &meetings)

	invites := make(models.MeetingInvites, 2)
	for i := range invites {
		invites[i].MeetingID = meetings[i].ID
		invites[i].Email = users[1].Email
		invites[i].Secret = domain.GetUUID()
		invites[i].InviterID = users[0].ID
	}

	createFixture(as, &invites)

	return meetingFixtures{
		Users:          users,
		Meetings:       meetings,
		File:           fileFixture,
		MeetingInvites: invites,
	}
}
