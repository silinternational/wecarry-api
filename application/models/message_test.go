package models

import (
	"testing"

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

	userResults, err := messages[0].GetSender([]string{"id", "nickname", "email"})

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

	threadResults, err := messages[0].GetThread([]string{"id", "uuid", "post_id"})

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(threads[0].ID, threadResults.ID, "Bad thread ID")
	ms.Equal(threads[0].Uuid, threadResults.Uuid, "Bad thread UUID")
	ms.Equal(threads[0].PostID, threadResults.PostID, "Bad thread PostID")
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
