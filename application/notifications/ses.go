package notifications

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
)

// SES sends email using Amazon Simple Email Service (SES)
type SES struct{}

// Send a message
func (s *SES) Send(msg Message) error {
	msg.Data["uiURL"] = domain.Env.UIURL
	msg.Data["appName"] = domain.Env.AppName

	bodyBuf := &bytes.Buffer{}
	if err := eR.HTML(msg.Template).Render(bodyBuf, msg.Data); err != nil {
		return errors.New("error rendering message body - " + err.Error())
	}
	body := bodyBuf.String()

	to := addressWithName(msg.ToName, msg.ToEmail)
	from := addressWithName(msg.FromName, msg.FromEmail)

	return aws.SendEmail(to, from, msg.Subject, body)
}

func addressWithName(name, address string) string {
	if name == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", name, address)
}
