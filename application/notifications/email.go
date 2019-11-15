package notifications

import (
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/silinternational/wecarry-api/domain"
)

var eR = render.New(render.Options{
	HTMLLayout:   "layout.html",
	TemplatesBox: packr.New("app:mailers:templates", "../templates/mail"),
	Helpers:      render.Helpers{},
})

type EmailService interface {
	Send(msg Message) error
}

func GetEmailTemplate(key string) string {

	weirdTemplates := map[string]string{
		domain.MessageTemplateRequestFromCommittedToDelivered: domain.MessageTemplateRequestDelivered,
		domain.MessageTemplateRequestFromAcceptedToDelivered:  domain.MessageTemplateRequestDelivered,
	}

	template, ok := weirdTemplates[key]
	if !ok {
		template = key
	}

	return template
}
