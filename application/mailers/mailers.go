package mailers

import (
	"github.com/gobuffalo/buffalo/mail"
	"github.com/gobuffalo/buffalo/render"
	ssender "github.com/paganotoni/sendgrid-sender"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/templates"
)

var (
	sender mail.Sender
	r      *render.Engine
)

func init() {
	// Pulling config from the env.
	sender = ssender.NewSendgridSender(domain.Env.SendGridAPIKey)

	r = render.New(render.Options{
		HTMLLayout:  "layout.plush.html",
		TemplatesFS: templates.FS(),
		Helpers:     render.Helpers{},
	})
}
