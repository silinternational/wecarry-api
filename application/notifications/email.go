package notifications

import (
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/silinternational/wecarry-api/domain"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		HTMLLayout:   "layout.html",
		TemplatesBox: packr.New("app:mailers:templates", "../templates/mail"),
		Helpers:      render.Helpers{},
	})
}

type EmailService interface {
	Send(msg Message) error
}

var emailTemplates = map[string]string{
	domain.MessageTemplateNewThreadMessage:                domain.MessageTemplateNewThreadMessage,
	domain.MessageTemplateNewRequest:                      domain.MessageTemplateNewRequest,
	domain.MessageTemplateRequestFromCommittedToOpen:      domain.MessageTemplateRequestFromCommittedToOpen,
	domain.MessageTemplateRequestFromAcceptedToOpen:       domain.MessageTemplateRequestFromAcceptedToOpen,
	domain.MessageTemplateRequestFromOpenToCommitted:      domain.MessageTemplateRequestFromOpenToCommitted,
	domain.MessageTemplateRequestFromCommittedToAccepted:  domain.MessageTemplateRequestFromCommittedToAccepted,
	domain.MessageTemplateRequestFromDeliveredToAccepted:  domain.MessageTemplateRequestFromDeliveredToAccepted,
	domain.MessageTemplateRequestFromReceivedToAccepted:   domain.MessageTemplateRequestFromReceivedToAccepted,
	domain.MessageTemplateRequestFromCommittedToDelivered: "request_delivered",
	domain.MessageTemplateRequestFromAcceptedToDelivered:  "request_delivered",
	domain.MessageTemplateRequestFromReceivedToDelivered:  domain.MessageTemplateRequestFromReceivedToDelivered,
	domain.MessageTemplateRequestFromCompletedToDelivered: domain.MessageTemplateRequestFromCompletedToDelivered,
	domain.MessageTemplateRequestFromAcceptedToReceived:   domain.MessageTemplateRequestFromAcceptedToReceived,
	domain.MessageTemplateRequestFromCompletedToReceived:  domain.MessageTemplateRequestFromCompletedToReceived,
	domain.MessageTemplateRequestFromDeliveredToCompleted: domain.MessageTemplateRequestFromDeliveredToCompleted,
	domain.MessageTemplateRequestFromReceivedToCompleted:  domain.MessageTemplateRequestFromReceivedToCompleted,
	domain.MessageTemplateRequestFromOpenToRemoved:        domain.MessageTemplateRequestFromOpenToRemoved,
	domain.MessageTemplateRequestFromCommittedToRemoved:   domain.MessageTemplateRequestFromCommittedToRemoved,
	domain.MessageTemplateRequestFromAcceptedToRemoved:    domain.MessageTemplateRequestFromAcceptedToRemoved,
}

func GetEmailTemplate(key string) string {
	return emailTemplates[key]
}
