package job

import (
	"bytes"
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/gobuffalo/suite/v4"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

type JobSuite struct {
	*suite.Model
}

func Test_JobSuite(t *testing.T) {
	model := suite.NewModel()

	ms := &JobSuite{
		Model: model,
	}
	suite.Run(t, ms)
}

func (js *JobSuite) TestOutdatedRequestsHandler() {
	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	f := CreateFixtures_TestOutdatedRequestHandler(js)
	notifications.TestEmailService.DeleteSentMessages()

	err := outdatedRequestsHandler(nil)
	js.NoError(err)
	js.Equal(1, notifications.TestEmailService.GetNumberOfMessagesSent())

	body := notifications.TestEmailService.GetLastBody()
	js.Contains(body, `We see that your request hasn't been fulfilled yet`)
	js.Contains(body, f.Requests[1].Title)
}

func (js *JobSuite) TestNewThreadMessageHandler() {
	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	f := CreateFixtures_TestNewThreadMessageHandler(js)

	tests := []struct {
		message            models.Message
		recipientID        int
		wantNumberOfEmails int
		wantErr            bool
	}{
		{
			message:            f.Messages[0],
			recipientID:        f.Users[1].ID,
			wantNumberOfEmails: 1,
		},
		{
			message:            f.Messages[1],
			wantNumberOfEmails: 0,
		},
		{
			message:            f.Messages[2],
			wantNumberOfEmails: 0,
		},
		{
			message:            f.Messages[3],
			recipientID:        f.Users[4].ID,
			wantNumberOfEmails: 1,
		},
		{
			message:            f.Messages[4],
			wantNumberOfEmails: 0,
		},
		{
			message:            f.Messages[5],
			wantNumberOfEmails: 0,
		},
	}
	for _, test := range tests {
		js.T().Run(test.message.Content, func(t *testing.T) {
			notifications.TestEmailService.DeleteSentMessages()

			args := map[string]interface{}{
				domain.ArgMessageID: test.message.ID,
			}
			err := newThreadMessageHandler(args)

			if test.wantErr {
				js.Error(err)
				return
			}

			js.NoError(err)
			js.Equal(test.wantNumberOfEmails, notifications.TestEmailService.GetNumberOfMessagesSent())

			if test.wantNumberOfEmails == 1 {
				var tp models.ThreadParticipant
				_ = tp.FindByThreadIDAndUserID(js.DB, test.message.ThreadID, test.recipientID)
				expect := time.Now()
				js.WithinDuration(expect, tp.LastNotifiedAt, time.Second,
					"last notified time not correct, got %v, wanted %v", tp.LastNotifiedAt, expect)

				body := notifications.TestEmailService.GetLastBody()
				js.Contains(body, template.HTMLEscapeString(test.message.Content))
			}
		})
	}
}
