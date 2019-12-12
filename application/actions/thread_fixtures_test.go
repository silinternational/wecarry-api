package actions

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type threadQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Threads
	models.Messages
	models.Locations
}

func createFixturesForThreadQuery(as *ActionSuite) threadQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, as.T(), 2)
	org := userFixtures.Organization

	locations := models.Locations{
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
		createFixture(as, &locations[i])
	}

	posts := models.Posts{
		{
			CreatedByID:    userFixtures.Users[0].ID,
			OrganizationID: org.ID,
			Type:           models.PostTypeRequest,
			Status:         models.PostStatusCommitted,
			Title:          "A Request",
			DestinationID:  locations[0].ID,
			Size:           models.PostSizeSmall,
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
		createFixture(as, &posts[i])
	}

	threads := models.Threads{
		{UUID: domain.GetUUID(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(as, &threads[i])
	}

	threadParticipants := models.ThreadParticipants{
		{ThreadID: threads[0].ID, UserID: posts[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(as, &threadParticipants[i])
	}

	messages := models.Messages{
		{
			ThreadID: threads[0].ID,
			SentByID: userFixtures.Users[1].ID,
			Content:  "Message from " + userFixtures.Users[1].Nickname,
		},
		{
			ThreadID: threads[0].ID,
			SentByID: userFixtures.Users[0].ID,
			Content:  "Reply from " + userFixtures.Users[0].Nickname,
		},
	}
	for i := range messages {
		messages[i].UUID = domain.GetUUID()
		createFixture(as, &messages[i])
	}

	return threadQueryFixtures{
		Organization: org,
		Users:        userFixtures.Users,
		Posts:        posts,
		Threads:      threads,
		Locations:    locations,
		Messages:     messages,
	}
}
