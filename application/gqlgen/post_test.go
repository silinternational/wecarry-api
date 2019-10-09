package gqlgen

import (
	"fmt"
	"net/http/httptest"
	"strconv"
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
	models.Organization
	models.Users
	models.Posts
	models.Files
	models.Threads
}

func getGqlClient() *client.Client {
	h := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)
	return c
}

func createFixture(t *testing.T, f interface{}) {
	err := models.DB.Create(f)
	if err != nil {
		t.Errorf("error creating %T fixture, %s", f, err)
		t.FailNow()
	}
}

func Fixtures_PostQuery(t *testing.T) PostQueryFixtures {
	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(t, &org)

	users := models.Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1 ", Uuid: domain.GetUuid()},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2 ", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(t, &(users[i]))
	}

	userOrgs := models.UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthID: t.Name() + "_auth_user1", AuthEmail: users[0].Email},
		{OrganizationID: org.ID, UserID: users[1].ID, AuthID: t.Name() + "_auth_user2", AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(t, &(userOrgs[i]))
	}

	posts := models.Posts{
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[0].ID,
			ReceiverID:     nulls.NewInt(users[0].ID),
			ProviderID:     nulls.NewInt(users[1].ID),
			OrganizationID: org.ID,
			Type:           PostTypeRequest.String(),
			Status:         PostStatusCommitted.String(),
			Title:          "A Request",
			Destination:    nulls.NewString("A place"),
			Origin:         nulls.NewString("Another place"),
			Size:           PostSizeSmall.String(),
			NeededAfter:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			Category:       "OTHER",
			Description:    nulls.NewString("This is a description"),
			URL:            nulls.NewString("https://www.example.com/items/101"),
			Cost:           nulls.NewFloat64(1.0),
		},
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[0].ID,
			ProviderID:     nulls.NewInt(users[0].ID),
			OrganizationID: org.ID,
		},
	}
	for i := range posts {
		createFixture(t, &(posts[i]))
	}

	threads := []models.Thread{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(t, &(threads[i]))
	}

	threadParticipants := []models.ThreadParticipant{
		{ThreadID: threads[0].ID, UserID: posts[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(t, &(threadParticipants[i]))
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	fileData := []struct {
		name    string
		content []byte
	}{
		{"photo.gif", []byte("GIF89a")},
		{"dummy.pdf", []byte("%PDF-")},
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

	if _, err := posts[0].AttachPhoto(fileFixtures[0].UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	if _, err := posts[0].AttachFile(fileFixtures[1].UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return PostQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        posts,
		Files:        fileFixtures,
		Threads:      threads,
	}
}

func (gs *GqlgenSuite) Test_PostQuery() {
	t := gs.T()
	models.ResetTables(t, models.DB)

	f := Fixtures_PostQuery(t)
	c := getGqlClient()

	query := `{ post(id: "` + f.Posts[0].Uuid.String() + `") 
		{ 
			id 
		    type
			title
			description
			destination
			origin
			size
			neededAfter
			neededBefore
			category
			status
			createdAt
			updatedAt
			myThreadID
			cost
			url
			createdBy { id }
			receiver { id }
			provider { id }
			organization { id }
			photo { id }
			files { id } 
		}}`

	var resp struct {
		Post struct {
			ID           string `json:"id"`
			Type         string `json:"type"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			Destination  string `json:"destination"`
			Origin       string `json:"origin"`
			Size         string `json:"size"`
			NeededAfter  string `json:"neededAfter"`
			NeededBefore string `json:"neededBefore"`
			Category     string `json:"category"`
			Status       string `json:"status"`
			CreatedAt    string `json:"createdAt"`
			UpdatedAt    string `json:"updatedAt"`
			MyThreadID   string `json:"myThreadID"`
			Cost         string `json:"cost"`
			Url          string `json:"url"`
			CreatedBy    struct {
				ID string `json:"id"`
			} `json:"createdBy"`
			Receiver struct {
				ID string `json:"id"`
			} `json:"receiver"`
			Provider struct {
				ID string `json:"id"`
			} `json:"provider"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			Photo struct {
				ID string `json:"id"`
			} `json:"photo"`
			Files []struct {
				ID string `json:"id"`
			} `json:"files"`
		} `json:"post"`
	}

	TestUser = f.Users[0]
	c.MustPost(query, &resp)

	gs.Equal(f.Posts[0].Uuid.String(), resp.Post.ID)
	gs.Equal(f.Posts[0].Type, resp.Post.Type)
	gs.Equal(f.Posts[0].Title, resp.Post.Title)
	gs.Equal(f.Posts[0].Description.String, resp.Post.Description)
	gs.Equal(f.Posts[0].Destination.String, resp.Post.Destination)
	gs.Equal(f.Posts[0].Origin.String, resp.Post.Origin)
	gs.Equal(f.Posts[0].Size, resp.Post.Size)
	gs.Equal(f.Posts[0].NeededAfter.Format(time.RFC3339), resp.Post.NeededAfter)
	gs.Equal(f.Posts[0].NeededBefore.Format(time.RFC3339), resp.Post.NeededBefore)
	gs.Equal(f.Posts[0].Category, resp.Post.Category)
	gs.Equal(f.Posts[0].Status, resp.Post.Status)
	gs.Equal(f.Posts[0].CreatedAt.Format(time.RFC3339), resp.Post.CreatedAt)
	gs.Equal(f.Posts[0].UpdatedAt.Format(time.RFC3339), resp.Post.UpdatedAt)
	cost, err := strconv.ParseFloat(resp.Post.Cost, 64)
	gs.NoError(err, "couldn't parse cost field as a float")
	gs.Equal(f.Posts[0].Cost.Float64, cost)
	gs.Equal(f.Posts[0].URL.String, resp.Post.Url)
	gs.Equal(f.Threads[0].Uuid.String(), resp.Post.MyThreadID)
	gs.Equal(f.Users[0].Uuid.String(), resp.Post.CreatedBy.ID)
	gs.Equal(f.Users[0].Uuid.String(), resp.Post.Receiver.ID)
	gs.Equal(f.Users[1].Uuid.String(), resp.Post.Provider.ID)
	gs.Equal(f.Organization.Uuid.String(), resp.Post.Organization.ID)
	gs.Equal(f.Files[0].UUID.String(), resp.Post.Photo.ID)
	gs.Equal(1, len(resp.Post.Files))
	gs.Equal(f.Files[1].UUID.String(), resp.Post.Files[0].ID)
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
