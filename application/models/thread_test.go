package models

import (
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

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

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

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	tests := []struct {
		name           string
		postID, userID int
		want           Thread
		wantErr        bool
	}{
		{name: "good", postID: posts[0].ID, userID: users[0].ID, want: threadFixtures.Threads[0]},
		{name: "wrong post ID", postID: posts[1].ID, userID: users[0].ID, want: Thread{}},
		{name: "wrong user ID", postID: posts[0].ID, userID: users[2].ID, want: Thread{}},
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

func (ms *ModelSuite) TestThread_GetPost() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])

	type args struct {
		thread       Thread
		selectFields []string
	}
	tests := []struct {
		name    string
		args    args
		want    Post
		wantErr bool
	}{
		{
			name: "good",
			args: args{
				thread:       threadFixtures.Threads[0],
				selectFields: []string{"uuid"},
			},
			want: posts[0],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.args.thread.GetPost(test.args.selectFields)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("GetPost() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GetPost() error = %v", err)
				} else if got.Uuid != test.want.Uuid {
					t.Errorf("GetPost() got = %s, want %s", got.Uuid, test.want.Uuid)
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

	type args struct {
		thread       Thread
		selectFields []string
	}
	tests := []struct {
		name    string
		args    args
		want    []uuid.UUID
		wantErr bool
	}{
		{
			name: "one message",
			args: args{
				thread:       threadFixtures.Threads[0],
				selectFields: []string{},
			},
			want: []uuid.UUID{
				threadFixtures.Messages[0].Uuid,
			},
		},
		{
			name: "two messages",
			args: args{
				thread:       threadFixtures.Threads[1],
				selectFields: []string{},
			},
			want: []uuid.UUID{
				threadFixtures.Messages[1].Uuid,
				threadFixtures.Messages[2].Uuid,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.args.thread.GetMessages(test.args.selectFields)
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
						ids[i] = got[i].Uuid
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

	type args struct {
		thread       Thread
		selectFields []string
	}
	tests := []struct {
		name    string
		args    args
		want    []uuid.UUID
		wantErr bool
	}{
		{
			name: "one participant",
			args: args{
				thread:       threadFixtures.Threads[0],
				selectFields: []string{},
			},
			want: []uuid.UUID{
				users[0].Uuid,
			},
		},
		{
			name: "two participants",
			args: args{
				thread:       threadFixtures.Threads[1],
				selectFields: []string{},
			},
			want: []uuid.UUID{
				users[1].Uuid,
				users[0].Uuid,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.args.thread.GetParticipants(test.args.selectFields)
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
						ids[i] = got[i].Uuid
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
	if err := thread.CreateWithParticipants(post.Uuid.String(), users[1]); err != nil {
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

	ids := make([]uuid.UUID, len(threadFromDB.Participants))
	for i := range threadFromDB.Participants {
		ids[i] = threadFromDB.Participants[i].Uuid
	}

	ms.Contains(ids, users[0].Uuid, "new thread doesn't include post creator as participant")
	ms.Contains(ids, users[1].Uuid, "new thread doesn't include provided user as participant")
	ms.Equal(2, len(ids), "incorrect number of participants found")

	var tp ThreadParticipants
	n, err := ms.DB.Where("thread_id = ?", thread.ID).Count(&tp)
	if err != nil {
		t.Errorf("TestThread_CreateWithParticipants() couldn't read from thread_participants: %s", err)
	}
	ms.Equal(2, n, "incorrect number of thread_participants records created")
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
			ms.Equal(test.want.Format(time.RFC3339), lastViewedAt.Format(time.RFC3339))
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

			got, err := test.threadP.Thread.UnreadMessageCount(&test.user, test.threadP.LastViewedAt)
			if test.wantErr {
				ms.Error(err, "did not get expected error")
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, got)
		})
	}
}
