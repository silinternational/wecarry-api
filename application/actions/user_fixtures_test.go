package actions

import (
	"strconv"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
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
	org := &models.Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(as, org)

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		createFixture(as, &locations[i])
	}

	unique := org.UUID.String()
	users := models.Users{
		{
			AdminRole: models.UserAdminRoleSuperAdmin,
		},
		{
			LocationID: nulls.NewInt(locations[0].ID),
			AdminRole:  models.UserAdminRoleSalesAdmin,
		},
	}
	userOrgs := make(models.UserOrganizations, len(users))
	accessTokenFixtures := make([]models.UserAccessToken, len(users))
	for i := range users {
		users[i].UUID = domain.GetUUID()
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].FirstName = "first" + strconv.Itoa(i)
		users[i].LastName = "last" + strconv.Itoa(i)
		users[i].AuthPhotoURL = nulls.NewString("http://example.com/" + users[i].Nickname + ".gif")
		createFixture(as, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = unique + "_auth_user" + strconv.Itoa(i)
		userOrgs[i].AuthEmail = unique + users[i].Email
		createFixture(as, &userOrgs[i])

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = userOrgs[i].ID
		accessTokenFixtures[i].AccessToken = models.HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		createFixture(as, &accessTokenFixtures[i])
	}

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

	as.NoError(aws.CreateS3Bucket())

	var f models.File
	as.NoError(f.Store("photo.gif", []byte("GIF89a")))

	_, err := users[1].AttachPhoto(f.UUID.String())
	as.NoError(err)

	return UserQueryFixtures{
		Organization:    *org,
		Users:           users,
		UserPreferences: userPreferences,
		Posts:           posts,
		Locations:       locations,
	}
}
