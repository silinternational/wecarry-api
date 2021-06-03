package notifications

import (
	"github.com/silinternational/wecarry-api/domain"
)

const (
	EmailServiceSendGrid = "sendgrid"
	EmailServiceSES      = "ses"
	EmailServiceDummy    = "dummy"
	MobileServiceTwilio  = "twilio"
	MobileServiceDummy   = "dummy"
)

// Notifier is an abstraction layer for multiple types of notifications: email, mobile, and push (TBD).
type Notifier interface {
	Send(msg Message) error
}

// EmailNotifier is an email notifier that conforms to the Notifier interface.
type EmailNotifier struct{}

// Send a notification using an email notifier.
func (e *EmailNotifier) Send(msg Message) error {
	var emailService EmailService

	emailServiceType := domain.Env.EmailService
	switch emailServiceType {
	case EmailServiceSendGrid:
		emailService = &SendGridService{}
	case EmailServiceSES:
		emailService = &SES{}
	case EmailServiceDummy:
		emailService = &TestEmailService
	default:
		emailService = &TestEmailService
	}

	emailMessage := Message{
		FromName:  msg.FromName,
		FromEmail: msg.FromEmail,
		ToName:    msg.ToName,
		ToEmail:   msg.ToEmail,
		Template:  msg.Template,
		Data:      msg.Data,
		Subject:   msg.Subject,
	}

	return emailService.Send(emailMessage)
}

// MobileNotifier is an email notifier that conforms to the Notifier interface.
type MobileNotifier struct{}

// Send a notification using a mobile notifier.
func (m *MobileNotifier) Send(msg Message) error {
	var mobileService MobileService

	mobileServiceType := domain.Env.MobileService
	switch mobileServiceType {
	case MobileServiceDummy:
		mobileService = &DummyMobileService{}
	default:
		mobileService = &DummyMobileService{}
	}

	mobileMessage := Message{
		FromName:  msg.FromName,
		FromPhone: msg.FromPhone,
		ToName:    msg.ToName,
		ToPhone:   msg.ToPhone,
		Template:  msg.Template,
	}

	return mobileService.Send(mobileMessage)
}
