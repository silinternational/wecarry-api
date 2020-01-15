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
	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[0].Title = "new title"
	posts[1].Title = ""

	return PostFixtures{
		Posts: posts,
	}
}

func createFixturesForTestPost_manageStatusTransition_forwardProgression(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[1].Status = PostStatusAccepted
	posts[1].CreatedByID = users[1].ID
	posts[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&posts[1]))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForTestPost_manageStatusTransition_backwardProgression(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[0].Status = PostStatusAccepted
	posts[0].CreatedByID = users[0].ID
	posts[0].ProviderID = nulls.NewInt(users[1].ID)
	posts[1].Status = PostStatusAccepted
	posts[1].CreatedByID = users[1].ID
	posts[1].ProviderID = nulls.NewInt(users[0].ID)
	ms.NoError(ms.DB.Save(&posts))

	return PostFixtures{
		Users: users,
		Posts: posts,
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
		ms.Nil(file.Store(fmt.Sprintf("file_%d.gif", i), []byte("GIF87a")),
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

	posts := createPostFixtures(ms.DB, 3, 0, false)
	posts[1].OrganizationID = orgs[1].ID
	posts[2].Status = PostStatusRemoved
	ms.NoError(ms.DB.Save(&posts))

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

	posts := createPostFixtures(ms.DB, 5, 0, false)
	posts[1].OrganizationID = orgs[1].ID
	posts[2].Status = PostStatusOpen
	posts[3].Status = PostStatusRemoved
	posts[4].CreatedByID = users[1].ID
	ms.NoError(ms.DB.Save(&posts))

	// can't go directly to "completed"
	posts[2].Status = PostStatusAccepted
	ms.NoError(ms.DB.Save(&posts[2]))
	posts[2].Status = PostStatusCompleted
	ms.NoError(ms.DB.Save(&posts[2]))

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
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[1].Status = PostStatusRemoved

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

	posts := createPostFixtures(ms.DB, 2, 0, false)
	posts[1].OrganizationID = orgs[1].ID
	ms.NoError(ms.DB.Save(&posts[1]))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}

func createFixturesForGetLocationForNotifications(ms *ModelSuite) PostFixtures {
	uf := createUserFixtures(ms.DB, 1)
	users := uf.Users

	posts := createPostFixtures(ms.DB, 2, 1, false)
	posts[0].OriginID = nulls.Int{}
	ms.NoError(ms.DB.Save(&posts[0]))

	return PostFixtures{
		Users: users,
		Posts: posts,
	}
}
