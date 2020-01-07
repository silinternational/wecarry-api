package test

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type UserFixtures struct {
	models.Organization
	models.Users
	models.Locations
}

// MustCreate saves a record to the database. Panics if any error occurs.
func MustCreate(tx *pop.Connection, f interface{}) {
	value := reflect.ValueOf(f)

	if value.Type().Kind() != reflect.Ptr {
		panic("MustCreate requires a pointer")
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
	var org models.Organization
	if err := tx.First(&org); err != nil {
		org = models.Organization{AuthConfig: "{}"}
		MustCreate(tx, &org)
	}

	unique := domain.GetUUID().String()

	locations := CreateLocationFixtures(tx, n)

	users := make(models.Users, n)
	userOrgs := make(models.UserOrganizations, n)
	accessTokenFixtures := make(models.UserAccessTokens, n)
	for i := range users {
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].FirstName = "first" + strconv.Itoa(i)
		users[i].LastName = "last" + strconv.Itoa(i)
		users[i].AuthPhotoURL = nulls.NewString("http://example.com/" + users[i].Nickname + ".gif")
		users[i].LocationID = nulls.NewInt(locations[i].ID)
		users[i].AdminRole = models.UserAdminRoleUser
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
		accessTokenFixtures[i].AccessToken = models.HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		MustCreate(tx, &accessTokenFixtures[i])
	}

	return UserFixtures{
		Organization: org,
		Users:        users,
		Locations:    locations,
	}
}

// CreatePostFixtures generates any number of post records for testing. Related Location and File records are also
// created. All post fixtures will be assigned to the first Organization in the DB. If no Organization exists,
// one will be created. All posts are created by the first User in the DB. If no User exists, one will be created.
func CreatePostFixtures(tx *pop.Connection, n int, createFiles bool) models.Posts {
	var org models.Organization
	if err := tx.First(&org); err != nil {
		org = models.Organization{AuthConfig: "{}"}
		MustCreate(tx, &org)
	}

	var user models.User
	if err := tx.First(&user); err != nil {
		user = models.User{}
		MustCreate(tx, &user)
	}

	locations := CreateLocationFixtures(tx, n*2)

	var files models.Files
	if createFiles {
		files = CreateFileFixtures(n)
	}

	posts := make(models.Posts, n)
	for i := range posts {
		posts[i].CreatedByID = user.ID
		posts[i].ReceiverID = nulls.NewInt(user.ID)
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i*2].ID
		posts[i].OriginID = nulls.NewInt(locations[i*2+1].ID)
		posts[i].Title = "title " + strconv.Itoa(i)
		posts[i].Description = nulls.NewString("description " + strconv.Itoa(i))
		posts[i].Size = models.PostSizeSmall
		posts[i].Type = models.PostTypeRequest
		posts[i].Status = models.PostStatusOpen
		posts[i].URL = nulls.NewString("https://www.example.com/" + strconv.Itoa(i))
		posts[i].Kilograms = float64(i) * 0.1

		if createFiles {
			posts[i].PhotoFileID = nulls.NewInt(files[i].ID)
		}

		MustCreate(tx, &posts[i])
	}

	return posts
}

// CreateLocationFixtures generates any number of location records for testing.
func CreateLocationFixtures(tx *pop.Connection, n int) models.Locations {
	countries := []string{"US", "CA", "MX", "TH", "FR", "PG"}
	locations := make(models.Locations, n)
	for i := range locations {
		locations[i] = models.Location{
			Country:     countries[rand.Intn(6)],
			Description: "Random Location " + strconv.Itoa(rand.Int()),
			Latitude:    nulls.NewFloat64(rand.Float64()*180 - 90),
			Longitude:   nulls.NewFloat64(rand.Float64()*360 - 180),
		}
		MustCreate(tx, &locations[i])
	}
	return locations
}

func CreateFileFixtures(n int) models.Files {
	fileFixtures := make([]models.File, n)
	for i := range fileFixtures {
		var f models.File
		if err := f.Store(strconv.Itoa(rand.Int())+".gif", []byte("GIF89a")); err != nil {
			panic(fmt.Sprintf("failed to create file fixture, %s", err))
		}
		fileFixtures[i] = f
	}
	return fileFixtures
}

func AssertStringContains(haystack, needle string, outputLen int) error {
	if strings.Contains(haystack, needle) {
		return nil
	}

	haystackOut := haystack

	if len(haystack) > outputLen {
		haystackOut = haystack[:outputLen-1]
	}

	return errors.New("-- string does not contain substring --\n  " +
		haystackOut +
		" ... \n-- does not contain --\n  " +
		needle)
}
