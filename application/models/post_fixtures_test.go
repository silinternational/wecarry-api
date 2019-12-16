package models

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

type PostFixtures struct {
	Users
	Posts
	PostHistories
	Files
	Locations
}

func CreateFixturesValidateUpdate_RequestStatus(status PostStatus, ms *ModelSuite, t *testing.T) Post {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	post := Post{
		CreatedByID:    user.ID,
		OrganizationID: org.ID,
		DestinationID:  location.ID,
		Type:           PostTypeRequest,
		Title:          "Test Request",
		Size:           PostSizeMedium,
		UUID:           domain.GetUUID(),
		Status:         status,
	}

	createFixture(ms, &post)

	return post
}

func CreatePostFixtures(ms *ModelSuite, t *testing.T, users Users) []Post {
	if err := ms.DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	locations := []Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	// Load Post test fixtures
	posts := []Post{
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeRequest,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "A Request",
			ProviderID:     nulls.NewInt(users[1].ID),
			DestinationID:  locations[0].ID,
		},
		{
			CreatedByID:    users[0].ID,
			Type:           PostTypeOffer,
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "An Offer",
			ReceiverID:     nulls.NewInt(users[1].ID),
			DestinationID:  locations[1].ID,
		},
	}
	for i := range posts {
		posts[i].Size = PostSizeMedium
		posts[i].Status = PostStatusOpen
		posts[i].UUID = domain.GetUUID()
		if err := ms.DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test post ... %v", err)
			t.FailNow()
		}
		if err := ms.DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}
	return posts
}

func createFixturesForTestPostCreate(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	posts := Posts{
		{UUID: domain.GetUUID(), Title: "title0"},
		{Title: "title1"},
		{},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].Status = PostStatusOpen
		posts[i].Type = PostTypeRequest
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = user.ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
	}
	createFixture(ms, &posts[2])

	return PostFixtures{
		Users: Users{user},
		Posts: posts,
	}
}

func createFixturesForTestPostUpdate(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	user := uf.Users[0]

	posts := Posts{
		{Title: "title"},
		{},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].UUID = domain.GetUUID()
		posts[i].Status = PostStatusOpen
		posts[i].Type = "type"
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = user.ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: Users{user},
		Posts: posts,
	}
}

func createFixturesForTestPost_manageStatusTransition_forwardProgression(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	posts := Posts{
		{Title: "Open Request", Status: PostStatusOpen},
		{Title: "Committed Request", Status: PostStatusCommitted, ProviderID: nulls.NewInt(users[0].ID)},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].UUID = domain.GetUUID()
		posts[i].Type = PostTypeRequest
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = users[i].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	postHistories := PostHistories{
		{Status: PostStatusOpen},
		{Status: PostStatusCommitted, ProviderID: posts[1].ProviderID},
	}

	for i := range postHistories {
		postHistories[i].PostID = posts[i].ID
		postHistories[i].ReceiverID = posts[i].ReceiverID
		createFixture(ms, &postHistories[i])
	}

	return PostFixtures{
		Users:         users,
		Posts:         posts,
		PostHistories: postHistories,
	}
}

func createFixturesForTestPost_manageStatusTransition_backwardProgression(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	posts := Posts{
		{Title: "Committed Request", Status: PostStatusCommitted, ProviderID: nulls.NewInt(users[1].ID)},
		{Title: "Accepted Request", Status: PostStatusAccepted, ProviderID: nulls.NewInt(users[0].ID)},
	}
	locations := make(Locations, len(posts))
	for i := range posts {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])

		posts[i].UUID = domain.GetUUID()
		posts[i].Type = PostTypeRequest
		posts[i].Size = PostSizeTiny
		posts[i].CreatedByID = users[i].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	postHistories := PostHistories{
		{Status: PostStatusOpen, PostID: posts[0].ID, ReceiverID: posts[0].ReceiverID},
		{Status: PostStatusCommitted, PostID: posts[0].ID, ReceiverID: posts[0].ReceiverID,
			ProviderID: posts[0].ProviderID},
		{Status: PostStatusOpen, PostID: posts[1].ID, ReceiverID: posts[1].ReceiverID},
		{Status: PostStatusCommitted, PostID: posts[1].ID, ReceiverID: posts[1].ReceiverID,
			ProviderID: posts[1].ProviderID},
		{Status: PostStatusAccepted, PostID: posts[1].ID, ReceiverID: posts[1].ReceiverID,
			ProviderID: posts[1].ProviderID},
	}

	for i := range postHistories {
		createFixture(ms, &postHistories[i])
	}

	return PostFixtures{
		Users:         users,
		Posts:         posts,
		PostHistories: postHistories,
	}
}

func CreateFixturesForPostsGetFiles(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	organization := uf.Organization
	user := uf.Users[0]

	location := Location{}
	createFixture(ms, &location)

	post := Post{CreatedByID: user.ID, OrganizationID: organization.ID, DestinationID: location.ID}
	createFixture(ms, &post)

	files := make(Files, 3)

	for i := range files {
		var file File
		ms.NoError(file.Store(fmt.Sprintf("file_%d.gif", i), []byte("GIF87a")),
			"failed to create file fixture")
		files[i] = file
		_, err := post.AttachFile(files[i].UUID.String())
		ms.NoError(err, "failed to attach file to post fixture")
	}

	return PostFixtures{
		Users: Users{user},
		Posts: Posts{post},
		Files: files,
	}
}

func createFixturesForPostFindByUserAndUUID(ms *ModelSuite) PostFixtures {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	users := createUserFixtures(ms.DB, 2).Users

	// both users are in org 0, but need user 0 to also be in org 1
	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	locations := make([]Location, 3)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, DestinationID: locations[0].ID},
		{CreatedByID: users[0].ID, OrganizationID: orgs[1].ID, DestinationID: locations[1].ID},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, DestinationID: locations[2].ID,
			Status: PostStatusRemoved},
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func CreateFixtures_Posts_FindByUser(ms *ModelSuite) PostFixtures {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	users := createUserFixtures(ms.DB, 2).Users

	// both users are in org 0, but need user 0 to also be in org 1
	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	locations := make([]Location, 5)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID},
		{CreatedByID: users[0].ID, OrganizationID: orgs[1].ID},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: PostStatusCompleted},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: PostStatusRemoved},
		{CreatedByID: users[1].ID, OrganizationID: orgs[0].ID},
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixtures_Posts_FilterByUserTypeAndContents(ms *ModelSuite) PostFixtures {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	unique := domain.GetUUID().String()
	users := Users{
		{Email: unique + "_user0@example.com", Nickname: unique + "User0"},
		{Email: unique + "_user1@example.com", Nickname: unique + "User1"},
	}
	for i := range users {
		users[i].UUID = domain.GetUUID()
		createFixture(ms, &users[i])
	}

	userOrgs := UserOrganizations{
		{OrganizationID: orgs[0].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[0].ID, AuthID: users[0].Email, AuthEmail: users[0].Email},
		{OrganizationID: orgs[1].ID, UserID: users[1].ID, AuthID: users[1].Email, AuthEmail: users[1].Email},
	}
	for i := range userOrgs {
		createFixture(ms, &userOrgs[i])
	}

	locations := make([]Location, 6)
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Title: "With Match"},
		{CreatedByID: users[0].ID, OrganizationID: orgs[1].ID, Title: "MXtch In Description",
			Description: nulls.NewString("This has the lower case match in it.")},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: PostStatusCompleted,
			Title: "With Match But Completed"},
		{CreatedByID: users[0].ID, OrganizationID: orgs[0].ID, Status: PostStatusRemoved,
			Title: "With Match But Removed"},
		{CreatedByID: users[1].ID, OrganizationID: orgs[1].ID, Title: "User1 No MXtch"},
		{CreatedByID: users[1].ID, OrganizationID: orgs[1].ID, Title: "User1 With MATCH"},
	}

	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		posts[i].DestinationID = locations[i].ID
		posts[i].Type = PostTypeRequest
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func CreateFixtures_Post_IsEditable(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	org := uf.Organization
	users := uf.Users

	locations := []Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{Status: PostStatusOpen},
		{Status: PostStatusCompleted},
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		posts[i].CreatedByID = users[0].ID
		posts[i].OrganizationID = org.ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForPostGetAudience(ms *ModelSuite) PostFixtures {
	orgs := make(Organizations, 2)
	for i := range orgs {
		orgs[i] = Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
		createFixture(ms, &orgs[i])
	}

	users := createUserFixtures(ms.DB, 2).Users

	locations := []Location{{}, {}}
	for i := range locations {
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{OrganizationID: orgs[0].ID}, // 2 users
		{OrganizationID: orgs[1].ID}, // no users
	}
	for i := range posts {
		posts[i].UUID = domain.GetUUID()
		posts[i].CreatedByID = users[0].ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForGetLocationForNotifications(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	org := uf.Organization
	users := uf.Users

	locations := make(Locations, 4)
	for i := range locations {
		locations[i].Description = "location " + strconv.Itoa(i)
		createFixture(ms, &locations[i])
	}

	posts := Posts{
		{
			Type:     PostTypeOffer,
			OriginID: nulls.Int{},
		},
		{
			Type:     PostTypeRequest,
			OriginID: nulls.NewInt(locations[3].ID),
		},
		{
			Type:     PostTypeRequest,
			OriginID: nulls.Int{},
		},
	}
	for i := range posts {
		posts[i].OrganizationID = org.ID
		posts[i].UUID = domain.GetUUID()
		posts[i].CreatedByID = users[0].ID
		posts[i].DestinationID = locations[i].ID
		createFixture(ms, &posts[i])
	}

	return PostFixtures{
		Users:     users,
		Posts:     posts,
		Locations: locations,
	}
}
