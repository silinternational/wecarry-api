package gqlgen

import (
	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"time"
)

type threadQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Threads
	models.Messages
	models.Locations
}

func createFixtures_ThreadQuery(gs *GqlgenSuite) threadQueryFixtures {
	t := gs.T()

	org := models.Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(gs, &org)

	users := models.Users{
		{Email: t.Name() + "_user1@example.com", Nickname: t.Name() + " User1 ", Uuid: domain.GetUuid()},
		{Email: t.Name() + "_user2@example.com", Nickname: t.Name() + " User2 ", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(gs, &users[i])
	}

	userOrgs := models.UserOrganizations{
		{OrganizationID: org.ID, UserID: users[0].ID, AuthID: t.Name() + "_auth_user1", AuthEmail: users[0].Email},
		{OrganizationID: org.ID, UserID: users[1].ID, AuthID: t.Name() + "_auth_user2", AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(gs, &userOrgs[i])
	}

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
		createFixture(gs, &locations[i])
	}

	posts := models.Posts{
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Type:           models.PostTypeRequest,
			Status:         models.PostStatusCommitted,
			Title:          "A Request",
			DestinationID:  locations[0].ID,
			Size:           models.PostSizeSmall,
			NeededAfter:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			Category:       "OTHER",
		},
		{
			Uuid:           domain.GetUuid(),
			CreatedByID:    users[0].ID,
			ProviderID:     nulls.NewInt(users[0].ID),
			OrganizationID: org.ID,
			DestinationID:  locations[2].ID,
		},
	}
	for i := range posts {
		createFixture(gs, &posts[i])
	}

	threads := models.Threads{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(gs, &threads[i])
	}

	threadParticipants := models.ThreadParticipants{
		{ThreadID: threads[0].ID, UserID: posts[0].CreatedByID},
	}
	for i := range threadParticipants {
		createFixture(gs, &threadParticipants[i])
	}

	messages := models.Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[1].ID,
			Content:  "Message from " + users[1].Nickname,
		},
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Reply from " + users[0].Nickname,
		},
	}
	for i := range messages {
		createFixture(gs, &messages[i])
	}

	return threadQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        posts,
		Threads:      threads,
		Locations:    locations,
		Messages:     messages,
	}
}
