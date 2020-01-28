package models

import (
	"testing"
)

func (ms *ModelSuite) TestFindUsersByPostID() {
	f := createProvidersFixtures(ms)
	posts := f.Posts
	rcs := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name    string
		post    Post
		wantIDs []int
	}{
		{
			name:    "first post",
			post:    posts[0],
			wantIDs: []int{rcs[0].UserID, rcs[1].UserID, rcs[2].UserID},
		},
		{
			name:    "second post",
			post:    posts[1],
			wantIDs: []int{rcs[3].UserID, rcs[4].UserID},
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

func (ms *ModelSuite) TestNewWithPostUUID() {
	f := createProvidersFixtures(ms)
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
