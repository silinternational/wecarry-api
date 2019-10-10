package notifications

import "github.com/silinternational/wecarry-api/domain"

// Enabled controls which notifications are sent. Eventually this will be a per-user config.
var Enabled struct {
	Email, Mobile bool
}

func init() {
	Enabled.Email = true
}

func Send(msg Message) error {
	if Enabled.Email {
		var e Email
		if err := e.Send(msg); err != nil {
			return err
		}
		domain.Logger.Printf("%s message sent by email", msg.Template)
	}

	if Enabled.Mobile {
		var m Mobile
		if err := m.Send(msg); err != nil {
			return err
		}
		domain.Logger.Printf("%s message sent by mobile", msg.Template)
	}

	return nil
}
