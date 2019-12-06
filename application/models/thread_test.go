package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/wecarry-api/domain"
)

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
				UUID:   domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "missing post_id",
			thread: Thread{
				UUID: domain.GetUUID(),
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

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	tests := []struct {
		name    string
		uuid    string
		want    Thread
		wantErr bool
	}{
		{name: "good", uuid: threadFixtures.Threads[0].UUID.String(), want: threadFixtures.Threads[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUUID().String(), wantErr: true},
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
				} else if thread.UUID != test.want.UUID {
					t.Errorf("FindByUUID() got = %s, want %s", thread.UUID, test.want.UUID)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_GetPost() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	tests := []struct {
		name    string
		thread  Thread
		want    Post
		wantErr bool
	}{
		{
			name:   "good",
			thread: threadFixtures.Threads[0],
			want:   posts[0],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.thread.GetPost()
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("GetPost() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GetPost() error = %v", err)
				} else if got.UUID != test.want.UUID {
					t.Errorf("GetPost() got = %s, want %s", got.UUID, test.want.UUID)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_GetMessages() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	tests := []struct {
		name    string
		thread  Thread
		want    []uuid.UUID
		wantErr bool
	}{
		{
			name:   "one message",
			thread: threadFixtures.Threads[0],
			want: []uuid.UUID{
				threadFixtures.Messages[0].UUID,
			},
		},
		{
			name:   "two messages",
			thread: threadFixtures.Threads[1],
			want: []uuid.UUID{
				threadFixtures.Messages[1].UUID,
				threadFixtures.Messages[2].UUID,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.thread.GetMessages()
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("GetMessages() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GetMessages() error = %v", err)
				} else {
					ids := make([]uuid.UUID, len(got))
					for i := range got {
						ids[i] = got[i].UUID
					}
					if !reflect.DeepEqual(ids, test.want) {
						t.Errorf("GetMessages() got = %s, want %s", ids, test.want)
					}
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_GetParticipants() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	tests := []struct {
		name    string
		thread  Thread
		want    []uuid.UUID
		wantErr bool
	}{
		{
			name:   "one participant",
			thread: threadFixtures.Threads[0],
			want: []uuid.UUID{
				users[0].UUID,
			},
		},
		{
			name:   "two participants",
			thread: threadFixtures.Threads[1],
			want: []uuid.UUID{
				users[1].UUID,
				users[0].UUID,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.thread.GetParticipants()
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("GetParticipants() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GetParticipants() error = %v", err)
				} else {
					ids := make([]uuid.UUID, len(got))
					for i := range got {
						ids[i] = got[i].UUID
					}
					if !reflect.DeepEqual(ids, test.want) {
						t.Errorf("GetParticipants() got = %s, want %s", ids, test.want)
					}
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_CreateWithParticipants() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	post := posts[0]

	var thread Thread
	if err := thread.CreateWithParticipants(post.UUID.String(), users[1]); err != nil {
		t.Errorf("TestThread_CreateWithParticipants() error = %v", err)
		t.FailNow()
	}

	var threadFromDB Thread
	if err := ms.DB.Eager().Find(&threadFromDB, thread.ID); err != nil {
		t.Errorf("TestThread_CreateWithParticipants() couldn't find new thread: %s", err)
	}

	if threadFromDB.PostID != post.ID {
		t.Errorf("TestThread_CreateWithParticipants() post ID is wrong, got %d, expected %d",
			threadFromDB.PostID, post.ID)
	}

	participants, err := threadFromDB.GetParticipants()

	ids := make([]uuid.UUID, len(participants))
	for i := range threadFromDB.Participants {
		ids[i] = threadFromDB.Participants[i].UUID
	}

	ms.Contains(ids, users[0].UUID, "new thread doesn't include post creator as participant")
	ms.Contains(ids, users[1].UUID, "new thread doesn't include provided user as participant")
	ms.Equal(2, len(ids), "incorrect number of participants found")

	var tp ThreadParticipants
	n, err := ms.DB.Where("thread_id = ?", thread.ID).Count(&tp)
	if err != nil {
		t.Errorf("TestThread_CreateWithParticipants() couldn't read from thread_participants: %s", err)
	}
	ms.Equal(2, n, "incorrect number of thread_participants records created")
}

func (ms *ModelSuite) TestThread_ensureParticipants() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	post := posts[0]

	thread := Thread{
		PostID: post.ID,
		UUID:   domain.GetUUID(),
	}

	err := DB.Save(&thread)
	ms.NoError(err, "TestThread_ensureParticipants() error saving new thread for test")

	tests := []struct {
		name   string
		userID int
		want   []uuid.UUID
	}{
		{
			name:   "just creator",
			userID: users[0].ID,
			want:   []uuid.UUID{users[0].UUID},
		},
		{
			name:   "add provider",
			userID: users[1].ID,
			want:   []uuid.UUID{users[0].UUID, users[1].UUID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := thread.ensureParticipants(post, test.userID)
			ms.NoError(err)

			participants, err := thread.GetParticipants()
			ms.NoError(err, "can't get thread participants from thread")

			ids := make([]uuid.UUID, len(participants))
			for i := range participants {
				ids[i] = participants[i].UUID
			}

			ms.Equal(len(test.want), len(ids), "incorrect number of participants found")

			ms.Contains(ids, test.want[0], "new thread doesn't include post creator as participant")

			if len(test.want) == 2 {
				ms.Contains(ids, users[1].UUID, "new thread doesn't include provided user as participant")
			}

		})
	}
}

func (ms *ModelSuite) TestThread_GetLastViewedAt() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	tests := []struct {
		name    string
		thread  Thread
		user    User
		want    time.Time
		wantErr bool
	}{
		{
			name:   "good",
			thread: threadFixtures.Threads[0],
			user:   users[0],
			want:   threadFixtures.ThreadParticipants[0].LastViewedAt,
		},
		{
			name:    "invalid user",
			thread:  threadFixtures.Threads[0],
			user:    User{},
			wantErr: true,
		},
		{
			name:    "invalid thread",
			thread:  Thread{},
			user:    users[0],
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lastViewedAt, err := test.thread.GetLastViewedAt(test.user)
			if test.wantErr {
				ms.Error(err, "did not get expected error")
				return
			}

			ms.NoError(err)
			ms.WithinDuration(test.want, *lastViewedAt, time.Duration(time.Second))
		})
	}
}

func (ms *ModelSuite) TestThread_UpdateLastViewedAt() {
	t := ms.T()
	f := CreateFixtures_ThreadParticipant_UpdateLastViewedAt(ms, t)

	tests := []struct {
		name         string
		thread       Thread
		user         User
		lastViewedAt time.Time
		wantErr      string
	}{
		{name: "good", thread: f.Threads[0], user: f.Users[0], lastViewedAt: time.Now()},
		{name: "wrong user", thread: f.Threads[0], user: f.Users[1], wantErr: "failed to find thread_participant"},
		// other combinations are tested in TestThreadParticipant_UpdateLastViewedAt and
		// TestThreadParticipant_FindByThreadIDAndUserID
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.thread.UpdateLastViewedAt(test.user.ID, test.lastViewedAt)

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error")
				return
			}
			ms.NoError(err)

			lastViewedAt, err := test.thread.GetLastViewedAt(test.user)
			ms.NoError(err)
			ms.WithinDuration(test.lastViewedAt, *lastViewedAt, time.Second,
				fmt.Sprintf("time not correct, got %v, wanted %v", lastViewedAt, test.lastViewedAt))
		})
	}
}

func (ms *ModelSuite) TestThread_UnreadMessageCount() {
	t := ms.T()

	f := CreateThreadFixtures_UnreadMessageCount(ms, t)

	tests := []struct {
		name    string
		threadP ThreadParticipant
		user    User
		want    int
		wantErr bool
	}{
		{
			name:    "Eager User Own Thread",
			threadP: f.ThreadParticipants[0],
			user:    f.Users[0],
			want:    0,
		},
		{
			name:    "Eager User Other Thread",
			threadP: f.ThreadParticipants[3],
			user:    f.Users[0],
			want:    0,
		},
		{
			name:    "Lazy User Other Thread",
			threadP: f.ThreadParticipants[1],
			user:    f.Users[1],
			want:    1,
		},
		{
			name:    "Lazy User Own Thread",
			threadP: f.ThreadParticipants[2],
			user:    f.Users[1],
			want:    2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := DB.Load(&test.threadP)
			ms.NoError(err)

			got, err := test.threadP.Thread.UnreadMessageCount(test.user.ID, test.threadP.LastViewedAt)
			if test.wantErr {
				ms.Error(err, "did not get expected error")
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, got)
		})
	}
}
