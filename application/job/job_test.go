package job

import (
	"bytes"
	"os"
	"testing"

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

	args := map[string]interface{}{
		domain.ArgMessageID: f.Messages[0].ID,
	}
	err := NewMessageHandler(args)
	js.NoError(err)

	errLog := buf.String()
	js.Equal("", errLog, "Got an unexpected error log entry")

	js.Equal(1, notifications.DummyEmailService.GetNumberOfMessagesSent())

	tests := []struct {
		message            models.Message
		wantNumberOfEmails int
		wantErr            bool
	}{
		{
			message:            f.Messages[0],
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
				//js.Equal(test.wantNumberOfEmails, notifications.DummyEmailService.GetNumberOfMessagesSent())
			}
		})
	}
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
