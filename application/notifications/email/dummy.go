package email

import (
	"errors"
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

var dummyTemplates = map[string]dummyTemplate{
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
	template, ok := dummyTemplates[msg.TemplateName]
	if !ok {
		errMsg := fmt.Sprintf("invalid template name: %s", msg.TemplateName)
		domain.ErrLogger.Print(errMsg)
		return errors.New(errMsg)
	}

	domain.Logger.Printf("dummy message subject: %s, recipient: %s, data: %+v",
		template.subject, msg.ToName, msg.TemplateData)
	t.sentMessages = append(t.sentMessages,
		dummyMessage{
			subject:   template.subject,
			body:      template.body,
			fromName:  msg.FromName,
			fromEmail: msg.FromEmail,
			toName:    msg.ToName,
			toEmail:   msg.ToEmail,
		})
	return nil
}
