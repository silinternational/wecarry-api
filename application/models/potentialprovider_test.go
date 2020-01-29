package models

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
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

func (ms *ModelSuite) TestFindWithPostUUIDAndUserID() {
	f := createProvidersFixtures(ms)
	posts := f.Posts
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		post        Post
		ppUserID    int
		wantID      int
		wantErr     string
	}{
		{
			name:        "good: post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			ppUserID:    users[1].ID,
			wantID:      pps[0].ID,
		},
		{
			name:        "good: potential provider as current user",
			currentUser: users[2],
			post:        posts[0],
			ppUserID:    users[2].ID,
			wantID:      pps[1].ID,
		},
		{
			name:        "bad: current user is not potential provider",
			currentUser: users[1],
			post:        posts[1],
			ppUserID:    users[3].ID,
			wantErr: fmt.Sprintf("user %v has insufficient permissions to access PotentialProvider %v",
				users[1].ID, pps[4].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.FindWithPostUUIDAndUserID(test.post.UUID.String(), test.ppUserID, test.currentUser)

			if test.wantErr != "" {
				ms.Error(err, "did not get error as expected")
				ms.Equal(test.wantErr, err.Error(), "wrong error message")
				return
			}

			ms.NoError(err, "unexpected error")
			ms.Equal(test.wantID, provider.ID)
		})
	}
}

func (ms *ModelSuite) TestFindWithPostUUIDAndUserUUID() {
	f := createProvidersFixtures(ms)
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
		wantErr     string
	}{
		{
			name:        "good: post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			ppUserUUID:  users[1].UUID,
			wantID:      pps[0].ID,
		},
		{
			name:        "good: potential provider as current user",
			currentUser: users[2],
			post:        posts[0],
			ppUserUUID:  users[2].UUID,
			wantID:      pps[1].ID,
		},
		{
			name:        "bad: current user is not potential provider",
			currentUser: users[1],
			post:        posts[1],
			ppUserUUID:  users[3].UUID,
			wantErr: fmt.Sprintf("user %v has insufficient permissions to access PotentialProvider %v",
				users[1].ID, pps[4].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.FindWithPostUUIDAndUserUUID(test.post.UUID.String(), test.ppUserUUID.String(), test.currentUser)

			if test.wantErr != "" {
				ms.Error(err, "did not get error as expected")
				ms.Equal(test.wantErr, err.Error(), "wrong error message")
				return
			}

			ms.NoError(err, "unexpected error")
			ms.Equal(test.wantID, provider.ID)
		})
	}
}

func (ms *ModelSuite) TestDestroyWithPostUUIDAndUserID() {
	f := createProvidersFixtures(ms)
	posts := f.Posts
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		post        Post
		ppUserID    int
		wantIDs     []int
		wantErr     string
	}{
		{
			name:        "good: post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			ppUserID:    users[1].ID,
			wantIDs:     []int{pps[1].ID, pps[2].ID, pps[3].ID, pps[4].ID},
		},
		{
			name:        "good: potential provider as current user",
			currentUser: users[2],
			post:        posts[0],
			ppUserID:    users[2].ID,
			wantIDs:     []int{pps[2].ID, pps[3].ID, pps[4].ID},
		},
		{
			name:        "bad: current user is not potential provider",
			currentUser: users[1],
			post:        posts[1],
			ppUserID:    users[3].ID,
			wantErr: fmt.Sprintf(`unable to find PotentialProvider in order to delete it: `+
				`user %v has insufficient permissions to access PotentialProvider %v`,
				users[1].ID, pps[4].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.DestroyWithPostUUIDAndUserID(test.post.UUID.String(), test.ppUserID, test.currentUser)

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

func (ms *ModelSuite) TestDestroyWithPostUUIDAndUserUUID() {
	f := createProvidersFixtures(ms)
	posts := f.Posts
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		post        Post
		ppUserUUID  uuid.UUID
		wantIDs     []int
		wantErr     string
	}{
		{
			name:        "good: post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			ppUserUUID:  users[1].UUID, // first PotentialProvider gets deleted
			wantIDs:     []int{pps[1].ID, pps[2].ID, pps[3].ID, pps[4].ID},
		},
		{
			name:        "good: potential provider as current user",
			currentUser: users[2],
			post:        posts[0],
			ppUserUUID:  users[2].UUID, // second PotentialProvider also gets deleted
			wantIDs:     []int{pps[2].ID, pps[3].ID, pps[4].ID},
		},
		{
			name:        "bad: current user is not potential provider",
			currentUser: users[1],
			post:        posts[1],
			ppUserUUID:  users[3].UUID,
			wantErr: fmt.Sprintf(`unable to find PotentialProvider in order to delete it: `+
				`user %v has insufficient permissions to access PotentialProvider %v`, users[1].ID, pps[4].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := PotentialProvider{}
			err := provider.DestroyWithPostUUIDAndUserUUID(
				test.post.UUID.String(), test.ppUserUUID.String(), test.currentUser)

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
