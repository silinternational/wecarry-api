package models

import (
	"testing"
)

func (ms *ModelSuite) TestFindByPostID() {
	f := createCommittersFixtures(ms)
	posts := f.Posts
	rcs := f.RequestCommitters
	t := ms.T()
	tests := []struct {
		name    string
		post    Post
		wantIDs []int
	}{
		{
			name:    "first post",
			post:    posts[0],
			wantIDs: []int{rcs[0].ID, rcs[1].ID, rcs[2].ID},
		},
		{
			name:    "second post",
			post:    posts[1],
			wantIDs: []int{rcs[3].ID, rcs[4].ID},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			committers := RequestCommitters{}
			err := committers.FindByPostID(test.post.ID)
			ms.NoError(err, "unexpected error")
			ids := make([]int, len(committers))
			for i, c := range committers {
				ids[i] = c.ID
			}

			ms.Equal(test.wantIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestNewWithPostUUID() {
	f := createCommittersFixtures(ms)
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
			wantErr: "Request Commmitter User must not be the Post's Receiver.",
		},
		{
			name:   "good - second post second user",
			post:   posts[1],
			userID: users[1].ID,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			committer := RequestCommitter{}
			err := committer.NewWithPostUUID(test.post.UUID.String(), test.userID)
			if test.wantErr != "" {
				ms.Error(err, "expected an error but did not get one")
				ms.Equal(test.wantErr, err.Error(), "incorrect error message")
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.post.ID, committer.PostID, "incorrect Post ID")
			ms.Equal(test.userID, committer.UserID, "incorrect User ID")
		})
	}
}
