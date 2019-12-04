package models

import (
	"strconv"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type PostHistoryFixtures struct {
	Users
	Posts
	PostHistories
	Files
	Locations
}

func createFixturesForTestPostHistory_Load(ms *ModelSuite) PostHistoryFixtures {
	org := Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(ms, &org)

	unique := org.Uuid.String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0", Uuid: domain.GetUuid()},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{Title: "Post1 Title"},
		{Title: "Post2 Title"},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].Uuid = domain.GetUuid()
		posts[i].Status = PostStatusOpen
		posts[i].Type = "type"
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		posts[i].ReceiverID = nulls.NewInt(users[i].ID)
		createFixture(ms, &posts[i])
	}

	pHistory := PostHistory{
		Status:     PostStatusOpen,
		PostID:     posts[0].ID,
		ReceiverID: nulls.NewInt(posts[0].CreatedByID),
	}
	createFixture(ms, &pHistory)

	return PostHistoryFixtures{
		Users:         users,
		Posts:         posts,
		PostHistories: PostHistories{pHistory},
	}
}
