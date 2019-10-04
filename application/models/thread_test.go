package models

import (
	"reflect"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/wecarry-api/domain"
)

type ThreadFixtures struct {
	Threads  Threads
	Messages Messages
}

func CreateThreadFixtures(ms *ModelSuite, t *testing.T, post Post) ThreadFixtures {
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
		{
			Uuid:   domain.GetUuid(),
			PostID: post.ID,
		},
	}
	for i := range threads {
		if err := ms.DB.Create(&threads[i]); err != nil {
			t.Errorf("could not create test threads ... %v", err)
			t.FailNow()
		}
	}

	// Load Thread Participants test fixtures
	threadParticipants := []ThreadParticipant{
		{
			ThreadID: threads[0].ID,
			UserID:   post.CreatedByID,
		},
		{
			ThreadID: threads[1].ID,
			UserID:   post.ProviderID.Int,
		},
		{
			ThreadID: threads[1].ID,
			UserID:   post.CreatedByID,
		},
	}
	for i := range threadParticipants {
		if err := ms.DB.Create(&threadParticipants[i]); err != nil {
			t.Errorf("could not create test thread participants ... %v", err)
			t.FailNow()
		}
	}

	// Load Message test fixtures
	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: post.CreatedByID,
			Content:  "I can being chocolate if you bring PB",
		},
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[1].ID,
			SentByID: post.ProviderID.Int,
			Content:  "I can being PB if you bring chocolate",
		},
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[1].ID,
			SentByID: post.CreatedByID,
			Content:  "Great!",
		},
	}

	for _, message := range messages {
		if err := ms.DB.Create(&message); err != nil {
			t.Errorf("could not create test message ... %v", err)
			t.FailNow()
		}
	}

	return ThreadFixtures{Threads: threads, Messages: messages}
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
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])

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
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])

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
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])

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
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])

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
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, t, posts[0])

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
	ResetTables(t, ms.DB)

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	post := posts[0]

	var thread Thread
	if err := thread.CreateWithParticipants(post.Uuid.String(), users[1]); err != nil {
		t.Errorf("TestThread_CreateWithParticipants() error = %v", err)
		t.FailNow()
	}

	var threadFromDB Thread
	if err := DB.Eager().Find(&threadFromDB, thread.ID); err != nil {
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
	n, err := DB.Where("thread_id = ?", thread.ID).Count(&tp)
	if err != nil {
		t.Errorf("TestThread_CreateWithParticipants() couldn't read from thread_participants: %s", err)
	}
	ms.Equal(2, n, "incorrect number of thread_participants records created")
}
