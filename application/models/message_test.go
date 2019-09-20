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

	ResetTables(t, ms.DB)
	messageFixtures := Fixtures_GetSender(ms, t)

	messages := messageFixtures.Messages
	users := messageFixtures.Users

	userResults, err := messages[1].GetSender([]string{"id", "nickname", "last_name", "first_name", "email"})

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(users[1].ID, userResults.ID, "Bad user ID")
	ms.Equal(users[1].Nickname, userResults.Nickname, "Bad user Nickname")
	ms.Equal(users[1].LastName, userResults.LastName, "Bad user LastName")
	ms.Equal(users[1].FirstName, userResults.FirstName, "Bad user FirstName")
	ms.Equal(users[1].Email, userResults.Email, "Bad user Email")
}

func (ms *ModelSuite) TestMessage_GetThread() {
	t := ms.T()

	ResetTables(t, ms.DB)
	messageFixtures := Fixtures_GetSender(ms, t)

	messages := messageFixtures.Messages
	threads := messageFixtures.Threads

	threadResults, err := messages[1].GetThread([]string{"id", "uuid", "post_id"})

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(threads[1].ID, threadResults.ID, "Bad thread ID")
	ms.Equal(threads[1].Uuid, threadResults.Uuid, "Bad thread UUID")
	ms.Equal(threads[1].PostID, threadResults.PostID, "Bad thread PostID")
}
