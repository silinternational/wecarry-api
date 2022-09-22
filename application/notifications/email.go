package notifications

import (
	"github.com/gobuffalo/buffalo/render"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/templates"
)

var eR = render.New(render.Options{
	HTMLLayout:  "mail/layout.plush.html",
	TemplatesFS: templates.FS(),
	Helpers:     render.Helpers{},
})

type EmailService interface {
	Send(msg Message) error
}

// GetEmailTemplate returns the filename of the email template corresponding to a particular status change.
//  Most of those will just be the same as the name of the status change.
func GetEmailTemplate(key string) string {
	weirdTemplates := map[string]string{
		domain.MessageTemplateRequestFromAcceptedToDelivered:  domain.MessageTemplateRequestDelivered,
		domain.MessageTemplateRequestFromAcceptedToReceived:   domain.MessageTemplateRequestReceived,
		domain.MessageTemplateRequestFromAcceptedToCompleted:  domain.MessageTemplateRequestReceived,
		domain.MessageTemplateRequestFromDeliveredToCompleted: domain.MessageTemplateRequestReceived,
		domain.MessageTemplateRequestFromCompletedToAccepted:  domain.MessageTemplateRequestNotReceivedAfterAll,
		domain.MessageTemplateRequestFromCompletedToDelivered: domain.MessageTemplateRequestNotReceivedAfterAll,
	}

	template, ok := weirdTemplates[key]
	if !ok {
		template = key
	}

	return template
}
