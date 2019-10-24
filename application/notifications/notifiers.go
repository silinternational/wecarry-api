package notifications

import (
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

// Notifier is an abstraction layer for multiple types of notifications: email, mobile, and push (TBD).
type Notifier interface {
	Send(msg Message) error
}

// EmailNotifier is an email notifier that conforms to the Notifier interface.
type EmailNotifier struct {
}

// DummyEmailService is an instance of a mocked email service used for test and development.
var DummyEmailService email.DummyService

// Send a notification using an email notifier.
func (e *EmailNotifier) Send(msg Message) error {
	var emailService email.Service

	emailServiceType := domain.Env.EmailService
	switch emailServiceType {
	case EmailServiceSendGrid:
		emailService = &email.SendGridService{}
	case EmailServiceDummy:
		emailService = &DummyEmailService
	default:
		emailService = &DummyEmailService
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

// MobileNotifier is an email notifier that conforms to the Notifier interface.
type MobileNotifier struct {
}

// Send a notification using a mobile notifier.
func (m *MobileNotifier) Send(msg Message) error {
	var mobileService mobile.Service

	mobileServiceType := domain.Env.MobileService
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
