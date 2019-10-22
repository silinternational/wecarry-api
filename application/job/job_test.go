package job

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/silinternational/wecarry-api/models"

	"github.com/silinternational/wecarry-api/notifications"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/suite"
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

func (js *JobSuite) TestNewMessageHandler() {
	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	f := CreateFixtures_TestNewMessageHandler(js)

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
			notifications.DummyEmailService.DeleteSentMessages()

			args := map[string]interface{}{
				domain.ArgMessageID: test.message.ID,
			}
			err := NewMessageHandler(args)

			if test.wantErr {
				js.Error(err)
			} else {
				js.NoError(err)
				js.Equal(test.wantNumberOfEmails, notifications.DummyEmailService.GetNumberOfMessagesSent())

				if test.wantNumberOfEmails == 1 {
					var tp models.ThreadParticipant
					_ = tp.FindByThreadIDAndUserID(test.message.ThreadID, test.recipientID)
					js.True(isNow(tp.LastNotifiedAt))
				}
			}
		})
	}
}

// isNow loosely compares the given time to time.Now()
func isNow(t time.Time) bool {
	if t.After(time.Now().Add(5 * time.Second)) {
		return false
	}
	if t.Before(time.Now().Add(-5 * time.Second)) {
		return false
	}
	return true
}

func (js *JobSuite) TestSubmit() {
	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	err := Submit("no_handler", nil)
	js.NoError(err)

	errLog := buf.String()
	js.Equal("", errLog, "Got an unexpected error log entry")
}
