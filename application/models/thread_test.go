package models

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
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
				RequestID: 1,
				UUID:      domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "missing request_id",
			thread: Thread{
				UUID: domain.GetUUID(),
			},
			wantErr:  true,
			errField: "request_id",
		},
		{
			name: "missing uuid",
			thread: Thread{
				RequestID: 1,
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

	_ = createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 1, false)
	threadFixtures := CreateThreadFixtures(ms, requests[0])

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
			err := thread.FindByUUID(ms.DB, test.uuid)
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

func (ms *ModelSuite) TestThread_GetRequest() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 1, false)
	threadFixtures := CreateThreadFixtures(ms, requests[0])

	tests := []struct {
		name    string
		thread  Thread
		want    Request
		wantErr bool
	}{
		{
			name:   "good",
			thread: threadFixtures.Threads[0],
			want:   requests[0],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.thread.LoadRequest(ms.DB)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("LoadRequest() did not return expected error")
				}
			} else {
				got := test.thread.Request
				if err != nil {
					t.Errorf("LoadRequest() error = %v", err)
				} else if got.UUID != test.want.UUID {
					t.Errorf("LoadRequest() got = %s, want %s", got.UUID, test.want.UUID)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_GetMessages() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)
	requests := createRequestFixtures(ms.DB, 1, false)
	threadFixtures := CreateThreadFixtures(ms, requests[0])

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
			err := test.thread.LoadMessages(ms.DB)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("Messages() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("Messages() error = %v", err)
				} else {
					got := test.thread.Messages
					ids := make([]uuid.UUID, len(got))
					for i := range got {
						ids[i] = got[i].UUID
					}
					if !reflect.DeepEqual(ids, test.want) {
						t.Errorf("Messages() got = %s, want %s", ids, test.want)
					}
				}
			}
		})
	}
}

func (ms *ModelSuite) TestThread_LoadParticipants() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	requests := createRequestFixtures(ms.DB, 1, false)
	threadFixtures := CreateThreadFixtures(ms, requests[0])

	tests := []struct {
		name    string
		thread  Thread
		want    []int
		wantErr bool
	}{
		{
			name:   "one participant",
			thread: threadFixtures.Threads[0],
			want: []int{
				users[0].ID,
			},
		},
		{
			name:   "two participants",
			thread: threadFixtures.Threads[1],
			want: []int{
				threadFixtures.Users[0].ID,
				users[0].ID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.thread.LoadParticipants(ms.DB)
			if tt.wantErr {
				ms.Error(err)
				return
			}
			ms.NoError(err)
			got := tt.thread.Participants

			ids := make([]int, len(got))
			for i := range got {
				ids[i] = got[i].ID
			}
			sort.Ints(tt.want)
			ms.EqualValues(tt.want, ids)

		})
	}
}

func (ms *ModelSuite) TestThread_CreateWithParticipants() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	requests := createRequestFixtures(ms.DB, 1, false)
	request := requests[0]

	var thread Thread
	if err := thread.CreateWithParticipants(ms.DB, request, users[1]); err != nil {
		t.Errorf("TestThread_CreateWithParticipants() error = %v", err)
		t.FailNow()
	}

	var threadFromDB Thread
	if err := ms.DB.Eager().Find(&threadFromDB, thread.ID); err != nil {
		t.Errorf("TestThread_CreateWithParticipants() couldn't find new thread: %s", err)
	}

	if threadFromDB.RequestID != request.ID {
		t.Errorf("TestThread_CreateWithParticipants() request ID is wrong, got %d, expected %d",
			threadFromDB.RequestID, request.ID)
	}

	err := threadFromDB.LoadParticipants(ms.DB)
	ms.NoError(err)

	ids := make([]uuid.UUID, len(threadFromDB.Participants))
	for i := range threadFromDB.Participants {
		ids[i] = threadFromDB.Participants[i].UUID
	}

	ms.Contains(ids, users[0].UUID, "new thread doesn't include request creator as participant")
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

	users := createUserFixtures(ms.DB, 2).Users
	requests := createRequestFixtures(ms.DB, 1, false)
	request := requests[0]

	thread := Thread{
		RequestID: request.ID,
		UUID:      domain.GetUUID(),
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
			err := thread.ensureParticipants(ms.DB, request, test.userID)
			ms.NoError(err)

			err = thread.LoadParticipants(ms.DB)
			ms.NoError(err, "can't get thread participants from thread")

			ids := make([]uuid.UUID, len(thread.Participants))
			for i := range ids {
				ids[i] = thread.Participants[i].UUID
			}

			ms.Equal(len(test.want), len(ids), "incorrect number of participants found")

			ms.Contains(ids, test.want[0], "new thread doesn't include request creator as participant")

			if len(test.want) == 2 {
				ms.Contains(ids, users[1].UUID, "new thread doesn't include provided user as participant")
			}
		})
	}
}

func (ms *ModelSuite) TestThread_GetLastViewedAt() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	requests := createRequestFixtures(ms.DB, 1, false)
	threadFixtures := CreateThreadFixtures(ms, requests[0])

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
			lastViewedAt, err := test.thread.GetLastViewedAt(ms.DB, test.user)
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
			err := test.thread.UpdateLastViewedAt(ms.DB, test.user.ID, test.lastViewedAt)

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error")
				return
			}
			ms.NoError(err)

			lastViewedAt, err := test.thread.GetLastViewedAt(ms.DB, test.user)
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

			got, err := test.threadP.Thread.GetUnreadMessageCount(ms.DB, test.user.ID, test.threadP.LastViewedAt)
			if test.wantErr {
				ms.Error(err, "did not get expected error")
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, got)
		})
	}
}

func (ms *ModelSuite) TestThread_IsVisible() {
	t := ms.T()

	request := createRequestFixtures(ms.DB, 1, false)[0]
	f := CreateThreadFixtures(ms, request)

	tests := []struct {
		name   string
		thread Thread
		user   User
		want   bool
	}{
		{
			name:   "bad thread",
			thread: Thread{},
			user:   f.Users[0],
			want:   false,
		},
		{
			name:   "bad user",
			thread: f.Threads[1],
			user:   User{},
			want:   false,
		},
		{
			name:   "no",
			thread: f.Threads[0],
			user:   f.Users[0],
			want:   false,
		},
		{
			name:   "yes",
			thread: f.Threads[1],
			user:   f.Users[0],
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.thread.IsVisible(ms.DB, tt.user.ID)
			ms.Equal(tt.want, got)
		})
	}
}
