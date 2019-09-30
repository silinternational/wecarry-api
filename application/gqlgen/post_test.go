package gqlgen

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/silinternational/wecarry-api/models"
)

type PostQueryFixtures struct {
	Posts       models.Posts
	Users       models.Users
	CurrentUser models.User
	ClientID    string
	AccessToken string
}

func getGqlClient() *client.Client {
	h := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)
	return c
}

func Fixtures_PostQuery(as *ActionSuite, t *testing.T) PostQueryFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   "saml",
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := as.DB.Create(org)
	if err != nil {
		t.Errorf("could not create organization for test, error: %s", err)
		t.FailNow()
	}

	// Load User test fixtures
	users := models.Users{
		{
			Email:     "user1@example.com",
			FirstName: "First",
			LastName:  "User",
			Nickname:  "User1",
			Uuid:      domain.GetUuid(),
			AdminRole: nulls.NewString(domain.AdminRoleSuperDuperAdmin),
		},
		{
			Email:     "user2@example.com",
			FirstName: "Second",
			LastName:  "User",
			Nickname:  "User2",
			Uuid:      domain.GetUuid(),
		},
	}

	for i := range users {
		if err := as.DB.Create(&users[i]); err != nil {
			t.Errorf("could not create test user ... %v", err)
			t.FailNow()
		}
	}

	// Load UserOrganization test fixtures
	userOrgs := models.UserOrganizations{
		{
			OrganizationID: org.ID,
			UserID:         users[0].ID,
			AuthID:         "auth_user1",
			AuthEmail:      users[0].Email,
		},
		{
			OrganizationID: org.ID,
			UserID:         users[1].ID,
			AuthID:         "auth_user2",
			AuthEmail:      users[1].Email,
		},
	}

	for i := range userOrgs {
		if err := as.DB.Create(&userOrgs[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	clientID := "12345678"
	accessToken := "ABCDEFGHIJKLMONPQRSTUVWXYZ123456"
	hash := models.HashClientIdAccessToken(clientID + accessToken)

	userAccessToken := models.UserAccessToken{
		UserID:             users[0].ID,
		UserOrganizationID: userOrgs[0].ID,
		AccessToken:        hash,
		ExpiresAt:          time.Now().Add(time.Hour),
	}

	if err := as.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	// Load Post test fixtures
	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           "Request",
			OrganizationID: org.ID,
			Title:          "A Request",
			Size:           "Small",
			Status:         "New",
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           "Offer",
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           "Large",
			Status:         "New",
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}

	for i := range posts {
		if err := as.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var f models.File

	// attach photo
	if err := f.Store("photo.gif", []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := posts[1].AttachPhoto(f.UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	// attach file
	if err := f.Store("dummy.pdf", []byte("%PDF-")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := posts[1].AttachFile(f.UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return PostQueryFixtures{
		Posts:       posts,
		Users:       users,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}

func (as *ActionSuite) Test_PostQuery() {
	t := as.T()
	models.ResetTables(t, as.DB)

	queryFixtures := Fixtures_PostQuery(as, t)
	userFixtures := queryFixtures.Users
	postFixtures := queryFixtures.Posts

	c := getGqlClient()

	query := `{ post(id: "` + postFixtures[1].Uuid.String() + `") { id photo { id } files { id } }}`

	var postsResp struct {
		Post struct {
			ID    string `json:"id"`
			Photo struct {
				ID string `json:"id"`
			} `json:"photo"`
			Files []struct {
				ID string `json:"id"`
			} `json:"files"`
		} `json:"post"`
	}

	TestUser = userFixtures[0]
	c.MustPost(query, &postsResp)

	if err := as.DB.Load(&(postFixtures[1]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
	}

	as.Equal(postFixtures[1].Uuid.String(), postsResp.Post.ID)
	as.Equal(postFixtures[1].PhotoFile.UUID.String(), postsResp.Post.Photo.ID)
	as.Equal(1, len(postsResp.Post.Files))
	as.Equal(postFixtures[1].Files[0].File.UUID.String(), postsResp.Post.Files[0].ID)
}
