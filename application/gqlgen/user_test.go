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
			ID int `json:"id"`
		} `json:"organizations"`
		Posts []struct {
			ID int `json:"id"`
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
func Fixtures_UserQuery(as *ActionSuite, t *testing.T) UserQueryFixtures {
	org := &models.Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(t, org)

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		createFixture(t, &locations[i])
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
		createFixture(t, &users[i])
	}

	userOrgs := models.UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthEmail: users[0].Email},
		{OrganizationID: org.ID, UserID: users[1].ID, AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(t, &userOrgs[i])
	}

	posts := models.Posts{
		{CreatedByID: users[1].ID, Uuid: domain.GetUuid(), ProviderID: nulls.NewInt(users[1].ID)},
	}
	for i := range posts {
		createFixture(t, &(posts[i]))
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
func (as *ActionSuite) Test_UserQuery() {
	t := as.T()
	models.ResetTables(t, as.DB)

	f := Fixtures_UserQuery(as, t)
	c := getGqlClient()

	query := `{user(id: "` + f.Users[1].Uuid.String() + `") {
		id
		email
		nickname
		adminRole
		photoURL
		posts (role: CREATEDBY) {id}
		organizations {id}
		location {description country latitude longitude}
	}}`

	var resp UserResponse

	TestUser = f.Users[0]
	c.MustPost(query, &resp)

	if err := as.DB.Load(&(f.Users[1]), "PhotoFile"); err != nil {
		t.Errorf("failed to load user fixture, %s", err)
	}
	as.Equal(f.Users[1].Uuid.String(), resp.User.ID)
	as.Equal(f.Users[1].Email, resp.User.Email)
	as.Equal(f.Users[1].Nickname, resp.User.Nickname)
	as.Equal(f.Users[1].AdminRole, resp.User.AdminRole)
	as.Equal(f.Users[1].PhotoFile.URL, resp.User.PhotoURL)
	as.Regexp("^https?", resp.User.PhotoURL)
	as.Equal(1, resp.User.Posts)
	as.Equal(f.Posts[0].ID, resp.User.Posts[0].ID)
	as.Equal(1, resp.User.Organizations)
	as.Equal(f.Organization.ID, resp.User.Organizations[0].ID)
	as.Equal(f.Locations[0].Description, resp.User.Location.Description)
	as.Equal(f.Locations[0].Country, resp.User.Location.Country)
	as.Equal(f.Locations[0].Latitude.Float64, resp.User.Location.Lat)
	as.Equal(f.Locations[0].Longitude.Float64, resp.User.Location.Long)
}
