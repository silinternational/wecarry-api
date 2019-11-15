package notifications

import (
	"testing"

	"github.com/silinternational/wecarry-api/domain"
)

func TestSend(t *testing.T) {
	msg := Message{
		FromName:  "from name",
		FromEmail: "from@example.com",
		ToName:    "to name",
		ToEmail:   "to@example.com",
		Template:  domain.MessageTemplateNewThreadMessage,
		Data: map[string]interface{}{
			"uiURL":          "example.com",
			"appName":        "Our App",
			"postURL":        "mypost.example.com",
			"postTitle":      "My Post",
			"messageContent": "I can bring it",
			"sentByNickname": "Fred",
			"threadURL":      "ourthread.example.com",
		},
	}
	var emailService EmailService
	var testService DummyEmailService
	emailService = &testService

	if err := emailService.Send(msg); err != nil {
		t.Errorf("error sending message, %s", err)
	}

	n := len(testService.sentMessages)
	if n != 1 {
		t.Errorf("incorrect number of messages sent (%d)", n)
		t.FailNow()
	}
}
