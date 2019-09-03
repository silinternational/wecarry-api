package models

import (
	"testing"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/handcarry-api/domain"
)

type ThreadFixtures struct {
	Threads Threads
}

func CreateThreadFixtures(t *testing.T, post Post) ThreadFixtures {
	// Load Thread test fixtures
	threads := []Thread{
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
	}
	for i := range threads {
		if err := DB.Create(&threads[i]); err != nil {
			t.Errorf("could not create test threads ... %v", err)
			t.FailNow()
		}
	}

	// Load Thread test fixtures
	threadParticipants := []ThreadParticipant{
		{
			ThreadID: threads[0].ID,
			UserID:   post.CreatedByID,
		},
		{
			ThreadID: threads[1].ID,
			UserID:   post.CreatedByID,
		},
	}
	for i := range threadParticipants {
		if err := DB.Create(&threadParticipants[i]); err != nil {
			t.Errorf("could not create test thread participants ... %v", err)
			t.FailNow()
		}
	}

	return ThreadFixtures{Threads: threads}
}

func (ms *ModelSuite) TestThread_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		thread   Thread
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			thread: Thread{
				PostID: 1,
				Uuid:   domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing post_id",
			thread: Thread{
				Uuid: domain.GetUuid(),
			},
			wantErr:  true,
			errField: "post_id",
		},
		{
			name: "missing uuid",
			thread: Thread{
				PostID: 1,
			},
			wantErr:  true,
			errField: "uuid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.thread.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestThread_FindByUUID() {
	t := ms.T()
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)
	threadFixtures := CreateThreadFixtures(t, posts[0])

	tests := []struct {
		name    string
		uuid    string
		want    Thread
		wantErr bool
	}{
		{name: "good", uuid: threadFixtures.Threads[0].Uuid.String(), want: threadFixtures.Threads[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUuid().String(), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var thread Thread
			err := thread.FindByUUID(test.uuid)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("FindByUUID() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("FindByUUID() error = %v", err)
				} else if thread.Uuid != test.want.Uuid {
					t.Errorf("FindByUUID() got = %s, want %s", thread.Uuid, test.want.Uuid)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_FindByPostIDAndUserID() {
	t := ms.T()
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)
	threadFixtures := CreateThreadFixtures(t, posts[0])

	tests := []struct {
		name           string
		postID, userID int
		want           Thread
		wantErr        bool
	}{
		{name: "good", postID: posts[0].ID, userID: users[0].ID, want: threadFixtures.Threads[0]},
		{name: "wrong post ID", postID: posts[1].ID, userID: users[0].ID, want: Thread{}},
		{name: "wrong user ID", postID: posts[0].ID, userID: users[1].ID, want: Thread{}},
		{name: "bad post ID", postID: -1, userID: users[0].ID, want: Thread{}},
		{name: "bad user ID", postID: posts[0].ID, userID: -1, want: Thread{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var thread Thread
			err := thread.FindByPostIDAndUserID(test.postID, test.userID)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("FindByPostIDAndUserID() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("FindByPostIDAndUserID() error = %v", err)
				} else if thread.Uuid != test.want.Uuid {
					t.Errorf("FindByPostIDAndUserID() got = %s, want %s", thread.Uuid, test.want.Uuid)
				}
			}
		})
	}
}
