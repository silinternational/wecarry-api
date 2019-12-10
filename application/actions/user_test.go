package actions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type UsersResponse struct {
	Users []User `json:"users"`
}

type UserResponse struct {
	User User `json:"user"`
}

// UserResponse is for marshalling User query and mutation responses
type User struct {
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
	AvatarURL string `json:"avatarURL"`
	Location  struct {
		Description string  `json:"description"`
		Country     string  `json:"country"`
		Lat         float64 `json:"latitude"`
		Long        float64 `json:"longitude"`
	} `json:"location"`
}

// TestUserQuery tests the User GraphQL query
func (as *ActionSuite) TestUserQuery() {
	t := as.T()

	f := fixturesForUserQuery(as)

	type testCase struct {
		Name        string
		Payload     string
		TestUser    models.User
		ExpectError bool
		Test        func(t *testing.T)
	}

	var resp UserResponse

	allFields := `{ id email nickname adminRole avatarURL preferences {language timeZone weightUnit}
		posts (role: CREATEDBY) {id} organizations {id}
		location {description country latitude longitude} }`
	testCases := []testCase{
		{
			Name:     "all fields",
			Payload:  `{user(id: "` + f.Users[1].UUID.String() + `")` + allFields + "}",
			TestUser: f.Users[0],
			Test: func(t *testing.T) {
				if err := as.DB.Load(&(f.Users[1]), "PhotoFile"); err != nil {
					t.Errorf("failed to load user fixture, %s", err)
				}
				as.Equal(f.Users[1].UUID.String(), resp.User.ID, "incorrect ID")
				as.Equal(f.Users[1].Email, resp.User.Email, "incorrect Email")
				as.Equal(f.Users[1].Nickname, resp.User.Nickname, "incorrect Nickname")
				as.Equal(f.Users[1].AdminRole, resp.User.AdminRole, "incorrect AdminRole")
				as.Equal(f.Users[1].PhotoFile.URL, resp.User.AvatarURL, "incorrect AvatarURL")
				as.Regexp("^https?", resp.User.AvatarURL, "invalid AvatarURL")
				as.Equal(1, len(resp.User.Posts), "wrong number of posts")
				as.Equal(f.Posts[0].UUID.String(), resp.User.Posts[0].ID, "incorrect Post ID")
				as.Equal(1, len(resp.User.Organizations), "wrong number of Organizations")
				as.Equal(f.Organization.UUID.String(), resp.User.Organizations[0].ID, "incorrect Organization ID")
				as.Equal(f.Locations[0].Description, resp.User.Location.Description, "incorrect location")
				as.Equal(f.Locations[0].Country, resp.User.Location.Country, "incorrect country")
				as.Equal(f.Locations[0].Latitude.Float64, resp.User.Location.Lat, "incorrect latitude")
				as.Equal(f.Locations[0].Longitude.Float64, resp.User.Location.Long, "incorrect longitude")

				as.Equal(strings.ToUpper(f.UserPreferences[0].Value), *resp.User.Preferences.Language,
					"incorrect preference - language")
				as.Equal(f.UserPreferences[1].Value, *resp.User.Preferences.TimeZone,
					"incorrect preference - time zone")
				as.Equal(strings.ToUpper(f.UserPreferences[2].Value), *resp.User.Preferences.WeightUnit,
					"incorrect preference - weight unit")
			},
		},
		{
			Name:     "current user",
			Payload:  `{user ` + allFields + "}",
			TestUser: f.Users[1],
			Test: func(t *testing.T) {
				as.Equal(f.Users[1].UUID.String(), resp.User.ID, "incorrect ID")
			},
		},
		{
			Name:        "not allowed",
			Payload:     `{user(id: "` + f.Users[0].UUID.String() + `")` + allFields + "}",
			TestUser:    f.Users[1],
			Test:        func(t *testing.T) {},
			ExpectError: true,
		},
	}

	for _, test := range testCases {
		err := as.testGqlQuery(test.Payload, test.TestUser.Nickname, &resp)

		if test.ExpectError {
			as.Error(err)
		} else {
			as.NoError(err)
		}
		t.Run(test.Name, test.Test)
	}
}

// TestUserQuery tests the Users GraphQL query
func (as *ActionSuite) TestUsersQuery() {
	t := as.T()

	f := fixturesForUserQuery(as)

	type testCase struct {
		Name        string
		Payload     string
		TestUser    models.User
		ExpectError bool
		Test        func(t *testing.T)
	}

	var resp UsersResponse

	testCases := []testCase{
		{
			Name:     "good",
			Payload:  "{users {id nickname}}",
			TestUser: f.Users[0],
			Test: func(t *testing.T) {
				as.Equal(2, len(resp.Users), "incorrect number of users returned")
				as.Equal(f.Users[0].UUID.String(), resp.Users[0].ID, "incorrect ID")
				as.Equal(f.Users[0].Nickname, resp.Users[0].Nickname, "incorrect nickname")
			},
		},
		{
			Name:        "not allowed",
			Payload:     "{users {id nickname}}",
			TestUser:    f.Users[1],
			Test:        func(t *testing.T) {},
			ExpectError: true,
		},
	}

	for _, test := range testCases {
		err := as.testGqlQuery(test.Payload, test.TestUser.Nickname, &resp)

		if test.ExpectError {
			as.Error(err)
		} else {
			as.NoError(err)
		}
		t.Run(test.Name, test.Test)
	}
}

// TestUpdateUser tests the updateUser GraphQL mutation
func (as *ActionSuite) TestUpdateUser() {
	t := as.T()

	f := fixturesForUserQuery(as)

	type testCase struct {
		Name        string
		Payload     string
		TestUser    models.User
		ExpectError bool
		Test        func(t *testing.T)
	}

	var resp UserResponse

	userID := f.Users[1].UUID.String()
	newNickname := "U1 New Nickname"
	location := `{description: "Paris, France", country: "FR", latitude: 48.8588377, longitude: 2.2770202}`

	preferences := fmt.Sprintf(`{weightUnit: %s}`, strings.ToUpper(domain.UserPreferenceWeightUnitKGs))

	requestedFields := `{id nickname avatarURL preferences {language, timeZone, weightUnit} location {description, country}}`

	update := fmt.Sprintf(`mutation { user: updateUser(input:{id: "%s", nickname: "%s", location: %s, preferences: %s}) %s }`,
		userID, newNickname, location, preferences, requestedFields)

	testCases := []testCase{
		{
			Name:     "allowed",
			Payload:  update,
			TestUser: f.Users[0],
			Test: func(t *testing.T) {
				if err := as.DB.Load(&(f.Users[1]), "PhotoFile"); err != nil {
					t.Errorf("failed to load user fixture, %s", err)
				}
				as.Equal(newNickname, resp.User.Nickname, "incorrect Nickname")
				as.Equal(f.Users[1].PhotoFile.URL, resp.User.AvatarURL, "incorrect AvatarURL")
				as.Regexp("^https?", resp.User.AvatarURL, "invalid AvatarURL")
				as.Equal("Paris, France", resp.User.Location.Description, "incorrect location")
				as.Equal("FR", resp.User.Location.Country, "incorrect country")

				as.Equal(strings.ToUpper(f.UserPreferences[0].Value), *resp.User.Preferences.Language,
					"incorrect preference - language")
				as.Equal(strings.ToUpper(domain.UserPreferenceWeightUnitKGs), *resp.User.Preferences.WeightUnit,
					"incorrect preference - weightUnit")
				as.Equal("America/New_York", *resp.User.Preferences.TimeZone,
					"incorrect preference - timeZone")
			},
		},
		{
			Name: "not allowed",
			Payload: fmt.Sprintf(`mutation {updateUser(input:{id: \"%v\", location: \"%v\"}) {nickname}}`,
				f.Users[0].UUID, location),
			TestUser:    f.Users[1],
			Test:        func(t *testing.T) {},
			ExpectError: true,
		},
	}

	for _, test := range testCases {
		err := as.testGqlQuery(test.Payload, test.TestUser.Nickname, &resp)

		if test.ExpectError {
			as.Error(err)
		} else {
			as.NoError(err)
		}
		t.Run(test.Name, test.Test)
	}
}
