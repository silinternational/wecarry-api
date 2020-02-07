package actions

import (
	"fmt"

	"github.com/silinternational/wecarry-api/internal/test"
)

func (as *ActionSuite) Test_AddMeAsPotentialProvider() {

	f := test.CreatePotentialProvidersFixtures(as.DB)
	posts := f.Posts

	const qTemplate = `mutation {post: addMeAsPotentialProvider (postID: "%s")` +
		` {id title potentialProviders{id nickname}}}`

	// Add one to Post with none
	query := fmt.Sprintf(qTemplate, posts[2].UUID.String())

	var resp PostResponse

	err := as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[2].UUID.String(), resp.Post.ID, "incorrect Post UUID")
	as.Equal(posts[2].Title, resp.Post.Title, "incorrect Post title")

	want := []PotentialProvider{{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname}}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")

	// Add one to Post with two already
	query = fmt.Sprintf(qTemplate, posts[1].UUID.String())

	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[1].UUID.String(), resp.Post.ID, "incorrect Post UUID")
	as.Equal(posts[1].Title, resp.Post.Title, "incorrect Post title")

	want = []PotentialProvider{
		{ID: f.Users[2].UUID.String(), Nickname: f.Users[2].Nickname},
		{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname},
		{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
	}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")

	// Add a repeat
	query = fmt.Sprintf(qTemplate, posts[1].UUID.String())

	err = as.testGqlQuery(query, f.Users[1].Nickname, &resp)
	as.Error(err, "expected an error (unique together) but didn't get one")

	want = []PotentialProvider{
		{ID: f.Users[2].UUID.String(), Nickname: f.Users[2].Nickname},
		{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname},
		{ID: f.Users[1].UUID.String(), Nickname: f.Users[1].Nickname},
	}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")
}

func (as *ActionSuite) Test_RemoveMeAsPotentialProvider() {

	f := test.CreatePotentialProvidersFixtures(as.DB)
	posts := f.Posts

	const qTemplate = `mutation {post: removeMeAsPotentialProvider (postID: "%s")` +
		` {id title potentialProviders{id nickname}}}`

	var resp PostResponse

	query := fmt.Sprintf(qTemplate, posts[1].UUID.String())

	err := as.testGqlQuery(query, f.Users[2].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[1].UUID.String(), resp.Post.ID, "incorrect Post UUID")
	as.Equal(posts[1].Title, resp.Post.Title, "incorrect Post title")

	want := []PotentialProvider{{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")
}

func (as *ActionSuite) Test_RemovePotentialProvider() {

	f := test.CreatePotentialProvidersFixtures(as.DB)
	posts := f.Posts

	const qTemplate = `mutation {post: removePotentialProvider (postID: "%s", userID: "%s")` +
		` {id title potentialProviders{id nickname}}}`

	var resp PostResponse

	// remove third User as a potential provider on second Post
	query := fmt.Sprintf(qTemplate, posts[1].UUID.String(), f.Users[2].UUID.String())

	err := as.testGqlQuery(query, f.Users[2].Nickname, &resp)
	as.NoError(err)
	as.Equal(posts[1].UUID.String(), resp.Post.ID, "incorrect Post UUID")
	as.Equal(posts[1].Title, resp.Post.Title, "incorrect Post title")

	want := []PotentialProvider{{ID: f.Users[3].UUID.String(), Nickname: f.Users[3].Nickname}}
	as.Equal(want, resp.Post.PotentialProviders, "incorrect potential providers")
}
