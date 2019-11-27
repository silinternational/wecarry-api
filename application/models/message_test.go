package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/validate"
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
				Uuid:     domain.GetUuid(),
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
				Uuid:     domain.GetUuid(),
				SentByID: 1,
				Content:  "foo",
			},
			wantErr:  true,
			errField: "thread_id",
		},
		{
			name: "missing sent_by_id",
			message: Message{
				Uuid:     domain.GetUuid(),
				ThreadID: 1,
				Content:  "foo",
			},
			wantErr:  true,
			errField: "sent_by_id",
		},
		{
			name: "missing content",
			message: Message{
				Uuid:     domain.GetUuid(),
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

	userResults, err := messages[0].GetSender()

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

	threadResults, err := messages[0].GetThread()

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(threads[0].ID, threadResults.ID, "Bad thread ID")
	ms.Equal(threads[0].Uuid, threadResults.Uuid, "Bad thread UUID")
	ms.Equal(threads[0].PostID, threadResults.PostID, "Bad thread PostID")
}

func (ms *ModelSuite) TestMessage_Create() {
	t := ms.T()

	f := Fixtures_Message_Create(ms, t)
	msg := Message{
		Uuid:     domain.GetUuid(),
		ThreadID: f.Threads[0].ID,
		SentByID: f.Users[0].ID,
		Content:  `Owe nothing to anyone, except to love one another.`,
	}

	tests := []struct {
		name    string
		msg     Message
		wantErr bool
	}{
		{name: "good", msg: msg},
		{name: "validation error", msg: Message{}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := test.msg
			err := message.Create()

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.msg.Uuid, message.Uuid, "incorrect message UUID")
				_ = ms.DB.Reload(&f.Threads[0])
				ms.WithinDuration(time.Now(), f.Threads[0].UpdatedAt, time.Second,
					"thread.updated_at was not set to the current time")
			}
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
		{name: "good with no extra fields",
			id:          f.Messages[0].ID,
			wantMessage: f.Messages[0],
		},
		{name: "good with two extra fields",
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
			err := message.FindByID(test.id, test.eagerFields...)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantMessage.ID, message.ID, "bad message id")
				ms.Equal(test.wantSentBy.ID, message.SentBy.ID, "bad message sent_by id")
				ms.Equal(test.wantThread.Uuid, message.Thread.Uuid, "bad message thread id")
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
		{name: "good",
			uuid:          f.Messages[0].Uuid.String(),
			wantID:        f.Messages[0].ID,
			wantContent:   f.Messages[0].Content,
			wantCreatedAt: f.Messages[0].CreatedAt,
		},
		{name: "empty ID", uuid: "", wantErr: "error: message uuid must not be blank"},
		{name: "wrong id", uuid: domain.GetUuid().String(), wantErr: "sql: no rows in result set"},
		{name: "invalid UUID", uuid: "40FE092C-8FF1-45BE-BCD4-65AD66C1D0DX", wantErr: "pq: invalid input syntax"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var message Message
			err := message.findByUUID(test.uuid)

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
	wrongUUID := domain.GetUuid().String()

	tests := []struct {
		name    string
		uuid    *string
		user    User
		message Message
		wantErr string
	}{
		{name: "empty ID", uuid: &emptyString, wantErr: "error: message uuid must not be blank"},
		{name: "wrong id", uuid: &wrongUUID, wantErr: "sql: no rows in result set"},
		{name: "invalid UUID", uuid: &badUUID, wantErr: "pq: invalid input syntax"},
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
				testUUID = test.message.Uuid.String()
			} else {
				testUUID = *test.uuid
			}
			err := message.FindByUserAndUUID(test.user, testUUID)

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

	eagerThreadP := f.ThreadParticipants[1] // Lazy User's side of Eager User's post thread
	lazyThreadP := f.ThreadParticipants[2]  // Lazy User's own post thread

	eagerThreadPLVA := eagerThreadP.LastViewedAt

	newMessage := Message{
		Uuid:     domain.GetUuid(),
		ThreadID: lazyThreadP.ThreadID,
		SentByID: lazyThreadP.UserID,
		Content:  "This message should update LastViewedAt",
	}

	err := DB.Create(&newMessage)
	ms.NoError(err)

	tSecond := time.Duration(time.Second)

	gotTP := ThreadParticipant{}
	err = DB.Find(&gotTP, eagerThreadP.ID)
	ms.NoError(err)
	ms.WithinDuration(eagerThreadPLVA, gotTP.LastViewedAt, tSecond)

	gotTP = ThreadParticipant{}
	err = DB.Find(&gotTP, lazyThreadP.ID)
	ms.NoError(err)
	ms.WithinDuration(time.Now(), gotTP.LastViewedAt, tSecond)
}
