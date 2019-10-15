package gqlgen

import (
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

type PostResponse struct {
	Post struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Destination struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Division1   string  `json:"division1"`
			Division2   string  `json:"division2"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"destination"`
		Origin struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
			Division1   string  `json:"division1"`
			Division2   string  `json:"division2"`
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"origin"`
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
			Uuid:                   domain.GetUuid(),
			CreatedByID:            users[0].ID,
			ReceiverID:             nulls.NewInt(users[0].ID),
			ProviderID:             nulls.NewInt(users[1].ID),
			OrganizationID:         org.ID,
			Type:                   PostTypeRequest.String(),
			Status:                 PostStatusCommitted.String(),
			Title:                  "A Request",
			DestinationDescription: "A place",
			DestinationCountry:     "US",
			DestinationDivision1:   "FL",
			DestinationDivision2:   "Miami",
			DestinationLat:         nulls.NewFloat64(25.7617),
			DestinationLong:        nulls.NewFloat64(-80.1918),
			OriginDescription:      "Another place",
			OriginCountry:          "CA",
			OriginDivision1:        "Ontario",
			OriginDivision2:        "Toronto",
			OriginLat:              nulls.NewFloat64(43.6532),
			OriginLong:             nulls.NewFloat64(-79.3832),
			Size:                   PostSizeSmall.String(),
			NeededAfter:            time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			NeededBefore:           time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			Category:               "OTHER",
			Description:            nulls.NewString("This is a description"),
			URL:                    nulls.NewString("https://www.example.com/items/101"),
			Cost:                   nulls.NewFloat64(1.0),
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
			destination {description country division1 division2 latitude longitude}
			origin {description country division1 division2 latitude longitude}
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

	var resp PostResponse

	TestUser = f.Users[0]
	c.MustPost(query, &resp)

	gs.Equal(f.Posts[0].Uuid.String(), resp.Post.ID)
	gs.Equal(f.Posts[0].Type, resp.Post.Type)
	gs.Equal(f.Posts[0].Title, resp.Post.Title)
	gs.Equal(f.Posts[0].Description.String, resp.Post.Description)

	gs.Equal(f.Posts[0].DestinationDescription, resp.Post.Destination.Description)
	gs.Equal(f.Posts[0].DestinationCountry, resp.Post.Destination.Country)
	gs.Equal(f.Posts[0].DestinationDivision1, resp.Post.Destination.Division1)
	gs.Equal(f.Posts[0].DestinationDivision2, resp.Post.Destination.Division2)
	gs.Equal(f.Posts[0].DestinationLat.Float64, resp.Post.Destination.Lat)
	gs.Equal(f.Posts[0].DestinationLong.Float64, resp.Post.Destination.Long)

	gs.Equal(f.Posts[0].OriginDescription, resp.Post.Origin.Description)
	gs.Equal(f.Posts[0].OriginCountry, resp.Post.Origin.Country)
	gs.Equal(f.Posts[0].OriginDivision1, resp.Post.Origin.Division1)
	gs.Equal(f.Posts[0].OriginDivision2, resp.Post.Origin.Division2)
	gs.Equal(f.Posts[0].OriginLat.Float64, resp.Post.Origin.Lat)
	gs.Equal(f.Posts[0].OriginLong.Float64, resp.Post.Origin.Long)

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
	Posts models.Posts
	Users models.Users
	Files models.Files
}

func Fixtures_UpdatePost(t *testing.T) UpdatePostFixtures {
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

	// Load Post test fixtures
	posts := models.Posts{
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
		createFixture(t, &(posts[i]))
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
	if _, err := posts[0].AttachPhoto(fileFixtures[0].UUID.String()); err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
		t.FailNow()
	}

	return UpdatePostFixtures{
		Posts: posts,
		Users: users,
		Files: fileFixtures,
	}
}

func (gs *GqlgenSuite) Test_UpdatePost() {
	t := gs.T()
	models.ResetTables(t, models.DB)

	f := Fixtures_UpdatePost(t)
	c := getGqlClient()

	var postsResp PostResponse

	input := `id: "` + f.Posts[0].Uuid.String() + `" photoID: "` + f.Files[1].UUID.String() + `"` +
		` 
			description: "new description"
			status: COMMITTED
			destination: {description:"dest" country:"dc" division1:"dd1" division2:"dd2" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" division1:"od1" division2:"od2" latitude:3.3 longitude:4.4}
			size: TINY
			neededAfter: "2019-11-01"
			neededBefore: "2019-12-25"
			category: "cat"
			url: "example.com" 
			cost: "1.00"
		`
	query := `mutation { post: updatePost(input: {` + input + `}) { id photo { id } description status 
			destination { description country division1 division2 latitude longitude} 
			origin { description country division1 division2 latitude longitude}
			size neededAfter neededBefore category url cost}}`

	TestUser = f.Users[0]
	c.MustPost(query, &postsResp)

	if err := models.DB.Load(&(f.Posts[0]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
		t.FailNow()
	}

	gs.Equal(f.Posts[0].Uuid.String(), postsResp.Post.ID)
	gs.Equal(f.Files[1].UUID.String(), postsResp.Post.Photo.ID)
	gs.Equal("new description", postsResp.Post.Description)
	gs.Equal("COMMITTED", postsResp.Post.Status)
	gs.Equal("dest", postsResp.Post.Destination.Description)
	gs.Equal("dc", postsResp.Post.Destination.Country)
	gs.Equal("dd1", postsResp.Post.Destination.Division1)
	gs.Equal("dd2", postsResp.Post.Destination.Division2)
	gs.Equal(1.1, postsResp.Post.Destination.Lat)
	gs.Equal(2.2, postsResp.Post.Destination.Long)
	gs.Equal("origin", postsResp.Post.Origin.Description)
	gs.Equal("oc", postsResp.Post.Origin.Country)
	gs.Equal("od1", postsResp.Post.Origin.Division1)
	gs.Equal("od2", postsResp.Post.Origin.Division2)
	gs.Equal(3.3, postsResp.Post.Origin.Lat)
	gs.Equal(4.4, postsResp.Post.Origin.Long)
	gs.Equal("TINY", postsResp.Post.Size)
	gs.Equal("2019-11-01T00:00:00Z", postsResp.Post.NeededAfter)
	gs.Equal("2019-12-25T00:00:00Z", postsResp.Post.NeededBefore)
	gs.Equal("cat", postsResp.Post.Category)
	gs.Equal("example.com", postsResp.Post.Url)
	gs.Equal("1", postsResp.Post.Cost)
}

type CreatePostFixtures struct {
	models.User
	models.Organization
	models.File
}

func Fixtures_CreatePost(t *testing.T) CreatePostFixtures {
	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(t, &org)

	user := models.User{
		Email:    t.Name() + "_user1@example.com",
		Nickname: t.Name() + " User1",
		Uuid:     domain.GetUuid(),
	}
	createFixture(t, &user)

	userOrg := models.UserOrganization{
		OrganizationID: org.ID,
		UserID:         user.ID,
		AuthID:         t.Name() + "_auth_user1",
		AuthEmail:      user.Email,
	}
	createFixture(t, &userOrg)

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var fileFixture models.File
	if err := fileFixture.Store("photo.gif", []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	return CreatePostFixtures{
		User:         user,
		Organization: org,
		File:         fileFixture,
	}
}

func (gs *GqlgenSuite) Test_CreatePost() {
	t := gs.T()
	models.ResetTables(t, models.DB)

	f := Fixtures_CreatePost(t)
	c := getGqlClient()

	var postsResp PostResponse

	input := `orgID: "` + f.Organization.Uuid.String() + `"` +
		`photoID: "` + f.File.UUID.String() + `"` +
		` 
			type: REQUEST
			title: "title"
			description: "new description"
			destination: {description:"dest" country:"dc" division1:"dd1" division2:"dd2" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" division1:"od1" division2:"od2" latitude:3.3 longitude:4.4}
			size: TINY
			neededAfter: "2019-11-01"
			neededBefore: "2019-12-25"
			category: "cat"
			url: "example.com" 
			cost: "1.00"
		`
	query := `mutation { post: createPost(input: {` + input + `}) { organization { id } photo { id } type title 
			description destination { description country division1 division2 latitude longitude } 
			origin { description country division1 division2 latitude longitude }
			size neededAfter neededBefore category url cost }}`

	TestUser = f.User
	c.MustPost(query, &postsResp)

	gs.Equal(f.Organization.Uuid.String(), postsResp.Post.Organization.ID)
	gs.Equal(f.File.UUID.String(), postsResp.Post.Photo.ID)
	gs.Equal("REQUEST", postsResp.Post.Type)
	gs.Equal("title", postsResp.Post.Title)
	gs.Equal("new description", postsResp.Post.Description)
	gs.Equal("", postsResp.Post.Status)
	gs.Equal("dest", postsResp.Post.Destination.Description)
	gs.Equal("dc", postsResp.Post.Destination.Country)
	gs.Equal("dd1", postsResp.Post.Destination.Division1)
	gs.Equal("dd2", postsResp.Post.Destination.Division2)
	gs.Equal(1.1, postsResp.Post.Destination.Lat)
	gs.Equal(2.2, postsResp.Post.Destination.Long)
	gs.Equal("origin", postsResp.Post.Origin.Description)
	gs.Equal("oc", postsResp.Post.Origin.Country)
	gs.Equal("od1", postsResp.Post.Origin.Division1)
	gs.Equal("od2", postsResp.Post.Origin.Division2)
	gs.Equal(3.3, postsResp.Post.Origin.Lat)
	gs.Equal(4.4, postsResp.Post.Origin.Long)
	gs.Equal("TINY", postsResp.Post.Size)
	gs.Equal("2019-11-01T00:00:00Z", postsResp.Post.NeededAfter)
	gs.Equal("2019-12-25T00:00:00Z", postsResp.Post.NeededBefore)
	gs.Equal("cat", postsResp.Post.Category)
	gs.Equal("example.com", postsResp.Post.Url)
	gs.Equal("1", postsResp.Post.Cost)
}
