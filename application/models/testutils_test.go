package models

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/wecarry-api/domain"
)

type UserFixtures struct {
	Organization
	Users
	UserOrganizations
	UserAccessTokens
	Locations
}

// MustCreate saves a record to the database. Panics if any error occurs.
func MustCreate(tx *pop.Connection, f interface{}) {
	value := reflect.ValueOf(f)

	if value.Type().Kind() != reflect.Ptr {
		panic("createFixture requires a pointer")
	}

	uuidField := value.Elem().FieldByName("UUID")
	if uuidField.IsValid() {
		uuidField.Set(reflect.ValueOf(domain.GetUUID()))
	}

	err := tx.Create(f)
	if err != nil {
		panic(fmt.Sprintf("error creating %T fixture, %s", f, err))
	}
}

// CreateUserFixtures generates any number of user records for testing. Locations, UserOrganizations, and
// UserAccessTokens are also created for each user. The access token for each user is the same as the user's nickname.
// All user fixtures will be assigned to the first Organization in the DB. If no Organization exists, one will be
// created.
func CreateUserFixtures(tx *pop.Connection, n int) UserFixtures {
	var org Organization
	if err := tx.First(&org); err != nil {
		org = Organization{AuthConfig: "{}"}
		MustCreate(tx, &org)
	}

	unique := domain.GetUUID().String()

	locations := createLocationFixtures(tx, n)

	users := make(Users, n)
	userOrgs := make(UserOrganizations, n)
	accessTokenFixtures := make(UserAccessTokens, n)
	for i := range users {
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].FirstName = "first" + strconv.Itoa(i)
		users[i].LastName = "last" + strconv.Itoa(i)
		users[i].AuthPhotoURL = nulls.NewString("http://example.com/" + users[i].Nickname + ".gif")
		users[i].LocationID = nulls.NewInt(locations[i].ID)
		users[i].AdminRole = UserAdminRoleUser
		MustCreate(tx, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		MustCreate(tx, &userOrgs[i])

		if err := tx.Load(&users[i], "Organizations"); err != nil {
			panic(fmt.Sprintf("failed to load organizations on users[%d] fixture, %s", i, err))
		}

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = userOrgs[i].ID
		accessTokenFixtures[i].AccessToken = HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		MustCreate(tx, &accessTokenFixtures[i])
	}

	return UserFixtures{
		Organization:      org,
		Users:             users,
		UserOrganizations: userOrgs,
		UserAccessTokens:  accessTokenFixtures,
		Locations:         locations,
	}
}

// createLocationFixtures generates any number of location records for testing.
func createLocationFixtures(tx *pop.Connection, n int) Locations {
	countries := []string{"US", "CA", "MX", "TH", "FR", "PG"}
	locations := make(Locations, n)
	for i := range locations {
		locations[i] = Location{
			Country:     countries[rand.Intn(6)],
			Description: "Random Location " + strconv.Itoa(rand.Int()),
			Latitude:    nulls.NewFloat64(rand.Float64()*180 - 90),
			Longitude:   nulls.NewFloat64(rand.Float64()*360 - 180),
		}
		MustCreate(tx, &locations[i])
	}
	return locations
}
