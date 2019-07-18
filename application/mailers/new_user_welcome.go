package mailers

import (
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/buffalo/mail"
)

func SendNewUserWelcomes() error {
	m := mail.NewMessage()

	// fill in with your stuff:
	m.Subject = "New User Welcome"
	m.From = ""
	m.To = []string{}
	err := m.AddBody(r.HTML("new_user_welcome.html"), render.Data{})
	if err != nil {
		return err
	}
	return smtp.Send(m)
}
