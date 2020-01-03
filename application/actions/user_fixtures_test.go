package actions

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

// UserQueryFixtures is for returning fixtures from Fixtures_UserQuery
type UserQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Locations
	models.UserPreferences
}

func fixturesForUserQuery(as *ActionSuite) UserQueryFixtures {
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

	postDestination := models.Location{}
	createFixture(as, &postDestination)

	posts := models.Posts{
		{
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			ProviderID:     nulls.NewInt(users[1].ID),
			DestinationID:  postDestination.ID,
		},
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		createFixture(as, &posts[i])
	}

	as.NoError(aws.CreateS3Bucket(), "unexpected error creating S3 bucket")

	var f models.File
	as.Nil(f.Store("photo.gif", []byte("GIF89a")), "unexpected error storing file")

	_, err := users[1].AttachPhoto(f.UUID.String())
	as.NoError(err, "unexpected error attaching photo to user")

	return UserQueryFixtures{
		Organization:    org,
		Users:           users,
		UserPreferences: userPreferences,
		Posts:           posts,
		Locations:       uf.Locations,
	}
}
