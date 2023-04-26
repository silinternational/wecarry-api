package notifications

import "github.com/silinternational/wecarry-api/log"

const mailTemplatePath = "mail/"

var notifiers []Notifier

func init() {
	email := EmailNotifier{} // The type of sender is determined by domain.Env.EmailService
	notifiers = append(notifiers, &email)
}

func Send(msg Message) error {
	for _, n := range notifiers {
		if err := n.Send(msg); err != nil {
			return err
		}
		log.Errorf("%T: %s message sent to %s", n, msg.Template, msg.ToEmail)
	}

	return nil
}
