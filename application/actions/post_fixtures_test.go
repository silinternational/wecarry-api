package actions

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/test"
)

type UpdatePostFixtures struct {
	models.Posts
	models.Users
	models.Files
	models.Locations
}

type CreatePostFixtures struct {
	models.Users
	models.Organization
	models.File
}

type UpdatePostStatusFixtures struct {
	models.Posts
	models.Users
}

func createFixturesForPostQuery(as *ActionSuite) PostQueryFixtures {
	t := as.T()

	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	test.CreateFixture(as.DB, t, &org)

	userFixtures := test.CreateUserFixtures(as.DB, t, 2)

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
		{},
	}
	for i := range locations {
		test.CreateFixture(as.DB, t, &locations[i])
	}

	posts := models.Posts{
		{
			CreatedByID:    userFixtures.Users[0].ID,
			ReceiverID:     nulls.NewInt(userFixtures.Users[0].ID),
			ProviderID:     nulls.NewInt(userFixtures.Users[1].ID),
			OrganizationID: org.ID,
			Type:           models.PostTypeRequest,
			Status:         models.PostStatusCommitted,
			Title:          "A Request",
			DestinationID:  locations[0].ID,
			OriginID:       nulls.NewInt(locations[1].ID),
			Size:           models.PostSizeSmall,
			Description:    nulls.NewString("This is a description"),
			URL:            nulls.NewString("https://www.example.com/items/101"),
			Kilograms:      11.11,
		},
		{
			CreatedByID:    userFixtures.Users[0].ID,
			ProviderID:     nulls.NewInt(userFixtures.Users[0].ID),
			OrganizationID: org.ID,
			DestinationID:  locations[2].ID,
		},
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		test.CreateFixture(as.DB, t, &posts[i])
	}

	threads := []models.Thread{
		{UUID: domain.GetUUID(), PostID: posts[0].ID},
	}
	for i := range threads {
		test.CreateFixture(as.DB, t, &threads[i])
	}

	threadParticipants := []models.ThreadParticipant{
		{ThreadID: threads[0].ID, UserID: posts[0].CreatedByID},
	}
	for i := range threadParticipants {
		test.CreateFixture(as.DB, t, &threadParticipants[i])
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
		Users:        userFixtures.Users,
		Posts:        posts,
		Files:        fileFixtures,
		Threads:      threads,
		Locations:    locations,
	}
}

func createFixturesForUpdatePost(as *ActionSuite) UpdatePostFixtures {
	t := as.T()

	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	test.CreateFixture(as.DB, t, &org)

	userFixtures := test.CreateUserFixtures(as.DB, t, 2)

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		test.CreateFixture(as.DB, t, &locations[i])
	}

	posts := models.Posts{
		{
			CreatedByID:    userFixtures.Users[0].ID,
			Type:           models.PostTypeRequest,
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           models.PostSizeLarge,
			Status:         models.PostStatusOpen,
			ReceiverID:     nulls.NewInt(userFixtures.Users[1].ID),
			DestinationID:  locations[0].ID, // test update of existing location
			// leave OriginID nil to test adding a location
		},
	}

	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		test.CreateFixture(as.DB, t, &posts[i])
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
		Users: userFixtures.Users,
		Files: fileFixtures,
	}
}

func createFixturesForCreatePost(as *ActionSuite) CreatePostFixtures {
	t := as.T()

	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	test.CreateFixture(as.DB, t, &org)

	userFixtures := test.CreateUserFixtures(as.DB, t, 1)

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
		Users:        userFixtures.Users,
		Organization: org,
		File:         fileFixture,
	}
}

func createFixturesForUpdatePostStatus(as *ActionSuite) UpdatePostStatusFixtures {
	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	test.CreateFixture(as.DB, as.T(), &org)

	userFixtures := test.CreateUserFixtures(as.DB, as.T(), 2)

	posts := make(models.Posts, 1)
	locations := make(models.Locations, len(posts))
	for i := range posts {
		test.CreateFixture(as.DB, as.T(), &locations[i])

		posts[i].CreatedByID = userFixtures.Users[0].ID
		posts[i].ReceiverID = nulls.NewInt(userFixtures.Users[0].ID)
		posts[i].OrganizationID = org.ID
		posts[i].UUID = domain.GetUUID()
		posts[i].DestinationID = locations[i].ID
		posts[i].Title = "title"
		posts[i].Size = models.PostSizeSmall
		posts[i].Type = models.PostTypeRequest
		posts[i].Status = models.PostStatusOpen
		test.CreateFixture(as.DB, as.T(), &posts[i])
	}

	return UpdatePostStatusFixtures{
		Posts: posts,
		Users: userFixtures.Users,
	}
}
