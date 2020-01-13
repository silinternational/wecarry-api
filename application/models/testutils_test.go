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

// mustCreate saves a record to the database. Panics if any error occurs.
func mustCreate(tx *pop.Connection, f interface{}) {
	value := reflect.ValueOf(f)

	if value.Type().Kind() != reflect.Ptr {
		panic("mustCreate requires a pointer")
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

// createOrganizationFixtures generates any number of organization records for testing.
//  Their names will be called "Org1", "Org2", ...
func createOrganizationFixtures(tx *pop.Connection, n int) Organizations {
	var file File
	if err := file.Store("logo.gif", []byte("GIF89a")); err != nil {
		panic("unexpected error storing org logo")
	}

	organizations := make(Organizations, n)
	for i := range organizations {
		organizations[i].Name = fmt.Sprintf("Org%v", i+1)
		organizations[i].AuthType = AuthTypeSaml
		organizations[i].AuthConfig = "{}"
		organizations[i].LogoFileID = nulls.NewInt(file.ID)
		mustCreate(tx, &organizations[i])
	}

	return organizations
}

// createUserFixtures generates any number of user records for testing. Locations, UserOrganizations, and
// UserAccessTokens are also created for each user. The access token for each user is the same as the user's nickname.
// All user fixtures will be assigned to the first Organization in the DB. If no Organization exists, one will be
// created.
func createUserFixtures(tx *pop.Connection, n int) UserFixtures {
	var org Organization
	if err := tx.First(&org); err != nil {
		org = Organization{AuthConfig: "{}"}
		mustCreate(tx, &org)
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
		mustCreate(tx, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = users[i].Email
		userOrgs[i].AuthEmail = users[i].Email
		mustCreate(tx, &userOrgs[i])

		if err := tx.Load(&users[i], "Organizations"); err != nil {
			panic(fmt.Sprintf("failed to load organizations on users[%d] fixture, %s", i, err))
		}

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = userOrgs[i].ID
		accessTokenFixtures[i].AccessToken = HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		mustCreate(tx, &accessTokenFixtures[i])
	}

	return UserFixtures{
		Organization:      org,
		Users:             users,
		UserOrganizations: userOrgs,
		UserAccessTokens:  accessTokenFixtures,
		Locations:         locations,
	}
}

// CreatePostFixtures generates any number of post records for testing. Related Location and File records are also
// created. All post fixtures will be assigned to the first Organization in the DB. If no Organization exists,
// one will be created. All posts are created by the first User in the DB. If no User exists, one will be created.
func createPostFixtures(tx *pop.Connection, nRequests, nOffers int, createFiles bool) Posts {
	var org Organization
	if err := tx.First(&org); err != nil {
		org = Organization{AuthConfig: "{}"}
		mustCreate(tx, &org)
	}

	var user User
	if err := tx.First(&user); err != nil {
		user = User{}
		mustCreate(tx, &user)
	}

	totalPosts := nRequests + nOffers
	locations := createLocationFixtures(tx, totalPosts*2)

	var files Files
	if createFiles {
		files = createFileFixtures(totalPosts)
	}

	posts := make(Posts, totalPosts)
	created := 0
	for i := range posts {
		if created < nRequests {
			posts[i].Type = PostTypeRequest
			posts[i].ReceiverID = nulls.NewInt(user.ID)
		} else {
			posts[i].Type = PostTypeOffer
			posts[i].ProviderID = nulls.NewInt(user.ID)
		}
		posts[i].CreatedByID = user.ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i*2].ID
		posts[i].OriginID = nulls.NewInt(locations[i*2+1].ID)
		posts[i].Title = "title " + strconv.Itoa(i)
		posts[i].Description = nulls.NewString("description " + strconv.Itoa(i))
		posts[i].Size = PostSizeSmall
		posts[i].Status = PostStatusOpen
		posts[i].URL = nulls.NewString("https://www.example.com/" + strconv.Itoa(i))
		posts[i].Kilograms = float64(i) * 0.1

		if createFiles {
			posts[i].PhotoFileID = nulls.NewInt(files[i].ID)
		}

		mustCreate(tx, &posts[i])
		created++
	}

	return posts
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
		mustCreate(tx, &locations[i])
	}
	return locations
}

func createFileFixtures(n int) Files {
	fileFixtures := make([]File, n)
	for i := range fileFixtures {
		var f File
		if err := f.Store(strconv.Itoa(rand.Int())+".gif", []byte("GIF89a")); err != nil {
			panic(fmt.Sprintf("failed to create file fixture, %s", err))
		}
		fileFixtures[i] = f
	}
	return fileFixtures
}
