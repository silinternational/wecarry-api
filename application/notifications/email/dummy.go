package email

import (
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
)

type DummyService struct {
	sentMessages []dummyMessage
}

type dummyMessage struct {
	subject, body, fromName, fromEmail, toName, toEmail string
}

type dummyTemplate struct {
	subject, body string
}

var dummyTemplateData = map[string]dummyTemplate{
	domain.MessageTemplateNewRequest: {
		subject: "new request",
		body:    "There is a new request for an item from your location.",
	},
	domain.MessageTemplateNewMessage: {
		subject: "new message",
		body:    "You have a new message.",
	},
}

func (t *DummyService) Send(msg Message) error {
	fmt.Printf("new message sent: %s", dummyTemplateData[msg.TemplateName].subject)
	t.sentMessages = append(t.sentMessages,
		dummyMessage{
			subject:   dummyTemplateData[msg.TemplateName].subject,
			body:      dummyTemplateData[msg.TemplateName].body,
			fromName:  msg.FromName,
			fromEmail: msg.FromEmail,
			toName:    msg.ToName,
			toEmail:   msg.ToEmail,
		})
	return nil
}
