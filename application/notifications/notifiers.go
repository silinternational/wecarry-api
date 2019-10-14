package notifications

import (
	"github.com/gobuffalo/envy"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/notifications/email"
	"github.com/silinternational/wecarry-api/notifications/mobile"
)

const (
	EmailServiceSendGrid = "sendgrid"
	EmailServiceDummy    = "dummy"
	MobileServiceTwilio  = "twilio"
	MobileServiceDummy   = "dummy"
)

type Notifier interface {
	Send(msg Message) error
}

type EmailNotifier struct {
}

func (e *EmailNotifier) Send(msg Message) error {
	var emailService email.Service

	emailServiceType := envy.Get(domain.EmailServiceEnv, "sendgrid")
	switch emailServiceType {
	case EmailServiceSendGrid:
		emailService = &email.SendGridService{}
	case EmailServiceDummy:
		emailService = &email.DummyService{}
	default:
		emailService = &email.DummyService{}
	}

	emailMessage := email.Message{
		FromName:     msg.FromName,
		FromEmail:    msg.FromEmail,
		ToName:       msg.ToName,
		ToEmail:      msg.ToEmail,
		TemplateName: msg.Template,
		TemplateData: msg.Data,
	}

	return emailService.Send(emailMessage)
}

type MobileNotifier struct {
}

func (m *MobileNotifier) Send(msg Message) error {
	var mobileService mobile.Service

	mobileServiceType := envy.Get(domain.MobileServiceEnv, "dummy")
	switch mobileServiceType {
	case MobileServiceDummy:
		mobileService = &mobile.DummyService{}
	default:
		mobileService = &mobile.DummyService{}
	}

	mobileMessage := mobile.Message{
		FromName:     msg.FromName,
		FromPhone:    msg.FromPhone,
		ToName:       msg.ToName,
		ToPhone:      msg.ToPhone,
		TemplateName: msg.Template,
	}

	return mobileService.Send(mobileMessage)
}
