package test

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/pop"
)

type UserFixtures struct {
	models.Organization
	models.Users
	models.Locations
}

// CreateFixture saves a record to the database. The given test is failed if any errors occur.
func CreateFixture(tx *pop.Connection, t *testing.T, f interface{}) {
	err := tx.Create(f)
	if err != nil {
		t.Errorf("error creating %T fixture, %s", f, err)
		t.FailNow()
	}
}

// CreateUserFixtures generates any number of user records for testing. Locations, UserOrganizations, and
// UserAccessTokens are also created for each user. The access token for each user is the same as the user's nickname.
// At least one Organization must exist. All user fixtures will be assigned to the first Organization in the DB.
func CreateUserFixtures(tx *pop.Connection, t *testing.T, n int) UserFixtures {
	var org models.Organization
	if err := tx.First(&org); err != nil {
		org = models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
		CreateFixture(tx, t, &org)
	}

	unique := org.UUID.String()

	users := make(models.Users, n)
	locations := make(models.Locations, n)
	userOrgs := make(models.UserOrganizations, n)
	accessTokenFixtures := make(models.UserAccessTokens, n)
	for i := range users {
		locations[i].Country = "US"
		locations[i].Description = "Miami, FL, US"
		locations[i].Latitude = nulls.NewFloat64(25.7617)
		locations[i].Longitude = nulls.NewFloat64(-80.1918)
		CreateFixture(tx, t, &locations[i])

		users[i].UUID = domain.GetUUID()
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].FirstName = "first" + strconv.Itoa(i)
		users[i].LastName = "last" + strconv.Itoa(i)
		users[i].AuthPhotoURL = nulls.NewString("http://example.com/" + users[i].Nickname + ".gif")
		users[i].LocationID = nulls.NewInt(locations[i].ID)
		users[i].AdminRole = models.UserAdminRoleUser
		CreateFixture(tx, t, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = unique + "_auth_user" + strconv.Itoa(i)
		userOrgs[i].AuthEmail = unique + users[i].Email
		CreateFixture(tx, t, &userOrgs[i])

		if err := tx.Load(&users[i], "Organizations"); err != nil {
			t.Errorf("failed to load organizations on users[%d] fixture, %s", i, err)
		}

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = userOrgs[i].ID
		accessTokenFixtures[i].AccessToken = models.HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		CreateFixture(tx, t, &accessTokenFixtures[i])
	}

	return UserFixtures{
		Organization: org,
		Users:        users,
		Locations:    locations,
	}
}

//func CreatePostFixtures(tx *pop.Connection, t *testing.T, n int) models.Posts {
//	postDestinations := make(models.Locations, n)
//	posts := make(models.Posts, n)
//	for i := range posts {
//		CreateFixture(tx, t, postDestinations[i])
//
//		posts[i].UUID = domain.GetUUID()
//		posts[i].CreatedByID = users[1].ID
//		posts[i].OrganizationID = org.ID
//		posts[i].ProviderID = nulls.NewInt(users[1].ID)
//		posts[i].DestinationID = postDestinations[i].ID
//		CreateFixture(tx, t, &posts[i])
//	}
//}
