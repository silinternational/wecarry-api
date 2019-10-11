package email

import (
	"testing"

	"github.com/silinternational/wecarry-api/domain"
)

func TestSend(t *testing.T) {
	msg := Message{
		FromName:     "from name",
		FromEmail:    "from@example.com",
		ToName:       "to name",
		ToEmail:      "steve_schram@sil.org",
		TemplateName: domain.MessageTemplateNewMessage,
	}
	var emailService Service
	var testService DummyService
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
