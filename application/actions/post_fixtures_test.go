package actions

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
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

	userFixtures := test.CreateUserFixtures(as.DB, t, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	posts := test.CreatePostFixtures(as.DB, 2, true)
	posts[0].Status = models.PostStatusCommitted
	posts[0].ProviderID = nulls.NewInt(users[1].ID)
	as.NoError(as.DB.Save(&posts[0]))

	threads := []models.Thread{
		{UUID: domain.GetUUID(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(as, &threads[i])
	}

	threadParticipants := []models.ThreadParticipant{
		{ThreadID: threads[0].ID, UserID: posts[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(as, &threadParticipants[i])
	}

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	var fileFixture models.File
	if err := fileFixture.Store("dummy.pdf", []byte("%PDF-")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	if _, err := posts[0].AttachFile(fileFixture.UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
		t.FailNow()
	}

	return PostQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        posts,
		Threads:      threads,
	}
}

func createFixturesForSearchRequestsQuery(as *ActionSuite) PostQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, as.T(), 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	posts := test.CreatePostFixtures(as.DB, 2, false)
	posts[0].Title = "A Match"
	as.NoError(as.DB.Save(&posts[0]))

	return PostQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        posts,
	}
}

func createFixturesForUpdatePost(as *ActionSuite) UpdatePostFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, t, 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	locations := []models.Location{
		{
			Description: "Miami, FL, USA",
			Country:     "US",
			Latitude:    nulls.NewFloat64(25.7617),
			Longitude:   nulls.NewFloat64(-80.1918),
		},
	}
	for i := range locations {
		createFixture(as, &locations[i])
	}

	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			Type:           models.PostTypeRequest,
			OrganizationID: org.ID,
			Title:          "An Offer",
			Size:           models.PostSizeLarge,
			Status:         models.PostStatusOpen,
			ReceiverID:     nulls.NewInt(users[1].ID),
			DestinationID:  locations[0].ID, // test update of existing location
			// leave OriginID nil to test adding a location
		},
	}

	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		createFixture(as, &posts[i])
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

func createFixturesForCreatePost(as *ActionSuite) CreatePostFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, t, 1)
	org := userFixtures.Organization

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
	userFixtures := test.CreateUserFixtures(as.DB, as.T(), 2)
	org := userFixtures.Organization
	users := userFixtures.Users

	posts := make(models.Posts, 1)
	locations := make(models.Locations, len(posts))
	for i := range posts {
		createFixture(as, &locations[i])

		posts[i].CreatedByID = users[0].ID
		posts[i].ReceiverID = nulls.NewInt(users[0].ID)
		posts[i].OrganizationID = org.ID
		posts[i].UUID = domain.GetUUID()
		posts[i].DestinationID = locations[i].ID
		posts[i].Title = "title"
		posts[i].Size = models.PostSizeSmall
		posts[i].Type = models.PostTypeRequest
		posts[i].Status = models.PostStatusOpen
		createFixture(as, &posts[i])
	}

	return UpdatePostStatusFixtures{
		Posts: posts,
		Users: users,
	}
}
