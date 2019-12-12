package actions

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

type messageQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Threads
	models.Messages
	models.Locations
}

func createFixtures_MessageQuery(as *ActionSuite) messageQueryFixtures {
	userFixtures := test.CreateUserFixtures(as.DB, as.T(), 2)
	org := userFixtures.Organization

	posts := models.Posts{
		{
			Type:   models.PostTypeRequest,
			Status: models.PostStatusCommitted,
			Title:  "A Request",
			Size:   models.PostSizeSmall,
		},
		{
			ProviderID: nulls.NewInt(userFixtures.Users[0].ID),
		},
	}
	locations := make(models.Locations, len(posts))
	for i := range posts {
		createFixture(as, &locations[i])

		posts[i].UUID = domain.GetUUID()
		posts[i].CreatedByID = userFixtures.Users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
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
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: userFixtures.Users[1].ID,
			Content:  "Message from " + userFixtures.Users[1].Nickname,
		},
		{
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: userFixtures.Users[0].ID,
			Content:  "Reply from " + userFixtures.Users[0].Nickname,
		},
	}
	for i := range messages {
		createFixture(as, &messages[i])
	}

	return messageQueryFixtures{
		Organization: org,
		Users:        userFixtures.Users,
		Posts:        posts,
		Threads:      threads,
		Locations:    locations,
		Messages:     messages,
	}
}
