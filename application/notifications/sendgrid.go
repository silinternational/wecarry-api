package notifications

import (
	"bytes"
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
	domain.MessageTemplateNewThreadMessage:                {id: "d-3c43c00e9c384aff99260d53f1b0d482"},
	domain.MessageTemplateNewRequest:                      {id: ""},
	domain.MessageTemplateRequestFromOpenToCommitted:      {id: "d-44a96bd9fb3846a9ab6ae9b933becf4e"},
	domain.MessageTemplateRequestFromCommittedToAccepted:  {id: "d-b03639e13cb6493998f946b8ef678fab"},
	domain.MessageTemplateRequestFromCommittedToOpen:      {id: "d-78c3c816fd7841909fcde2455e74a986"},
	domain.MessageTemplateRequestFromCommittedToDelivered: {id: "d-11a7c98ceb8a424ca1c9e619ba9c1f16"},
	domain.MessageTemplateRequestFromCommittedToRemoved:   {id: "d-c5dd107ede1f4a11b7f1fbc79a2ddf2a"},
	domain.MessageTemplateRequestFromAcceptedToOpen:       {id: "d-4203f4bed73543468751a9667834dbd9"},
	domain.MessageTemplateRequestFromAcceptedToDelivered:  {id: "d-11a7c98ceb8a424ca1c9e619ba9c1f16"},
	domain.MessageTemplateRequestFromAcceptedToReceived:   {id: "d-f963c27c30d44b1fb06e17973865bb3a"},
	domain.MessageTemplateRequestFromAcceptedToRemoved:    {id: "d-82f1b7a2d6974ac88f7979763ec081e5"},
}

func (e *SendGridService) Send(msg Message) error {
	apiKey := domain.Env.SendGridAPIKey
	if apiKey == "" {
		return errors.New("SendGrid API key is required")
	}

	subject := "New Message on " + domain.Env.AppName
	from := mail.NewEmail(msg.FromName, msg.FromEmail)
	to := mail.NewEmail(msg.ToName, msg.ToEmail)

	msg.Data["uiURL"] = domain.Env.UIURL
	msg.Data["appName"] = domain.Env.AppName

	bodyBuf := &bytes.Buffer{}
	if err := r.HTML(msg.Template).Render(bodyBuf, msg.Data); err != nil {
		return errors.New("error rendering message body - " + err.Error())
	}
	body := bodyBuf.String()

	m := mail.NewSingleEmail(from, subject, to, body, body)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(m)

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
