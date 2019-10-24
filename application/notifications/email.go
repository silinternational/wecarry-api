package notifications

type EmailService interface {
	Send(msg Message) error
}
