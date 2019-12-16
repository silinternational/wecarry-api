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

	userFixtures := test.CreateUserFixtures(as.DB, 2)
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
	userFixtures := test.CreateUserFixtures(as.DB, 2)
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

	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	posts := test.CreatePostFixtures(as.DB, 1, false)
	posts[0].OriginID = nulls.Int{}
	as.NoError(as.DB.Save(&posts[0]))

	var fileFixture models.File
	if err := fileFixture.Store("new_photo.webp", []byte("RIFFxxxxWEBPVP")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
		t.FailNow()
	}

	return UpdatePostFixtures{
		Posts: posts,
		Users: users,
		Files: models.Files{fileFixture},
	}
}

func createFixturesForCreatePost(as *ActionSuite) CreatePostFixtures {
	t := as.T()

	userFixtures := test.CreateUserFixtures(as.DB, 1)
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
	userFixtures := test.CreateUserFixtures(as.DB, 2)
	users := userFixtures.Users

	posts := test.CreatePostFixtures(as.DB, 1, false)

	return UpdatePostStatusFixtures{
		Posts: posts,
		Users: users,
	}
}