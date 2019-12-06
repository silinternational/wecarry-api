package gqlgen

import (
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func newHandler() http.HandlerFunc {
	return handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))

}

func (gs *GqlgenSuite) TestResolver() {
	t := gs.T()
	fmt.Printf("")

	// Load Organization test fixtures
	orgFix := []models.Organization{
		{
			Name:       "ACME",
			UUID:       domain.GetUUID(),
			AuthType:   models.AuthTypeSaml,
			AuthConfig: "[]",
		},
	}
	createFixture(gs, &orgFix[0])

	// Load User test fixtures
	userFix := models.Users{
		{
			UUID:      domain.GetUUID(),
			Email:     "clark.kent@example.org",
			FirstName: "Clark",
			LastName:  "Kent",
			Nickname:  "Reporter38",
		},
	}
	createFixture(gs, &userFix[0])

	userFix[0].Organizations = []models.Organization{orgFix[0]}

	// Load USER_ORGANIZATIONS fixtures
	UserOrgsFix := models.UserOrganizations{
		{
			OrganizationID: orgFix[0].ID,
			UserID:         userFix[0].ID,
			Role:           models.UserOrganizationRoleAdmin,
		},
	}
	createFixture(gs, &UserOrgsFix[0])

	//  Load Location test fixture
	location := models.Location{}
	createFixture(gs, &location)

	//  Load Post test fixtures
	postFix := models.Posts{
		{
			CreatedByID:    userFix[0].ID,
			Type:           models.PostTypeRequest,
			OrganizationID: orgFix[0].ID,
			Status:         models.PostStatusOpen,
			Title:          "Maple Syrup",
			Size:           models.PostSizeMedium,
			UUID:           domain.GetUUID(),
			ReceiverID:     nulls.NewInt(userFix[0].ID),
			Description:    nulls.NewString("Missing my good, old, Canadian maple syrupy goodness"),
			DestinationID:  location.ID,
		},
	}
	createFixture(gs, &postFix[0])

	// Load Thread test fixtures
	threadFix := models.Threads{
		{
			UUID:   domain.GetUUID(),
			PostID: postFix[0].ID,
		},
	}
	createFixture(gs, &threadFix[0])

	// Load THREAD_PARTICIPANTS fixtures
	threadPartFix := []models.ThreadParticipant{
		{
			ThreadID: threadFix[0].ID,
			UserID:   userFix[0].ID,
		},
	}
	createFixture(gs, &threadPartFix[0])

	// Load MESSAGES fixtures
	MessageFix := models.Messages{
		{
			ThreadID: threadFix[0].ID,
			UUID:     domain.GetUUID(),
			SentByID: userFix[0].ID,
			Content:  "Any chance you can bring some PB?",
		},
	}
	createFixture(gs, &MessageFix[0])

	// Prep gql server
	c := client.New(newHandler())

	var intResults, intExpected int
	var strResults, strExpected string

	//////////////////////////////////////
	// Test simple posts query resolver
	//////////////////////////////////////

	TestUser = userFix[0]

	// It appears that everything needs to be exported in order to be recognized
	var postsResp struct {
		Posts []struct {
			ID          string            `json:"id"`
			Type        models.PostType   `json:"type"`
			Status      models.PostStatus `json:"status"`
			Title       string            `json:"title"`
			Destination struct {
				Description string `json:"description"`
			} `json:"destination"`
			Size        models.PostSize `json:"size"`
			Description string          `json:"description"`
		} `json:"posts"`
	}

	c.MustPost(
		`{posts {id type status title destination {description} size description}}`,
		&postsResp,
	)

	TestUser = models.User{}

	intResults = len(postsResp.Posts)
	intExpected = 1

	if intResults != intExpected {
		t.Errorf("bad Post count results. Expected %v, but got %v", intExpected, intResults)
		return
	}

	strResults = postsResp.Posts[0].ID
	strExpected = postFix[0].UUID.String()

	if strResults != strExpected {
		t.Errorf("bad Post ID results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	strResults = postsResp.Posts[0].Title
	strExpected = postFix[0].Title

	if strResults != strExpected {
		t.Errorf("bad Post Title results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	///////////////////////////////////////
	// Test simple threads query resolver
	//////////////////////////////////////
	var threadsResp struct {
		Threads []struct {
			ID     string `json:"id"`
			PostID string `json:"postID"`
		} `json:"threads"`
	}

	c.MustPost(`{threads {id postID}}`, &threadsResp)

	intResults = len(threadsResp.Threads)
	intExpected = 1

	if intResults != intExpected {
		t.Errorf("bad Threads count results. Expected %v, but got %v", intExpected, intResults)
		return
	}

	strResults = threadsResp.Threads[0].ID
	strExpected = threadFix[0].UUID.String()

	if strResults != strExpected {
		t.Errorf("bad thread ID results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	strResults = threadsResp.Threads[0].PostID
	strExpected = postFix[0].UUID.String()

	if strResults != strExpected {
		t.Errorf("bad thread postID results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	///////////////////////////////////////
	// Test simple users query resolver
	//////////////////////////////////////

	TestUser = userFix[0]
	TestUser.AdminRole = models.UserAdminRoleSuperAdmin

	var usersResp struct {
		Users []struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"users"`
	}

	c.MustPost(`{users {id nickname}}`, &usersResp)

	TestUser = models.User{}

	intResults = len(usersResp.Users)
	intExpected = 1

	if intResults != intExpected {
		t.Errorf("bad users count results. Expected %v, but got %v", intExpected, intResults)
		return
	}

	strResults = usersResp.Users[0].ID
	strExpected = userFix[0].UUID.String()

	if strResults != strExpected {
		t.Errorf("bad user ID results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	strResults = usersResp.Users[0].Nickname
	strExpected = userFix[0].Nickname

	if strResults != strExpected {
		t.Errorf("bad user Nickname results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}
}
