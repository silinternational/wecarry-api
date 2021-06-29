package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/validate/v3"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestMessage_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		message  Message
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			message: Message{
				UUID:     domain.GetUUID(),
				ThreadID: 1,
				SentByID: 1,
				Content:  "foo",
			},
			wantErr: false,
		},
		{
			name: "missing uuid",
			message: Message{
				ThreadID: 1,
				SentByID: 1,
				Content:  "foo",
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing thread_id",
			message: Message{
				UUID:     domain.GetUUID(),
				SentByID: 1,
				Content:  "foo",
			},
			wantErr:  true,
			errField: "thread_id",
		},
		{
			name: "missing sent_by_id",
			message: Message{
				UUID:     domain.GetUUID(),
				ThreadID: 1,
				Content:  "foo",
			},
			wantErr:  true,
			errField: "sent_by_id",
		},
		{
			name: "missing content",
			message: Message{
				UUID:     domain.GetUUID(),
				ThreadID: 1,
				SentByID: 1,
			},
			wantErr:  true,
			errField: "content",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.message.Validate(DB)
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

func (ms *ModelSuite) TestMessage_GetSender() {
	t := ms.T()

	messageFixtures := Fixtures_Message_GetSender(ms, t)

	messages := messageFixtures.Messages
	users := messageFixtures.Users

	userResults, err := messages[0].GetSender(ms.DB)
	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(users[0].ID, userResults.ID, "Bad user ID")
	ms.Equal(users[0].Nickname, userResults.Nickname, "Bad user Nickname")
	ms.Equal(users[0].Email, userResults.Email, "Bad user Email")
}

func (ms *ModelSuite) TestMessage_GetThread() {
	t := ms.T()

	messageFixtures := Fixtures_Message_GetSender(ms, t)

	messages := messageFixtures.Messages
	threads := messageFixtures.Threads

	threadResults, err := messages[0].GetThread(ms.DB)
	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(threads[0].ID, threadResults.ID, "Bad thread ID")
	ms.Equal(threads[0].UUID, threadResults.UUID, "Bad thread UUID")
	ms.Equal(threads[0].RequestID, threadResults.RequestID, "Bad thread RequestID")
}

func (ms *ModelSuite) TestMessage_Create() {
	t := ms.T()

	f := Fixtures_Message_Create(ms, t)
	threadUUID := f.Threads[0].UUID.String()

	tests := []struct {
		name        string
		user        User
		requestUUID string
		threadUUID  *string
		content     string
		wantErr     bool
	}{
		{
			name:        "good, 1st message, visible request",
			user:        f.Users[2],
			requestUUID: f.Requests[1].UUID.String(),
			threadUUID:  nil,
			content:     "Owe nothing to anyone, except to love one another.",
			wantErr:     false,
		},
		{
			name:        "good, already a participant",
			user:        f.Users[0],
			requestUUID: f.Requests[1].UUID.String(),
			threadUUID:  &threadUUID,
			content:     "Hatred stirs up conflict, but love covers over all wrongs.",
			wantErr:     false,
		},
		{
			name:        "bad, not a participant",
			user:        f.Users[1],
			requestUUID: f.Requests[1].UUID.String(),
			threadUUID:  &threadUUID,
			content:     "bad message",
			wantErr:     true,
		},
		{
			name:        "bad, not a visible request",
			user:        f.Users[1],
			requestUUID: f.Requests[0].UUID.String(),
			threadUUID:  nil,
			content:     "another bad message",
			wantErr:     true,
		},
		{
			name:        "bad, thread not valid for request",
			user:        f.Users[0],
			requestUUID: f.Requests[2].UUID.String(),
			threadUUID:  &threadUUID,
			content:     "another bad message",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message Message
			err := message.Create(ms.DB, tt.user, tt.requestUUID, tt.threadUUID, tt.content)

			if tt.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			ms.Greater(message.ID, 0, "new message contains invalid ID")
		})
	}
}

func (ms *ModelSuite) TestMessage_CreateForRest() {
	t := ms.T()

	f := Fixtures_Message_Create(ms, t)
	threadUUID := f.Threads[0].UUID.String()

	tests := []struct {
		name        string
		user        User
		requestUUID string
		threadUUID  *string
		content     string
		wantErr     bool
	}{
		{
			name:        "good, 1st message, visible request",
			user:        f.Users[2],
			requestUUID: f.Requests[1].UUID.String(),
			threadUUID:  nil,
			content:     "Owe nothing to anyone, except to love one another.",
			wantErr:     false,
		},
		{
			name:        "good, already a participant",
			user:        f.Users[0],
			requestUUID: f.Requests[1].UUID.String(),
			threadUUID:  &threadUUID,
			content:     "Hatred stirs up conflict, but love covers over all wrongs.",
			wantErr:     false,
		},
		{
			name:        "bad, not a participant",
			user:        f.Users[1],
			requestUUID: f.Requests[1].UUID.String(),
			threadUUID:  &threadUUID,
			content:     "bad message",
			wantErr:     true,
		},
		{
			name:        "bad, not a visible request",
			user:        f.Users[1],
			requestUUID: f.Requests[0].UUID.String(),
			threadUUID:  nil,
			content:     "another bad message",
			wantErr:     true,
		},
		{
			name:        "bad, thread not valid for request",
			user:        f.Users[0],
			requestUUID: f.Requests[2].UUID.String(),
			threadUUID:  &threadUUID,
			content:     "another bad message",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message Message
			input := api.MessageInput{
				Content:   tt.content,
				RequestID: tt.requestUUID,
				ThreadID:  tt.threadUUID,
			}
			err := message.CreateForRest(ms.DB, tt.user, input)

			if tt.wantErr {
				ms.Error(err)
				return
			}

			ms.Nil(err, "received unexpected error: ")
			ms.Greater(message.ID, 0, "new message contains invalid ID")
		})
	}
}

func (ms *ModelSuite) TestMessage_FindByID() {
	t := ms.T()

	f := Fixtures_Message_FindByID(ms, t)

	tests := []struct {
		name        string
		id          int
		eagerFields []string
		wantMessage Message
		wantSentBy  User
		wantThread  Thread
		wantErr     bool
	}{
		{
			name:        "good with no extra fields",
			id:          f.Messages[0].ID,
			wantMessage: f.Messages[0],
		},
		{
			name:        "good with two extra fields",
			id:          f.Messages[0].ID,
			eagerFields: []string{"SentBy", "Thread"},
			wantMessage: f.Messages[0],
			wantSentBy:  f.Users[0],
			wantThread:  f.Threads[0],
		},
		{name: "zero ID", id: 0, wantErr: true},
		{name: "wrong id", id: 99999, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var message Message
			err := message.FindByID(ms.DB, test.id, test.eagerFields...)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantMessage.ID, message.ID, "bad message id")
				ms.Equal(test.wantSentBy.ID, message.SentBy.ID, "bad message sent_by id")
				ms.Equal(test.wantThread.UUID, message.Thread.UUID, "bad message thread id")
			}
		})
	}
}

func (ms *ModelSuite) TestMessage_FindByUUID() {
	t := ms.T()

	f := Fixtures_Message_Find(ms)

	tests := []struct {
		name          string
		uuid          string
		wantID        int
		wantContent   string
		wantCreatedAt time.Time
		wantErr       string
	}{
		{
			name:          "good",
			uuid:          f.Messages[0].UUID.String(),
			wantID:        f.Messages[0].ID,
			wantContent:   f.Messages[0].Content,
			wantCreatedAt: f.Messages[0].CreatedAt,
		},
		{name: "empty ID", uuid: "", wantErr: "error: message uuid must not be blank"},
		{name: "wrong id", uuid: domain.GetUUID().String(), wantErr: "sql: no rows in result set"},
		{name: "invalid UUID", uuid: "40FE092C-8FF1-45BE-BCD4-65AD66C1D0DX", wantErr: "invalid input syntax"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var message Message
			err := message.findByUUID(ms.DB, test.uuid)

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error, %s", err)
				return
			}
			ms.NoError(err)
			ms.Equal(test.wantID, message.ID, "bad message ID")
			ms.Equal(test.wantContent, message.Content, "bad message Content")
			ms.WithinDuration(test.wantCreatedAt, message.CreatedAt, time.Second, "bad message CreatedAt")
		})
	}
}

func (ms *ModelSuite) TestMessage_FindByUserAndUUID() {
	t := ms.T()

	f := Fixtures_Message_Find(ms)

	emptyString := ""
	badUUID := "40FE092C-8FF1-45BE-BCD4-65AD66C1D0DX"
	wrongUUID := domain.GetUUID().String()

	tests := []struct {
		name    string
		uuid    *string
		user    User
		message Message
		wantErr string
	}{
		{name: "empty ID", uuid: &emptyString, wantErr: "error: message uuid must not be blank"},
		{name: "wrong id", uuid: &wrongUUID, wantErr: "sql: no rows in result set"},
		{name: "invalid UUID", uuid: &badUUID, wantErr: "invalid input syntax"},
		{name: "on thread, user", user: f.Users[1], message: f.Messages[0]},
		{name: "on thread, admin", user: f.Users[2], message: f.Messages[0]},
		{name: "on thread, salesAdmin", user: f.Users[3], message: f.Messages[0]},
		{name: "on thread, superAdmin", user: f.Users[4], message: f.Messages[0]},
		{name: "not on thread, user", user: f.Users[1], message: f.Messages[1], wantErr: "insufficient permissions"},
		{name: "not on thread, admin", user: f.Users[2], message: f.Messages[1], wantErr: "insufficient permissions"},
		{name: "not on thread, salesAdmin", user: f.Users[3], message: f.Messages[1], wantErr: "insufficient permissions"},
		{name: "not on thread, superAdmin", user: f.Users[4], message: f.Messages[1]},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var message Message
			var testUUID string
			if test.uuid == nil {
				testUUID = test.message.UUID.String()
			} else {
				testUUID = *test.uuid
			}
			err := message.FindByUserAndUUID(ms.DB, test.user, testUUID)

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error")
				return
			}
			ms.NoError(err)
			ms.Equal(test.message.ID, message.ID, "bad ID")
			ms.Equal(test.message.Content, message.Content, "bad Content")
		})
	}
}

func (ms *ModelSuite) TestMessage_AfterCreate() {
	f := CreateMessageFixtures_AfterCreate(ms, ms.T())

	eagerThreadP := f.ThreadParticipants[1] // Lazy User's side of Eager User's request thread
	lazyThreadP := f.ThreadParticipants[2]  // Lazy User's own request thread

	eagerThreadPLVA := eagerThreadP.LastViewedAt

	newMessage := Message{
		UUID:     domain.GetUUID(),
		ThreadID: lazyThreadP.ThreadID,
		SentByID: lazyThreadP.UserID,
		Content:  "This message should update LastViewedAt",
	}

	var eventDetected bool
	deleteFn, err := events.Listen(func(e events.Event) {
		if e.Kind == domain.EventApiMessageCreated {
			eventDetected = true
		}
	})
	ms.NoError(err, "error registering event listener")
	defer deleteFn()

	err = DB.Create(&newMessage)
	ms.NoError(err)

	const tSecond = time.Second

	gotTP := ThreadParticipant{}
	err = DB.Find(&gotTP, eagerThreadP.ID)
	ms.NoError(err)
	ms.WithinDuration(eagerThreadPLVA, gotTP.LastViewedAt, tSecond)

	gotTP = ThreadParticipant{}
	err = DB.Find(&gotTP, lazyThreadP.ID)
	ms.NoError(err)
	ms.WithinDuration(time.Now(), gotTP.LastViewedAt, tSecond)

	_ = ms.DB.Reload(&f.Threads[1])
	ms.WithinDuration(time.Now(), f.Threads[1].UpdatedAt, tSecond,
		"thread.updated_at was not set to the current time")

	ms.True(eventDetected, "EventApiMessageCreated event was not emitted")
}
