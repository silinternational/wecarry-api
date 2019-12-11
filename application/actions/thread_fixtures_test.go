package actions

import (
	"github.com/gobuffalo/nulls"
	"strconv"
	"time"

	"github.com/silinternational/wecarry-api/domain"
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
	org := models.Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(as, &org)

	unique := org.UUID.String()
	users := make(models.Users, 2)
	userOrgs := make(models.UserOrganizations, len(users))
	accessTokenFixtures := make([]models.UserAccessToken, len(users))
	for i := range users {
		users[i].UUID = domain.GetUUID()
		users[i].Email = unique + "_user" + strconv.Itoa(i) + "@example.com"
		users[i].Nickname = unique + "_auth_user" + strconv.Itoa(i)
		users[i].AuthPhotoURL = nulls.NewString(users[i].Nickname + ".gif")
		createFixture(as, &users[i])

		userOrgs[i].UserID = users[i].ID
		userOrgs[i].OrganizationID = org.ID
		userOrgs[i].AuthID = unique + "_auth_user" + strconv.Itoa(i)
		userOrgs[i].AuthEmail = unique + users[i].Email
		createFixture(as, &userOrgs[i])

		accessTokenFixtures[i].UserID = users[i].ID
		accessTokenFixtures[i].UserOrganizationID = userOrgs[i].ID
		accessTokenFixtures[i].AccessToken = models.HashClientIdAccessToken(users[i].Nickname)
		accessTokenFixtures[i].ExpiresAt = time.Now().Add(time.Minute * 60)
		createFixture(as, &accessTokenFixtures[i])
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
		createFixture(as, &locations[i])
	}

	posts := models.Posts{
		{
			CreatedByID:    users[0].ID,
			OrganizationID: org.ID,
			Type:           models.PostTypeRequest,
			Status:         models.PostStatusCommitted,
			Title:          "A Request",
			DestinationID:  locations[0].ID,
			Size:           models.PostSizeSmall,
		},
		{
			CreatedByID:    users[0].ID,
			ProviderID:     nulls.NewInt(users[0].ID),
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
			SentByID: users[1].ID,
			Content:  "Message from " + users[1].Nickname,
		},
		{
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "Reply from " + users[0].Nickname,
		},
	}
	for i := range messages {
		messages[i].UUID = domain.GetUUID()
		createFixture(as, &messages[i])
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
