package mobile

import (
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
)

type Message struct {
	FromName, FromPhone, ToName, ToPhone, TemplateName string
}

type Service interface {
	Send(msg Message) error
}

type DummyService struct {
	numberSent int
}

type template struct {
	subject, body string
}

var templateData = map[string]template{
	domain.MessageTemplateNewMessage: {
		subject: "new message", body: "You have a new message."},
	domain.MessageTemplateNewRequest: {
		subject: "new request", body: "There is a new request for an item from your location."},
	domain.MessageTemplateRequestFromOpenToCommitted: {
		subject: "potential provider", body: "Someone has offered to fulfill your request."},
	domain.MessageTemplateRequestFromCommittedToAccepted: {
		subject: "offer accepted", body: "The requester has accepted your offer."},
}

func (t *DummyService) Send(msg Message) error {
	fmt.Printf("new message sent: %s", templateData[msg.TemplateName].subject)
	t.numberSent++
	return nil
}
