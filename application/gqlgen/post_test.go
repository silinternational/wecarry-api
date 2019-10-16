package gqlgen

import (
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type PostQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Files
	models.Threads
	models.Locations
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
			Lat         float64 `json:"latitude"`
			Long        float64 `json:"longitude"`
		} `json:"destination"`
		Origin struct {
			Description string  `json:"description"`
			Country     string  `json:"country"`
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

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
		{
			Description: "Toronto, Canada",
			Country:     "CA",
			Latitude:    nulls.NewFloat64(43.6532),
			Longitude:   nulls.NewFloat64(-79.3832),
		},
	}
	for i := range locations {
		createFixture(t, &(locations[i]))
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
			DestinationID:  nulls.NewInt(locations[0].ID),
			OriginID:       nulls.NewInt(locations[1].ID),
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
		Locations:    locations,
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
			destination {description country latitude longitude}
			origin {description country latitude longitude}
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

	gs.Equal(f.Locations[0].Description, resp.Post.Destination.Description)
	gs.Equal(f.Locations[0].Country, resp.Post.Destination.Country)
	gs.Equal(f.Locations[0].Latitude.Float64, resp.Post.Destination.Lat)
	gs.Equal(f.Locations[0].Longitude.Float64, resp.Post.Destination.Long)

	gs.Equal(f.Locations[1].Description, resp.Post.Origin.Description)
	gs.Equal(f.Locations[1].Country, resp.Post.Origin.Country)
	gs.Equal(f.Locations[1].Latitude.Float64, resp.Post.Origin.Lat)
	gs.Equal(f.Locations[1].Longitude.Float64, resp.Post.Origin.Long)

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
	models.Posts
	models.Users
	models.Files
	models.Locations
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

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		createFixture(t, &(locations[i]))
	}

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
			DestinationID:  nulls.NewInt(locations[0].ID), // test update of existing location
			// leave OriginID nil to test adding a location
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
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" latitude:3.3 longitude:4.4}
			size: TINY
			neededAfter: "2019-11-01"
			neededBefore: "2019-12-25"
			category: "cat"
			url: "example.com" 
			cost: "1.00"
		`
	query := `mutation { post: updatePost(input: {` + input + `}) { id photo { id } description status 
			destination { description country latitude longitude} 
			origin { description country latitude longitude}
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
	gs.Equal(1.1, postsResp.Post.Destination.Lat)
	gs.Equal(2.2, postsResp.Post.Destination.Long)
	gs.Equal("origin", postsResp.Post.Origin.Description)
	gs.Equal("oc", postsResp.Post.Origin.Country)
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
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			size: TINY
			neededAfter: "2019-11-01"
			neededBefore: "2019-12-25"
			category: "cat"
			url: "example.com" 
			cost: "1.00"
		`
	query := `mutation { post: createPost(input: {` + input + `}) { organization { id } photo { id } type title 
			description destination { description country latitude longitude } 
			origin { description country latitude longitude }
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
	gs.Equal(1.1, postsResp.Post.Destination.Lat)
	gs.Equal(2.2, postsResp.Post.Destination.Long)
	gs.Equal("", postsResp.Post.Origin.Description)
	gs.Equal("", postsResp.Post.Origin.Country)
	gs.Equal(0.0, postsResp.Post.Origin.Lat)
	gs.Equal(0.0, postsResp.Post.Origin.Long)
	gs.Equal("TINY", postsResp.Post.Size)
	gs.Equal("2019-11-01T00:00:00Z", postsResp.Post.NeededAfter)
	gs.Equal("2019-12-25T00:00:00Z", postsResp.Post.NeededBefore)
	gs.Equal("cat", postsResp.Post.Category)
	gs.Equal("example.com", postsResp.Post.Url)
	gs.Equal("1", postsResp.Post.Cost)
}
