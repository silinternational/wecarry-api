package test

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"

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
		accessTokenFixtures[i].UserOrganizationID = nulls.NewInt(userOrgs[i].ID)
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

// CreateRequestFixtures generates any number of request records for testing. Related Location and File records are also
// created. All request fixtures will be assigned to the first Organization in the DB. If no Organization exists,
// one will be created. All requests are created by the first User in the DB. If no User exists, one will be created.
func CreateRequestFixtures(tx *pop.Connection, n int, createFiles bool) models.Requests {
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

	futureDate := time.Now().Add(4 * domain.DurationWeek)

	requests := make(models.Requests, n)
	for i := range requests {
		requests[i].CreatedByID = user.ID
		requests[i].OrganizationID = org.ID
		requests[i].NeededBefore = nulls.NewTime(futureDate)
		requests[i].DestinationID = locations[i*2].ID
		requests[i].OriginID = nulls.NewInt(locations[i*2+1].ID)
		requests[i].Title = "title " + strconv.Itoa(i)
		requests[i].Description = nulls.NewString("description " + strconv.Itoa(i))
		requests[i].Size = models.RequestSizeSmall
		requests[i].Status = models.RequestStatusOpen
		requests[i].URL = nulls.NewString("https://www.example.com/" + strconv.Itoa(i))
		requests[i].Kilograms = nulls.NewFloat64(float64(i) * 0.1)
		requests[i].Visibility = models.RequestVisibilitySame

		if createFiles {
			requests[i].FileID = nulls.NewInt(files[i].ID)
		}

		MustCreate(tx, &requests[i])
	}

	return requests
}

// CreateLocationFixtures generates any number of location records for testing.
func CreateLocationFixtures(tx *pop.Connection, n int) models.Locations {
	countries := []string{"US", "CA", "MX", "TH", "FR", "PG"}
	locations := make(models.Locations, n)
	for i := range locations {
		// #nosec G404
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
		// #nosec G404
		if err := f.Store(strconv.Itoa(rand.Int())+".gif", []byte("GIF89a")); err != nil {
			panic(fmt.Sprintf("failed to create file fixture, %s", err))
		}
		fileFixtures[i] = f
	}
	return fileFixtures
}

// AssertStringContains makes the test fail if the string does not contain the substring.
// It outputs one line from the stack trace along with a message about the failure.
// The stack trace line chosen is the first one that contains "_test.go" in the hope
// of showing which line called this function.
func AssertStringContains(t *testing.T, haystack, needle string, outputLen int) {
	if strings.Contains(haystack, needle) {
		return
	}

	haystackOut := haystack

	if len(haystack) > outputLen {
		haystackOut = haystack[:outputLen-1]
	}

	stack := string(debug.Stack())
	stackRows := strings.Split(stack, "\n")
	testLine := ""

	for _, row := range stackRows {
		if strings.Contains(row, "_test.go") {
			testLine = "   runtime/debug.Stack  ...  " + row + "\n"
			break
		}
	}

	msg := testLine + "-- string does not contain substring --\n  " +
		haystackOut +
		" ... \n-- does not contain --\n  " +
		needle

	t.Errorf(msg)
	return
}

type PotentialProvidersFixtures struct {
	models.Users
	models.Requests
	models.PotentialProviders
}

// CreatePotentialProviderFixtures generates five PotentialProvider records for testing.
// Five User and three Request fixtures will also be created.  The Requests will
// all be created by the first user.
// The first Request will have all but the first and fifth user as a potential provider.
// The second Request will have the last two users as potential providers.
// The third Request won't have any potential providers.
// The Fifth User will be with a different Organization.
func CreatePotentialProvidersFixtures(tx *pop.Connection) PotentialProvidersFixtures {
	uf := CreateUserFixtures(tx, 5)
	requests := CreateRequestFixtures(tx, 3, false)
	providers := models.PotentialProviders{}

	// ensure the first user is actually the creator (timing issues tend to make this unreliable otherwise)
	for i := range requests {
		requests[i].CreatedByID = uf.Users[0].ID
	}
	tx.Update(&requests)

	for i, r := range requests[:2] {
		for _, u := range uf.Users[i+1 : 4] {
			c := models.PotentialProvider{RequestID: r.ID, UserID: u.ID}
			c.Create()
			providers = append(providers, c)
		}
	}

	// Put the last user in a new org
	org2 := models.Organization{
		Name:       "Extra Org",
		AuthType:   models.AuthTypeAzureAD,
		AuthConfig: "{}",
		UUID:       domain.GetUUID(),
	}
	MustCreate(tx, &org2)

	users := uf.Users

	// Switch User4's org to org2
	uo, err := users[4].FindUserOrganization(uf.Organization)
	if err != nil {
		panic("Couldn't find User4's UserOrg: " + err.Error())
	}

	uo.OrganizationID = org2.ID
	if err := tx.UpdateColumns(&uo, "organization_id"); err != nil {
		panic("Couldn't change User4's UserOrg: " + err.Error())
	}

	return PotentialProvidersFixtures{
		Users:              uf.Users,
		Requests:           requests,
		PotentialProviders: providers,
	}
}
