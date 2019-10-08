package gqlgen

import (
	"fmt"
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
	Files       models.Files
	ClientID    string
	AccessToken string
}

func getGqlClient() *client.Client {
	h := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)
	return c
}

func Fixtures_PostQuery(t *testing.T) PostQueryFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   models.AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := models.DB.Create(org)
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
		if err := models.DB.Create(&users[i]); err != nil {
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
		if err := models.DB.Create(&userOrgs[i]); err != nil {
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

	if err := models.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	// Load Post test fixtures
	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeRequest.String(),
			OrganizationID: org.ID,
			Title:          "A Request",
			Size:           PostSizeSmall.String(),
			Status:         PostStatusOpen.String(),
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeOffer.String(),
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           PostSizeLarge.String(),
			Status:         PostStatusOpen.String(),
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}

	for i := range posts {
		if err := models.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	// create file fixtures
	fileData := []struct {
		name    string
		content []byte
	}{
		{
			name:    "photo.gif",
			content: []byte("GIF89a"),
		},
		{
			name:    "dummy.pdf",
			content: []byte("%PDF-"),
		},
	}
	fileFixtures := make([]models.File, len(fileData))
	for i, fileDatum := range fileData {
		var f models.File
		if err := f.Store(fileDatum.name, fileDatum.content); err != nil {
			t.Errorf("failed to create file fixture, %s", err)
			t.FailNow()
		}
		fileFixtures[i] = f
	}

	// attach photo
	if _, err := posts[1].AttachPhoto(fileFixtures[0].UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	// attach file
	if _, err := posts[1].AttachFile(fileFixtures[1].UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return PostQueryFixtures{
		Posts:       posts,
		Users:       users,
		Files:       fileFixtures,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}

func (gs *GqlgenSuite) Test_PostQuery() {
	t := gs.T()
	models.ResetTables(t, models.DB)

	queryFixtures := Fixtures_PostQuery(t)
	userFixtures := queryFixtures.Users
	postFixtures := queryFixtures.Posts
	fileFixtures := queryFixtures.Files

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

	gs.Equal(postFixtures[1].Uuid.String(), postsResp.Post.ID)
	gs.Equal(fileFixtures[0].UUID.String(), postsResp.Post.Photo.ID)
	gs.Equal(1, len(postsResp.Post.Files))
	gs.Equal(fileFixtures[1].UUID.String(), postsResp.Post.Files[0].ID)
}

type UpdatePostFixtures struct {
	Posts       models.Posts
	Users       models.Users
	Files       models.Files
	ClientID    string
	AccessToken string
}

func Fixtures_UpdatePost(t *testing.T) UpdatePostFixtures {
	// Load Org test fixtures
	org := &models.Organization{
		Name:       "TestOrg1",
		Url:        nulls.String{},
		AuthType:   models.AuthTypeSaml,
		AuthConfig: "{}",
		Uuid:       domain.GetUuid(),
	}
	err := models.DB.Create(org)
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
		if err := models.DB.Create(&users[i]); err != nil {
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
		if err := models.DB.Create(&userOrgs[i]); err != nil {
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

	if err := models.DB.Create(&userAccessToken); err != nil {
		t.Errorf("could not create test userAccessToken ... %v", err)
		t.FailNow()
	}

	// Load Post test fixtures
	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeRequest.String(),
			OrganizationID: org.ID,
			Title:          "A Request",
			Size:           PostSizeSmall.String(),
			Status:         PostStatusOpen.String(),
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeOffer.String(),
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           PostSizeLarge.String(),
			Status:         PostStatusOpen.String(),
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}

	for i := range posts {
		if err := models.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	// create file fixtures
	fileData := []struct {
		name    string
		content []byte
	}{
		{
			name:    "photo.gif",
			content: []byte("GIF89a"),
		},
		{
			name:    "dummy.pdf",
			content: []byte("%PDF-"),
		},
		{
			name:    "new_photo.webp",
			content: []byte("RIFFxxxxWEBPVP"),
		},
	}
	fileFixtures := make([]models.File, len(fileData))
	for i, fileDatum := range fileData {
		var f models.File
		if err := f.Store(fileDatum.name, fileDatum.content); err != nil {
			t.Errorf("failed to create file fixture, %s", err)
			t.FailNow()
		}
		fileFixtures[i] = f
	}

	// attach photo
	if _, err := posts[1].AttachPhoto(fileFixtures[0].UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	// attach file
	if _, err := posts[1].AttachFile(fileFixtures[1].UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return UpdatePostFixtures{
		Posts:       posts,
		Users:       users,
		Files:       fileFixtures,
		ClientID:    clientID,
		AccessToken: accessToken,
	}
}

func (gs *GqlgenSuite) Test_UpdatePost() {
	t := gs.T()
	models.ResetTables(t, models.DB)

	queryFixtures := Fixtures_UpdatePost(t)
	userFixtures := queryFixtures.Users
	postFixtures := queryFixtures.Posts
	fileFixtures := queryFixtures.Files

	c := getGqlClient()

	input := `id: "` + postFixtures[1].Uuid.String() + `" photoID: "` + fileFixtures[2].UUID.String() + `"`
	query := `mutation { updatePost(input: {` + input + `}) { id photo { id } }}`

	fmt.Printf("------ query=%s\n", query)
	var postsResp struct {
		Post struct {
			ID    string `json:"id"`
			Photo struct {
				ID string `json:"id"`
			} `json:"photo"`
		} `json:"updatePost"`
	}

	TestUser = userFixtures[0]
	c.MustPost(query, &postsResp)

	if err := models.DB.Load(&(postFixtures[1]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
		t.FailNow()
	}

	gs.Equal(postFixtures[1].Uuid.String(), postsResp.Post.ID)
	gs.Equal(fileFixtures[2].UUID.String(), postsResp.Post.Photo.ID)
}
