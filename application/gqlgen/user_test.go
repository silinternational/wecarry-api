package gqlgen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// UserResponse is for marshalling User query and mutation responses
type UserResponse struct {
	User struct {
		ID            string               `json:"id"`
		Email         string               `json:"email"`
		Nickname      string               `json:"nickname"`
		CreatedAt     string               `json:"createdAt"`
		UpdatedAt     string               `json:"updatedAt"`
		AdminRole     models.UserAdminRole `json:"adminRole"`
		Organizations []struct {
			ID string `json:"id"`
		} `json:"organizations"`
		Posts []struct {
			ID string `json:"id"`
		}
		Preferences struct {
			Language   *string `json:"language"`
			TimeZone   *string `json:"timeZone"`
			WeightUnit *string `json:"weightUnit"`
		}
		PhotoURL string `json:"photoURL"`
		Location struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"location"`
	} `json:"user"`
}

// UserQueryFixtures is for returning fixtures from Fixtures_UserQuery
type UserQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Locations
	models.UserPreferences
}

// Fixtures_UserQuery creates fixtures for Test_UserQuery
func Fixtures_UserQuery(gs *GqlgenSuite, t *testing.T) UserQueryFixtures {
	org := &models.Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(gs, org)

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		createFixture(gs, &locations[i])
	}

	users := models.Users{
		{
			Uuid:      domain.GetUuid(),
			Email:     t.Name() + "_user1@example.com",
			Nickname:  t.Name() + " User1",
			FirstName: "First1",
			LastName:  "Last1",
			AdminRole: models.UserAdminRoleSuperAdmin,
		},
		{
			Uuid:       domain.GetUuid(),
			Email:      t.Name() + "_user2@example.com",
			Nickname:   t.Name() + " User2",
			AdminRole:  models.UserAdminRoleSalesAdmin,
			FirstName:  "First2",
			LastName:   "Last2",
			LocationID: nulls.NewInt(locations[0].ID),
		},
	}
	for i := range users {
		createFixture(gs, &users[i])
	}

	userOrgs := models.UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: org.ID, UserID: users[1].ID, AuthID: users[1].Email, AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(gs, &userOrgs[i])
	}

	location := models.Location{}
	createFixture(gs, &location)

	// Load UserPreferences test fixtures
	userPreferences := models.UserPreferences{
		{
			Uuid:   domain.GetUuid(),
			UserID: users[1].ID,
			Key:    domain.UserPreferenceKeyLanguage,
			Value:  domain.UserPreferenceLanguageFrench,
		},
		{
			Uuid:   domain.GetUuid(),
			UserID: users[1].ID,
			Key:    domain.UserPreferenceKeyTimeZone,
			Value:  "America/New_York",
		},
		{
			Uuid:   domain.GetUuid(),
			UserID: users[1].ID,
			Key:    domain.UserPreferenceKeyWeightUnit,
			Value:  domain.UserPreferenceWeightUnitPounds,
		},
	}

	for i := range userPreferences {
		createFixture(gs, &userPreferences[i])
	}

	posts := models.Posts{
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[1].ID,
			OrganizationID: org.ID,
			ProviderID:     nulls.NewInt(users[1].ID),
			DestinationID:  location.ID,
		},
	}
	for i := range posts {
		createFixture(gs, &posts[i])
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var f models.File

	if err := f.Store("photo.gif", []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := users[1].AttachPhoto(f.UUID.String()); err != nil {
		t.Errorf("failed to attach photo to user, %s", err)
		t.FailNow()
	}

	return UserQueryFixtures{
		Organization:    *org,
		Users:           users,
		UserPreferences: userPreferences,
		Posts:           posts,
		Locations:       locations,
	}
}

// TestUserQuery tests the User GraphQL query
func (gs *GqlgenSuite) TestUserQuery() {
	t := gs.T()

	f := Fixtures_UserQuery(gs, t)
	c := getGqlClient()

	type testCase struct {
		Name        string
		Payload     string
		TestUser    models.User
		ExpectError bool
		Test        func(t *testing.T)
	}

	var resp UserResponse

	allFields := `{ id email nickname adminRole photoURL preferences {language timeZone weightUnit}  
		posts (role: CREATEDBY) {id} organizations {id}
		location {description country latitude longitude} }`
	testCases := []testCase{
		{
			Name:     "all fields",
			Payload:  `{user(id: "` + f.Users[1].Uuid.String() + `")` + allFields + "}",
			TestUser: f.Users[0],
			Test: func(t *testing.T) {
				if err := gs.DB.Load(&(f.Users[1]), "PhotoFile"); err != nil {
					t.Errorf("failed to load user fixture, %s", err)
				}
				gs.Equal(f.Users[1].Uuid.String(), resp.User.ID, "incorrect ID")
				gs.Equal(f.Users[1].Email, resp.User.Email, "incorrect Email")
				gs.Equal(f.Users[1].Nickname, resp.User.Nickname, "incorrect Nickname")
				gs.Equal(f.Users[1].AdminRole, resp.User.AdminRole, "incorrect AdminRole")
				gs.Equal(f.Users[1].PhotoFile.URL, resp.User.PhotoURL, "incorrect PhotoURL")
				gs.Regexp("^https?", resp.User.PhotoURL, "invalid PhotoURL")
				gs.Equal(1, len(resp.User.Posts), "wrong number of posts")
				gs.Equal(f.Posts[0].Uuid.String(), resp.User.Posts[0].ID, "incorrect Post ID")
				gs.Equal(1, len(resp.User.Organizations), "wrong number of Organizations")
				gs.Equal(f.Organization.Uuid.String(), resp.User.Organizations[0].ID, "incorrect Organization ID")
				gs.Equal(f.Locations[0].Description, resp.User.Location.Description, "incorrect location")
				gs.Equal(f.Locations[0].Country, resp.User.Location.Country, "incorrect country")
				gs.Equal(f.Locations[0].Latitude.Float64, resp.User.Location.Lat, "incorrect latitude")
				gs.Equal(f.Locations[0].Longitude.Float64, resp.User.Location.Long, "incorrect longitude")

				gs.Equal(strings.ToUpper(f.UserPreferences[0].Value), *resp.User.Preferences.Language,
					"incorrect preference - language")
				gs.Equal(f.UserPreferences[1].Value, *resp.User.Preferences.TimeZone,
					"incorrect preference - time zone")
				gs.Equal(strings.ToUpper(f.UserPreferences[2].Value), *resp.User.Preferences.WeightUnit,
					"incorrect preference - weight unit")
			},
		},
		{
			Name:     "current user",
			Payload:  `{user ` + allFields + "}",
			TestUser: f.Users[1],
			Test: func(t *testing.T) {
				gs.Equal(f.Users[1].Uuid.String(), resp.User.ID, "incorrect ID")
			},
		},
		{
			Name:        "not allowed",
			Payload:     `{user(id: "` + f.Users[0].Uuid.String() + `")` + allFields + "}",
			TestUser:    f.Users[1],
			Test:        func(t *testing.T) {},
			ExpectError: true,
		},
	}

	for _, test := range testCases {
		TestUser = test.TestUser
		err := c.Post(test.Payload, &resp)

		if test.ExpectError {
			gs.Error(err)
		} else {
			gs.NoError(err)
		}
		t.Run(test.Name, test.Test)
	}
}

// TestUpdateUser tests the updateUser GraphQL mutation
func (gs *GqlgenSuite) TestUpdateUser() {
	t := gs.T()

	f := Fixtures_UserQuery(gs, t)
	c := getGqlClient()

	type testCase struct {
		Name        string
		Payload     string
		TestUser    models.User
		ExpectError bool
		Test        func(t *testing.T)
	}

	var resp UserResponse

	userID := f.Users[1].Uuid.String()
	newNickname := "U1 New Nickname"
	location := `{description: "Paris, France", country: "FR", latitude: 48.8588377, longitude: 2.2770202}`

	preferences := fmt.Sprintf(`{weightUnit: %s}`, strings.ToUpper(domain.UserPreferenceWeightUnitKGs))

	requestedFields := `{id nickname photoURL preferences {language, timeZone, weightUnit} location {description, country}}`

	update := fmt.Sprintf(`mutation { user: updateUser(input:{id: "%s", nickname: "%s", location: %s, preferences: %s}) %s }`,
		userID, newNickname, location, preferences, requestedFields)

	testCases := []testCase{
		{
			Name:     "allowed",
			Payload:  update,
			TestUser: f.Users[0],
			Test: func(t *testing.T) {
				if err := gs.DB.Load(&(f.Users[1]), "PhotoFile"); err != nil {
					t.Errorf("failed to load user fixture, %s", err)
				}
				gs.Equal(newNickname, resp.User.Nickname, "incorrect Nickname")
				gs.Equal(f.Users[1].PhotoFile.URL, resp.User.PhotoURL, "incorrect PhotoURL")
				gs.Regexp("^https?", resp.User.PhotoURL, "invalid PhotoURL")
				gs.Equal("Paris, France", resp.User.Location.Description, "incorrect location")
				gs.Equal("FR", resp.User.Location.Country, "incorrect country")

				gs.Equal(strings.ToUpper(f.UserPreferences[0].Value), *resp.User.Preferences.Language,
					"incorrect preference - language")
				gs.Equal(strings.ToUpper(domain.UserPreferenceWeightUnitKGs), *resp.User.Preferences.WeightUnit,
					"incorrect preference - weightUnit")
				gs.Equal("America/New_York", *resp.User.Preferences.TimeZone,
					"incorrect preference - timeZone")
			},
		},
		{
			Name: "not allowed",
			Payload: fmt.Sprintf(`mutation {updateUser(input:{id: \"%v\", location: \"%v\"}) {nickname}}`,
				f.Users[0].Uuid, location),
			TestUser:    f.Users[1],
			Test:        func(t *testing.T) {},
			ExpectError: true,
		},
	}

	for _, test := range testCases {
		TestUser = test.TestUser

		err := c.Post(test.Payload, &resp)

		if test.ExpectError {
			gs.Error(err)
		} else {
			gs.NoError(err)
		}
		t.Run(test.Name, test.Test)
	}
}
