package actions

import (
	"time"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type PostQueryFixtures struct {
	models.Organization
	models.Users
	models.Posts
	models.Threads
}

type PostsResponse struct {
	Posts []Post `json:"posts"`
}

type PostResponse struct {
	Post Post `json:"post"`
}

type Post struct {
	ID           string          `json:"id"`
	Type         models.PostType `json:"type"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	NeededBefore string          `json:"neededBefore"`
	Destination  struct {
		Description string  `json:"description"`
		Country     string  `json:"country"`
		Lat         float64 `json:"latitude"`
		Long        float64 `json:"longitude"`
	} `json:"destination"`
	Origin struct {
		Description string  `json:"description"`
		Country     string  `json:"country"`
		Lat         float64 `json:"latitude"`
		Long        float64 `json:"longitude"`
	} `json:"origin"`
	Size       models.PostSize   `json:"size"`
	Status     models.PostStatus `json:"status"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	Kilograms  float64           `json:"kilograms"`
	IsEditable bool              `json:"isEditable"`
	Url        string            `json:"url"`
	Visibility string            `json:"visibility"`
	CreatedBy  struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarURL"`
	} `json:"createdBy"`
	Receiver struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarURL"`
	} `json:"receiver"`
	Provider struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarURL"`
	} `json:"provider"`
	Organization struct {
		ID string `json:"id"`
	} `json:"organization"`
	Photo struct {
		ID string `json:"id"`
	} `json:"photo"`
	Files []struct {
		ID string `json:"id"`
	} `json:"files"`
	Meeting struct {
		ID string `json:"id"`
	} `json:"meeting"`
}

func (as *ActionSuite) Test_PostQuery() {
	f := createFixturesForPostQuery(as)

	query := `{ post(id: "` + f.Posts[0].UUID.String() + `")
		{
			id
		    type
			title
			description
			neededBefore
			destination {description country latitude longitude}
			origin {description country latitude longitude}
			size
			status
			createdAt
			updatedAt
			kilograms
			isEditable
			url
			visibility
			createdBy { id nickname avatarURL }
			receiver { id nickname avatarURL }
			provider { id nickname avatarURL }
			organization { id }
			photo { id }
			files { id }
		}}`

	var resp PostResponse

	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)

	as.Equal(f.Posts[0].UUID.String(), resp.Post.ID)
	as.Equal(f.Posts[0].Type, resp.Post.Type)
	as.Equal(f.Posts[0].Title, resp.Post.Title)
	as.Equal(f.Posts[0].Description.String, resp.Post.Description)

	wantDate := f.Posts[0].NeededBefore.Time.Format(domain.DateFormat)
	as.Equal(wantDate, resp.Post.NeededBefore, "incorrect NeededBefore date")

	as.NoError(as.DB.Load(&f.Posts[0], "Destination", "Origin", "PhotoFile", "Files.File"))

	as.Equal(f.Posts[0].Destination.Description, resp.Post.Destination.Description)
	as.Equal(f.Posts[0].Destination.Country, resp.Post.Destination.Country)
	as.Equal(f.Posts[0].Destination.Latitude.Float64, resp.Post.Destination.Lat)
	as.Equal(f.Posts[0].Destination.Longitude.Float64, resp.Post.Destination.Long)

	as.Equal(f.Posts[0].Origin.Description, resp.Post.Origin.Description)
	as.Equal(f.Posts[0].Origin.Country, resp.Post.Origin.Country)
	as.Equal(f.Posts[0].Origin.Latitude.Float64, resp.Post.Origin.Lat)
	as.Equal(f.Posts[0].Origin.Longitude.Float64, resp.Post.Origin.Long)

	as.Equal(f.Posts[0].Size, resp.Post.Size)
	as.Equal(f.Posts[0].Status, resp.Post.Status)
	as.Equal(f.Posts[0].CreatedAt.Format(time.RFC3339), resp.Post.CreatedAt)
	as.Equal(f.Posts[0].UpdatedAt.Format(time.RFC3339), resp.Post.UpdatedAt)
	as.Equal(f.Posts[0].Kilograms, resp.Post.Kilograms)
	as.Equal(f.Posts[0].URL.String, resp.Post.Url)
	as.Equal(f.Posts[0].Visibility.String(), resp.Post.Visibility)
	as.Equal(false, resp.Post.IsEditable)
	as.Equal(f.Users[0].UUID.String(), resp.Post.CreatedBy.ID, "creator ID doesn't match")
	as.Equal(f.Users[0].Nickname, resp.Post.CreatedBy.Nickname, "creator nickname doesn't match")
	as.Equal(f.Users[0].AuthPhotoURL.String, resp.Post.CreatedBy.AvatarURL, "creator avatar URL doesn't match")
	as.Equal(f.Users[0].UUID.String(), resp.Post.Receiver.ID, "receiver ID doesn't match")
	as.Equal(f.Users[0].Nickname, resp.Post.Receiver.Nickname, "receiver nickname doesn't match")
	as.Equal(f.Users[0].AuthPhotoURL.String, resp.Post.Receiver.AvatarURL, "receiver avatar URL doesn't match")
	as.Equal(f.Users[1].UUID.String(), resp.Post.Provider.ID, "provider ID doesn't match")
	as.Equal(f.Users[1].Nickname, resp.Post.Provider.Nickname, "provider nickname doesn't match")
	as.Equal(f.Users[1].AuthPhotoURL.String, resp.Post.Provider.AvatarURL, "provider avatar URL doesn't match")
	as.Equal(f.Organization.UUID.String(), resp.Post.Organization.ID)
	as.Equal(f.Posts[0].PhotoFile.UUID.String(), resp.Post.Photo.ID)
	as.Equal(1, len(resp.Post.Files))
	as.Equal(f.Posts[0].Files[0].File.UUID.String(), resp.Post.Files[0].ID)
}

func (as *ActionSuite) Test_PostsQuery() {
	f := createFixturesForPostQuery(as)

	query := `{ posts
		{
			id
			title
		}}`

	var resp PostsResponse

	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)

	as.Equal(2, len(resp.Posts))
	as.Equal(f.Posts[1].UUID.String(), resp.Posts[0].ID)
	as.Equal(f.Posts[1].Title, resp.Posts[0].Title)
}

func (as *ActionSuite) Test_UpdatePost() {
	t := as.T()

	f := createFixturesForUpdatePost(as)

	var postsResp PostResponse

	input := `id: "` + f.Posts[0].UUID.String() + `" photoID: "` + f.Files[0].UUID.String() + `"` +
		`
			description: "new description"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			origin: {description:"origin" country:"oc" latitude:3.3 longitude:4.4}
			size: TINY
			url: "example.com"
			kilograms: 22.22
			visibility: ALL
		`
	query := `mutation { post: updatePost(input: {` + input + `}) { id photo { id } description
			neededBefore
			destination { description country latitude longitude}
			origin { description country latitude longitude}
			size url kilograms visibility isEditable}}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	if err := as.DB.Load(&(f.Posts[0]), "PhotoFile", "Files"); err != nil {
		t.Errorf("failed to load post fixture, %s", err)
		t.FailNow()
	}

	as.Equal(f.Posts[0].UUID.String(), postsResp.Post.ID)
	as.Equal(f.Files[0].UUID.String(), postsResp.Post.Photo.ID)
	as.Equal("new description", postsResp.Post.Description)
	as.Equal(f.Posts[0].NeededBefore.Time.Format(domain.DateFormat), postsResp.Post.NeededBefore)
	as.Equal("dest", postsResp.Post.Destination.Description)
	as.Equal("dc", postsResp.Post.Destination.Country)
	as.Equal(1.1, postsResp.Post.Destination.Lat)
	as.Equal(2.2, postsResp.Post.Destination.Long)
	as.Equal("origin", postsResp.Post.Origin.Description)
	as.Equal("oc", postsResp.Post.Origin.Country)
	as.Equal(3.3, postsResp.Post.Origin.Lat)
	as.Equal(4.4, postsResp.Post.Origin.Long)
	as.Equal(models.PostSizeTiny, postsResp.Post.Size)
	as.Equal("example.com", postsResp.Post.Url)
	as.Equal(22.22, postsResp.Post.Kilograms)
	as.Equal("ALL", postsResp.Post.Visibility)
	as.Equal(true, postsResp.Post.IsEditable)

	// Attempt to edit a locked post
	input = `id: "` + f.Posts[0].UUID.String() + `" description: "new description"`
	query = `mutation { post: updatePost(input: {` + input + `}) { id status}}`

	as.Error(as.testGqlQuery(query, f.Users[1].Nickname, &postsResp))

	newNeededBefore := "2099-12-25"
	// Modify post's NeededBefore
	input = `id: "` + f.Posts[0].UUID.String() + `"
		neededBefore: "` + newNeededBefore + `"`
	query = `mutation { post: updatePost(input: {` + input + `}) { id neededBefore }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))
	as.Equal(newNeededBefore, postsResp.Post.NeededBefore, "incorrect NeededBefore")

	// Null out post's NeededBefore
	input = `id: "` + f.Posts[0].UUID.String() + `"	neededBefore: ""`
	query = `mutation { post: updatePost(input: {` + input + `}) { id neededBefore }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))
	as.Equal("", postsResp.Post.NeededBefore, "incorrect NeededBefore")
}

func (as *ActionSuite) Test_CreatePost() {
	f := createFixturesForCreatePost(as)

	var postsResp PostResponse

	neededBefore := "2030-12-25"

	input := `orgID: "` + f.Organization.UUID.String() + `"` +
		`photoID: "` + f.File.UUID.String() + `"` +
		`
			type: REQUEST
			title: "title"
			description: "new description"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			size: TINY
			url: "example.com"
			visibility: ALL
		`
	query := `mutation { post: createPost(input: {` + input + `}) { organization { id } photo { id } type title
			neededBefore description destination { description country latitude longitude }
			origin { description country latitude longitude }
			size url kilograms visibility }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	as.Equal(f.Organization.UUID.String(), postsResp.Post.Organization.ID)
	as.Equal(f.File.UUID.String(), postsResp.Post.Photo.ID)
	as.Equal(models.PostTypeRequest, postsResp.Post.Type)
	as.Equal("title", postsResp.Post.Title)
	as.Equal("new description", postsResp.Post.Description)
	as.Equal("", postsResp.Post.NeededBefore)
	as.Equal(models.PostStatus(""), postsResp.Post.Status)
	as.Equal("dest", postsResp.Post.Destination.Description)
	as.Equal("dc", postsResp.Post.Destination.Country)
	as.Equal(1.1, postsResp.Post.Destination.Lat)
	as.Equal(2.2, postsResp.Post.Destination.Long)
	as.Equal("", postsResp.Post.Origin.Description)
	as.Equal("", postsResp.Post.Origin.Country)
	as.Equal(0.0, postsResp.Post.Origin.Lat)
	as.Equal(0.0, postsResp.Post.Origin.Long)
	as.Equal(models.PostSizeTiny, postsResp.Post.Size)
	as.Equal("example.com", postsResp.Post.Url)
	as.Equal(0.0, postsResp.Post.Kilograms)
	as.Equal("ALL", postsResp.Post.Visibility)

	// meeting-based request
	input = `orgID: "` + f.Organization.UUID.String() + `"` +
		`meetingID: "` + f.Meetings[0].UUID.String() + `"` +
		`
			type: REQUEST
			title: "title"
			description: "new description"
			neededBefore: "` + neededBefore + `"
			destination: {description:"dest" country:"dc" latitude:1.1 longitude:2.2}
			size: TINY
			url: "example.com"
		`
	query = `mutation { post: createPost(input: {` + input + `}) {
		neededBefore destination { description country latitude longitude }
		meeting { id } }}`

	as.NoError(as.testGqlQuery(query, f.Users[0].Nickname, &postsResp))

	as.Equal(f.Meetings[0].UUID.String(), postsResp.Post.Meeting.ID)

	as.NoError(as.DB.Load(&f.Meetings[0]), "Location")
	as.Equal(f.Meetings[0].Location.Description, postsResp.Post.Destination.Description)

	as.NotNil(postsResp.Post.NeededBefore)
	as.Equal(neededBefore, postsResp.Post.NeededBefore)

	as.Equal(f.Meetings[0].Location.Country, postsResp.Post.Destination.Country)
	as.Equal(f.Meetings[0].Location.Latitude.Float64, postsResp.Post.Destination.Lat)
	as.Equal(f.Meetings[0].Location.Longitude.Float64, postsResp.Post.Destination.Long)
}

func (as *ActionSuite) Test_UpdatePostStatus() {
	f := createFixturesForUpdatePostStatus(as)

	var postsResp PostResponse

	creator := f.Users[0]
	provider := f.Users[1]

	steps := []struct {
		status  models.PostStatus
		user    models.User
		wantErr bool
	}{
		{status: models.PostStatusCommitted, user: provider, wantErr: false},
		{status: models.PostStatusAccepted, user: provider, wantErr: true},
		{status: models.PostStatusAccepted, user: creator, wantErr: false},
		{status: models.PostStatusReceived, user: provider, wantErr: true},
		{status: models.PostStatusReceived, user: creator, wantErr: false},
		{status: models.PostStatusDelivered, user: provider, wantErr: false},
		{status: models.PostStatusCompleted, user: provider, wantErr: true},
		{status: models.PostStatusCompleted, user: creator, wantErr: false},
		{status: models.PostStatusRemoved, user: creator, wantErr: true},
	}

	for _, step := range steps {
		input := `id: "` + f.Posts[0].UUID.String() + `", status: ` + step.status.String()
		query := `mutation { post: updatePostStatus(input: {` + input + `}) {id status}}`

		err := as.testGqlQuery(query, step.user.Nickname, &postsResp)
		if step.wantErr {
			as.Error(err, "user=%s, query=%s", step.user.Nickname, query)
		} else {
			as.NoError(err, "user=%s, query=%s", step.user.Nickname, query)
			as.Equal(step.status, postsResp.Post.Status)
		}
	}
}

func (as *ActionSuite) Test_SearchRequests() {
	f := createFixturesForSearchRequestsQuery(as)
	query := `{ posts: searchRequests(text: "match")
		{
			id
			title
		}}`

	var resp PostsResponse

	err := as.testGqlQuery(query, f.Users[0].Nickname, &resp)
	as.NoError(err)

	as.Equal(1, len(resp.Posts), "incorrect number of posts returned")
	as.Equal(f.Posts[0].UUID.String(), resp.Posts[0].ID)
	as.Equal(f.Posts[0].Title, resp.Posts[0].Title)
}
