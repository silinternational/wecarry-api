package actions

import (
	"github.com/gobuffalo/nulls"
	"strconv"
	"time"

	"github.com/silinternational/wecarry-api/domain"
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
	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(as, &org)

	unique := org.UUID.String()
	users := make(models.Users, 2)
	accessTokenFixtures := make([]models.UserAccessToken, len(users))
	userOrgs := make(models.UserOrganizations, len(users))
	for i := range users {
		users[i].UUID = domain.GetUUID()
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		createFixture(as, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = unique + "_auth_user" + strconv.Itoa(i)
		userOrgs[i].AuthEmail = unique + users[i].Email
		createFixture(as, &userOrgs[i])

		accessTokenFixtures[i].UserID =             users[i].ID
		accessTokenFixtures[i].UserOrganizationID = userOrgs[i].ID
		accessTokenFixtures[i].AccessToken =        models.HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt =          time.Now().Add(time.Minute * 60)
		createFixture(as, &accessTokenFixtures[i])
	}

	posts := models.Posts{
		{
			Type:           models.PostTypeRequest,
			Status:         models.PostStatusCommitted,
			Title:          "A Request",
			Size:           models.PostSizeSmall,
		},
		{
			ProviderID:     nulls.NewInt(users[0].ID),
		},
	}
	locations := make(models.Locations, len(posts))
	for i := range posts {
		createFixture(as, &locations[i])

		posts[i].UUID = domain.GetUUID()
		posts[i].CreatedByID = users[0].ID
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
			SentByID: users[1].ID,
			Content:  "Message from " + users[1].Nickname,
		},
		{
			UUID:     domain.GetUUID(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Reply from " + users[0].Nickname,
		},
	}
	for i := range messages {
		createFixture(as, &messages[i])
	}

	return messageQueryFixtures{
		Organization: org,
		Users:        users,
		Posts:        posts,
		Threads:      threads,
		Locations:    locations,
		Messages:     messages,
	}
}
