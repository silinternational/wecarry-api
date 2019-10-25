package notifications

import (
	"errors"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/silinternational/wecarry-api/domain"
)

const (
	SendGridEndpointMailSend = "/v3/mail/send"
	SendGridAPIUrl           = "https://api.sendgrid.com"
)

type SendGridService struct {
}

type sendGridTemplate struct {
	// ID of SendGrid hosted template
	id string
}

var sendGridTemplates = map[string]sendGridTemplate{
	domain.MessageTemplateNewMessage:                     {id: "d-3c43c00e9c384aff99260d53f1b0d482"},
	domain.MessageTemplateNewRequest:                     {id: ""},
	domain.MessageTemplateRequestFromOpenToCommitted:     {id: "d-44a96bd9fb3846a9ab6ae9b933becf4e"},
	domain.MessageTemplateRequestFromCommittedToAccepted: {id: "d-b03639e13cb6493998f946b8ef678fab"},
}

func (e *SendGridService) Send(msg Message) error {
	apiKey := domain.Env.SendGridAPIKey
	if apiKey == "" {
		return errors.New("SendGrid API key is required")
	}

	template, ok := sendGridTemplates[msg.Template]
	if !ok {
		return fmt.Errorf("invalid message template name: %s", msg.Template)
	}

	p := mail.NewPersonalization()
	p.AddTos(mail.NewEmail(msg.ToName, msg.ToEmail))
	for key, val := range msg.Data {
		p.SetDynamicTemplateData(key, val)
	}

	message := mail.NewV3Mail()
	message.SetTemplateID(template.id)
	message.SetFrom(mail.NewEmail(msg.FromName, msg.FromEmail))
	message.AddPersonalizations(p)

	domain.Logger.Printf("email data: %+v\n", message.Personalizations[0].DynamicTemplateData)

	request := sendgrid.GetRequest(apiKey, SendGridEndpointMailSend, SendGridAPIUrl)
	request.Method = "POST"
	request.Body = mail.GetRequestBody(message)

	response, err := sendgrid.API(request)
	if err != nil {
		return fmt.Errorf("error attempting to send message, %s", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("error response (%d) from sendgrid API, %s", response.StatusCode, response.Body)
	}

	domain.Logger.Printf("mail sent, status=%v, body=%v, headers=%v",
		response.StatusCode, response.Body, response.Headers)

	return nil
}