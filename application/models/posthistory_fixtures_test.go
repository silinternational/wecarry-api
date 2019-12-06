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
	org := Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, &org)

	unique := org.UUID.String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
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

		posts[i].UUID = domain.GetUUID()
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

func createFixturesForTestPostHistory_pop(ms *ModelSuite) PostFixtures {
	org := Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, &org)

	unique := org.UUID.String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	posts := Posts{
		{Title: "Post1 Title", ProviderID: nulls.NewInt(users[1].ID)},
		{Title: "Post2 Title"},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].UUID = domain.GetUUID()
		posts[i].Status = PostStatusAccepted
		posts[i].Type = "type"
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		posts[i].ReceiverID = nulls.NewInt(users[i].ID)
		createFixture(ms, &posts[i])
	}

	pHistories := PostHistories{
		{
			Status:     PostStatusOpen,
			PostID:     posts[0].ID,
			ReceiverID: nulls.NewInt(posts[0].CreatedByID),
		},
		{
			Status:     PostStatusCommitted,
			PostID:     posts[0].ID,
			ReceiverID: nulls.NewInt(posts[0].CreatedByID),
			ProviderID: nulls.NewInt(users[1].ID),
		},
		{
			Status:     PostStatusAccepted,
			PostID:     posts[0].ID,
			ReceiverID: nulls.NewInt(posts[0].CreatedByID),
			ProviderID: nulls.NewInt(users[1].ID),
		},
	}

	for i := range pHistories {
		createFixture(ms, &pHistories[i])
	}

	return PostFixtures{
		Users:         users,
		Posts:         posts,
		PostHistories: pHistories,
	}
}

func createFixturesForTestPostHistory_createForPost(ms *ModelSuite) PostFixtures {
	org := Organization{AuthConfig: "{}", UUID: domain.GetUUID()}
	createFixture(ms, &org)

	unique := org.UUID.String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
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

		posts[i].UUID = domain.GetUUID()
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

	return PostFixtures{
		Users:         users,
		Posts:         posts,
		PostHistories: PostHistories{pHistory},
	}
}
