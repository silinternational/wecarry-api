package listeners

import (
	"bytes"
	"fmt"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
	"os"
	"testing"
)

func (ms *ModelSuite) TestGetPostUsers() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_GetPostRecipients(ms, t)
	users := orgUserPostFixtures.users
	posts := orgUserPostFixtures.posts

	tests := []struct {
		name          string
		id            int
		wantRequester PostUser
		wantProvider  PostUser
		wantErr       bool
	}{
		{name: "Request by User0 with User1 as Provider",
			id: posts[0].ID,
			wantRequester: PostUser{
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
			wantProvider: PostUser{
				Nickname: users[1].Nickname,
				Email:    users[1].Email,
			},
		},
		{name: "Request by User0 with no Provider",
			id: posts[1].ID,
			wantRequester: PostUser{
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var post models.Post
			err := post.FindByID(test.id)
			ms.NoError(err, "error finding post for test")

			postUsers := GetPostUsers(post)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantRequester, postUsers.Requester)
				ms.Equal(test.wantProvider, postUsers.Provider)
			}
		})
	}
}

func (ms *ModelSuite) TestRequestStatusUpdatedNotifications() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_RequestStatusUpdatedNotifications(ms, t)
	posts := orgUserPostFixtures.posts

	postStatusEData := models.PostStatusEventData{
		OldStatus: models.PostStatusOpen,
		NewStatus: models.PostStatusCommitted,
		PostID:    posts[0].ID,
	}

	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	// No logging message expected
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got := buf.String()
	ms.Equal("", got, "Got an unexpected error log entry")

	buf.Reset()

	// Logging message expected about bad transition
	postStatusEData.NewStatus = models.PostStatusAccepted
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got = buf.String()
	ms.Contains(got, "unexpected status transition 'OPEN-ACCEPTED'", "Got an unexpected error log entry")

}

func (ms *ModelSuite) TestSendNotificationRequestFromStatus() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_sendNotificationRequestFromStatus(ms, t)
	posts := orgUserPostFixtures.posts

	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	tests := []struct {
		name           string
		post           models.Post
		template       string
		sendFunction   func(string, models.Post)
		wantEmailsSent int
		wantToEmail    string
		wantErrLog     string
	}{
		{name: "Good - Open to Committed",
			post:           posts[0],
			template:       domain.MessageTemplateRequestFromOpenToCommitted,
			sendFunction:   sendNotificationRequestFromOpenToCommitted,
			wantEmailsSent: 1,
			wantToEmail:    posts[0].CreatedBy.Email,
		},
		{name: "Bad - Open to Committed",
			template:     domain.MessageTemplateRequestFromOpenToCommitted,
			sendFunction: sendNotificationRequestFromOpenToCommitted,
			post:         posts[1],
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromOpenToCommitted),
		},
		{name: "Good - Committed to Accepted",
			post:           posts[0],
			template:       domain.MessageTemplateRequestFromCommittedToAccepted,
			sendFunction:   sendNotificationRequestFromCommittedToAccepted,
			wantEmailsSent: 1,
			wantToEmail:    posts[0].Provider.Email,
		},
		{name: "Bad - Committed to Accepted",
			template:     domain.MessageTemplateRequestFromCommittedToAccepted,
			sendFunction: sendNotificationRequestFromCommittedToAccepted,
			post:         posts[1],
			wantErrLog: fmt.Sprintf("error preparing '%s' notification - no provider\n",
				domain.MessageTemplateRequestFromCommittedToAccepted),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			notifications.TestEmailService.DeleteSentMessages()

			test.sendFunction(test.template, test.post)
			gotBuf := buf.String()
			buf.Reset()

			emailCount := notifications.TestEmailService.GetNumberOfMessagesSent()
			lastToEmail := notifications.TestEmailService.GetLastToEmail()
			ms.Equal(test.wantEmailsSent, emailCount, "bad To Email")
			ms.Equal(test.wantToEmail, lastToEmail, "bad To Email")
			ms.Equal(test.wantErrLog, gotBuf, "wrong error log entry")
		})
	}
}
