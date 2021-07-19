package actions

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

// UserFixtures is for returning fixtures from `fixturesForUsers`
type UserFixtures struct {
	models.Organization
	models.Users
	models.Requests
	models.Locations
	models.UserPreferences
	models.Files
	models.Meetings
}

func fixturesForUsers(as *ActionSuite) UserFixtures {
	uf := test.CreateUserFixtures(as.DB, 2)
	users := uf.Users
	org := uf.Organization

	users[0].AdminRole = models.UserAdminRoleSuperAdmin
	users[1].AdminRole = models.UserAdminRoleSalesAdmin
	as.NoError(as.DB.Save(&users))

	// Load UserPreferences test fixtures
	userPreferences := models.UserPreferences{
		{
			UserID: users[1].ID,
			Key:    domain.UserPreferenceKeyLanguage,
			Value:  domain.UserPreferenceLanguageFrench,
		},
		{
			UserID: users[1].ID,
			Key:    domain.UserPreferenceKeyTimeZone,
			Value:  "America/New_York",
		},
		{
			UserID: users[1].ID,
			Key:    domain.UserPreferenceKeyWeightUnit,
			Value:  domain.UserPreferenceWeightUnitPounds,
		},
	}

	for i := range userPreferences {
		userPreferences[i].UUID = domain.GetUUID()
		createFixture(as, &userPreferences[i])
	}

	loc := test.CreateLocationFixtures(as.DB, 2)
	requestDestination := loc[0]
	meetingLocation := loc[1]

	requests := models.Requests{
		{
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			ProviderID:     nulls.NewInt(users[1].ID),
			DestinationID:  requestDestination.ID,
		},
	}
	for i := range requests {
		requests[i].UUID = domain.GetUUID()
		createFixture(as, &requests[i])
	}

	as.NoError(aws.CreateS3Bucket(), "unexpected error creating S3 bucket")

	f := test.CreateFileFixture(as.DB)

	_, err := users[1].AttachPhoto(as.DB, f.UUID.String())
	as.NoError(err, "unexpected error attaching photo to user")

	meetings := models.Meetings{
		{
			CreatedByID: users[0].ID,
			Name:        "Meeting Name",
			LocationID:  meetingLocation.ID,
			StartDate:   time.Now().Add(domain.DurationWeek * 4),
			EndDate:     time.Now().Add(domain.DurationWeek * 5),
		},
	}
	test.MustCreate(as.DB, &meetings[0])

	mp := models.MeetingParticipant{MeetingID: meetings[0].ID, UserID: users[1].ID}
	test.MustCreate(as.DB, &mp)

	return UserFixtures{
		Organization:    org,
		Users:           users,
		UserPreferences: userPreferences,
		Requests:        requests,
		Locations:       uf.Locations,
		Files:           models.Files{f},
		Meetings:        meetings,
	}
}
