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
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, false)

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
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

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
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	pHistories := PostHistories{
		{
			Status:     PostStatusOpen,
			PostID:     posts[0].ID,
			ReceiverID: nulls.NewInt(posts[0].CreatedByID),
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
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, false)

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}
