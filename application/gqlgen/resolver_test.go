package gqlgen

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/handler"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func newHandler() http.HandlerFunc {
	return handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{}}))

}

func (gs *GqlgenSuite) TestResolver() {
	t := gs.T()
	models.BounceTestDB()
	fmt.Printf("")

	// Load Organization test fixtures
	orgUuid1, _ := uuid.FromString("51b5321d-2769-48a0-908a-7af1d15083e2")
	orgFix := []models.Organization{
		{
			ID:         1,
			Name:       "ACME",
			Uuid:       orgUuid1,
			AuthType:   models.AuthTypeSaml,
			AuthConfig: "[]",
		},
	}
	if err := models.CreateOrgs(orgFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	// Load User test fixtures
	userUuid1, _ := uuid.FromString("0265d116-b54e-4712-952f-eae1d6bcdcd1")
	userFix := models.Users{
		{
			ID:        1,
			Uuid:      userUuid1,
			Email:     "clark.kent@example.org",
			FirstName: "Clark",
			LastName:  "Kent",
			Nickname:  "Reporter38",
		},
	}
	if err := models.CreateUsers(userFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	userFix[0].Organizations = []models.Organization{orgFix[0]}

	// Load USER_ORGANIZATIONS fixtures
	UserOrgsFix := models.UserOrganizations{
		{
			ID:             1,
			OrganizationID: 1,
			UserID:         1,
			Role:           RoleAdmin.String(),
		},
	}
	if err := models.CreateUserOrgs(UserOrgsFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	//  Load Post test fixtures
	postUuid1, _ := uuid.FromString("c67d507b-6c1c-4d0a-b1e6-d726a5b48c26")
	postFix := models.Posts{
		{
			ID:             1,
			CreatedByID:    1,
			Type:           PostTypeRequest.String(),
			OrganizationID: 1,
			Status:         PostStatusOpen.String(),
			Title:          "Maple Syrup",
			Destination:    nulls.NewString("Madrid, Spain"),
			Size:           PostSizeMedium.String(),
			Uuid:           postUuid1,
			ReceiverID:     nulls.NewInt(1),
			NeededAfter:    time.Date(2019, time.July, 19, 0, 0, 0, 0, time.UTC),
			NeededBefore:   time.Date(2019, time.August, 3, 0, 0, 0, 0, time.UTC),
			Category:       "Unknown",
			Description:    nulls.NewString("Missing my good, old, Canadian maple syrupy goodness"),
		},
	}
	if err := models.CreatePosts(postFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	// Load Thread test fixtures
	threadUuid1, _ := uuid.FromString("bdb7515d-06a9-4896-97a4-aeae962b85e2")
	threadFix := models.Threads{
		{
			ID:     1,
			Uuid:   threadUuid1,
			PostID: 1,
		},
	}
	if err := models.CreateThreads(threadFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	// Load THREAD_PARTICIPANTS fixtures
	threadPartFix := []models.ThreadParticipant{
		{
			ID:       1,
			ThreadID: 1,
			UserID:   1,
		},
	}
	if err := models.CreateThreadParticipants(threadPartFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	// Load MESSAGES fixtures
	messageUuid1, _ := uuid.FromString("b0d7c515-e74c-4af7-a937-f1deb9369831")
	MessageFix := models.Messages{
		{
			ThreadID: 1,
			ID:       1,
			Uuid:     messageUuid1,
			SentByID: 1,
			Content:  "Any chance you can bring some PB?",
		},
	}

	if err := models.CreateMessages(MessageFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

	// Prep gql server
	h := newHandler()
	srv := httptest.NewServer(h)
	c := client.New(srv.URL)

	var intResults, intExpected int
	var strResults, strExpected string

	//////////////////////////////////////
	// Test simple posts query resolver
	//////////////////////////////////////

	TestUser = userFix[0]

	// It appears that everything needs to be exported in order to be recognized
	var postsResp struct {
		Posts []struct {
			ID           string `json:"id"`
			Type         string `json:"type"`
			Status       string `json:"status"`
			Title        string `json:"title"`
			Destination  string `json:"destination"`
			Size         string `json:"size"`
			NeededAfter  string `json:"neededAfter"`
			NeededBefore string `json:"neededBefore"`
			Category     string `json:"category"`
			Description  string `json:"description"`
		} `json:"posts"`
	}

	c.MustPost(
		`{posts {id type status title destination size neededAfter neededBefore category description}}`,
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
	strExpected = postFix[0].Uuid.String()

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
	strExpected = threadFix[0].Uuid.String()

	if strResults != strExpected {
		t.Errorf("bad thread ID results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	strResults = threadsResp.Threads[0].PostID
	strExpected = postFix[0].Uuid.String()

	if strResults != strExpected {
		t.Errorf("bad thread postID results. \n  Expected %v, \n   but got %v", strExpected, strResults)
		return
	}

	///////////////////////////////////////
	// Test simple users query resolver
	//////////////////////////////////////

	TestUser = userFix[0]
	TestUser.AdminRole = nulls.NewString(domain.AdminRoleSuperDuperAdmin)

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
	strExpected = userFix[0].Uuid.String()

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

	models.BounceTestDB()
}
