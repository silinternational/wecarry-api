package models

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
)

func (ms *ModelSuite) TestPotentialProviders_FindUsersByPostID() {
	f := createPotentialProvidersFixtures(ms)
	posts := f.Posts
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name    string
		post    Post
		wantIDs []int
	}{
		{
			name:    "first post",
			post:    posts[0],
			wantIDs: []int{pps[0].UserID, pps[1].UserID, pps[2].UserID},
		},
		{
			name:    "second post",
			post:    posts[1],
			wantIDs: []int{pps[3].UserID, pps[4].UserID},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			providers := PotentialProviders{}
			users, err := providers.FindUsersByPostID(test.post.ID)
			ms.NoError(err, "unexpected error")
			ids := make([]int, len(users))
			for i, u := range users {
				ids[i] = u.ID
			}

			ms.Equal(test.wantIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestPotentialProvider_FindWithPostUUIDAndUserUUID() {
	f := createPotentialProvidersFixtures(ms)
	posts := f.Posts
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		post        Post
		ppUserUUID  uuid.UUID
		wantID      int
	}{
		{
			name:        "post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			ppUserUUID:  users[1].UUID,
			wantID:      pps[0].ID,
		},
		{
			name:        "potential provider as current user",
			currentUser: users[2],
			post:        posts[0],
			ppUserUUID:  users[2].UUID,
			wantID:      pps[1].ID,
		},
		{
			name:        "current user is not potential provider",
			currentUser: users[1],
			post:        posts[1],
			ppUserUUID:  users[3].UUID,
			wantID:      pps[4].ID,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.FindWithPostUUIDAndUserUUID(test.post.UUID.String(), test.ppUserUUID.String(), test.currentUser)
			ms.NoError(err, "unexpected error")
			ms.Equal(test.wantID, provider.ID)
		})
	}
}

func (ms *ModelSuite) TestPotentialProviders_DestroyAllWithPostUUID() {
	f := createPotentialProvidersFixtures(ms)
	posts := f.Posts
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		post        Post
		wantIDs     []int
		wantErr     string
	}{
		{
			name:        "good: Post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			wantIDs:     []int{pps[3].ID, pps[4].ID},
		},
		{
			name:        "bad: current user is potential provider but not Post Creator",
			currentUser: users[2],
			post:        posts[0],
			wantErr: fmt.Sprintf(`user %v has insufficient permissions to destroy PotentialProviders for Post %v`,
				users[2].ID, posts[0].ID),
		},
		{
			name:        "bad: current user is not Post Creator",
			currentUser: users[1],
			post:        posts[1],
			wantErr: fmt.Sprintf(`user %v has insufficient permissions to destroy PotentialProviders for Post %v`,
				users[1].ID, posts[1].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			providers := PotentialProviders{}
			err := providers.DestroyAllWithPostUUID(test.post.UUID.String(), test.currentUser)

			if test.wantErr != "" {
				ms.Error(err, "did not get error as expected")
				ms.Equal(test.wantErr, err.Error(), "wrong error message")
				return
			}

			ms.NoError(err, "unexpected error")

			var provs PotentialProviders
			err = DB.All(&provs)
			ms.NoError(err, "error just getting PotentialProviders back out of the DB.")

			pIDs := make([]int, len(provs))
			for i, p := range provs {
				pIDs[i] = p.ID
			}

			ms.Equal(test.wantIDs, pIDs)
		})
	}
}

func (ms *ModelSuite) TestNewWithPostUUID() {
	f := createPotentialProvidersFixtures(ms)
	users := f.Users
	posts := f.Posts

	t := ms.T()
	tests := []struct {
		name    string
		post    Post
		userID  int
		wantIDs []int
		wantErr string
	}{
		{
			name:    "bad - using post's CreatedBy",
			post:    posts[0],
			userID:  users[0].ID,
			wantErr: "PotentialProvider User must not be the Post's Receiver.",
		},
		{
			name:   "good - second post second user",
			post:   posts[1],
			userID: users[1].ID,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.NewWithPostUUID(test.post.UUID.String(), test.userID)
			if test.wantErr != "" {
				ms.Error(err, "expected an error but did not get one")
				ms.Equal(test.wantErr, err.Error(), "incorrect error message")
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.post.ID, provider.PostID, "incorrect Post ID")
			ms.Equal(test.userID, provider.UserID, "incorrect User ID")
		})
	}
}

func (ms *ModelSuite) TestPotentialProvider_Validate() {
	f := createPotentialProvidersFixtures(ms)
	users := f.Users
	posts := f.Posts

	t := ms.T()
	tests := []struct {
		name     string
		postID   int
		userID   int
		wantIDs  []int
		wantErrs map[string][]string
	}{
		{
			name:     "good - second post second user",
			postID:   posts[1].ID,
			userID:   users[1].ID,
			wantErrs: map[string][]string{},
		},
		{
			name:   "bad - duplicate",
			postID: posts[1].ID,
			userID: users[3].ID,
			wantErrs: map[string][]string{
				"unique_together": {
					fmt.Sprintf("Duplicate potential provider exists with PostID: %v and UserID: %v", posts[1].ID, users[3].ID)}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{PostID: test.postID, UserID: test.userID}
			vErrors, err := provider.Validate(ms.DB)

			ms.NoError(err, "unexpected error")
			ms.Equal(test.wantErrs, vErrors.Errors, "incorrect validation errors")
		})
	}
}
