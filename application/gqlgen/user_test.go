package gqlgen

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// UserResponse is for marshalling User query and mutation responses
type UserResponse struct {
	User struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Nickname      string `json:"nickname"`
		CreatedAt     string `json:"createdAt"`
		UpdatedAt     string `json:"updatedAt"`
		AdminRole     string `json:"adminRole"`
		Organizations []struct {
			ID string `json:"id"`
		} `json:"organizations"`
		Posts []struct {
			ID string `json:"id"`
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
			AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin),
		},
		{
			Uuid:       domain.GetUuid(),
			Email:      t.Name() + "_user2@example.com",
			Nickname:   t.Name() + " User2",
			AdminRole:  nulls.NewString(domain.AdminRoleSalesAdmin),
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
		createFixture(gs, &(posts[i]))
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
		Organization: *org,
		Users:        users,
		Posts:        posts,
		Locations:    locations,
	}
}

// Test_UserQuery tests the User GraphQL query
func (gs *GqlgenSuite) Test_UserQuery() {
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

	allFields := `{ id email nickname adminRole photoURL posts (role: CREATEDBY) {id} organizations {id}
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
				gs.Equal(f.Users[1].AdminRole.String, resp.User.AdminRole, "incorrect AdminRole")
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
