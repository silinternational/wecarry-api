package mailers

import (
	"github.com/gobuffalo/buffalo/mail"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	ssender "github.com/paganotoni/sendgrid-sender"
	"github.com/silinternational/wecarry-api/domain"
)

var sender mail.Sender
var r *render.Engine

func init() {

	// Pulling config from the env.
	sender = ssender.NewSendgridSender(domain.Env.SendGridAPIKey)

	r = render.New(render.Options{
		HTMLLayout:   "layout.plush.html",
		TemplatesBox: packr.New("app:mailers:templates", "../templates/mail"),
		Helpers:      render.Helpers{},
	})
}
