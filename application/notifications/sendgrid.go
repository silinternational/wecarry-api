package notifications

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"jaytaylor.com/html2text"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
)

type SendGridService struct{}

type sendGridTemplate struct {
	// ID of SendGrid hosted template
	id string
}

func (e *SendGridService) Send(msg Message) error {
	apiKey := domain.Env.SendGridAPIKey
	if apiKey == "" {
		return errors.New("SendGrid API key is required")
	}

	from := mail.NewEmail(msg.FromName, msg.FromEmail)
	to := mail.NewEmail(msg.ToName, msg.ToEmail)

	msg.Data["uiURL"] = domain.Env.UIURL
	msg.Data["appName"] = domain.Env.AppName

	bodyBuf := &bytes.Buffer{}
	if err := eR.HTML(mailTemplatePath+msg.Template).Render(bodyBuf, msg.Data); err != nil {
		return errors.New("error rendering message body - " + err.Error())
	}
	body := bodyBuf.String()

	tbody, err := html2text.FromString(body)
	if err != nil {
		log.Errorf("error converting html email to plain text ... %s", err.Error())
		tbody = body
	}

	m := mail.NewSingleEmail(from, msg.Subject, to, tbody, body)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(m)
	if err != nil {
		return fmt.Errorf("error attempting to send message, %s", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("error response (%d) from sendgrid API, %s", response.StatusCode, response.Body)
	}

	log.Infof("mail sent, status=%v, body=%v, headers=%v",
		response.StatusCode, response.Body, response.Headers)

	return nil
}
