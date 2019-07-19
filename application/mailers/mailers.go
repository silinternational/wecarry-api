package mailers

import (
	"github.com/gobuffalo/buffalo/mail"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr/v2"
	ssender "github.com/paganotoni/sendgrid-sender"
)

var sender mail.Sender
var r *render.Engine

func init() {

	// Pulling config from the env.
	APIKey := envy.Get("SENDGRID_API_KEY", "")
	sender = ssender.NewSendgridSender(APIKey)

	r = render.New(render.Options{
		HTMLLayout:   "layout.html",
		TemplatesBox: packr.New("app:mailers:templates", "../templates/mail"),
		Helpers:      render.Helpers{},
	})
}
