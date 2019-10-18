package listeners

import (
	"github.com/silinternational/wecarry-api/domain"
	m "github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

type PostMsgRecipient struct {
	nickname string
	email    string
}

// GetPostRecipients returns up to two entries for the Post Requestor and
// Post Provider assuming their email is not blank.
func GetPostRecipients(eData m.PostStatusEventData) []PostMsgRecipient {
	post := eData.Post

	var recipients []PostMsgRecipient

	if post.Receiver.Email != "" {
		r := PostMsgRecipient{nickname: post.Receiver.Nickname, email: post.Receiver.Email}
		recipients = []PostMsgRecipient{r}
	}

	if post.Provider.Email != "" {
		r := PostMsgRecipient{nickname: post.Provider.Nickname, email: post.Provider.Email}
		recipients = append(recipients, r)
	}

	return recipients
}

func sendNotification(template string, recipient PostMsgRecipient) {
	msg := notifications.Message{
		Template: template,
		ToName:   recipient.nickname,
		ToEmail:  recipient.email,
	}
	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendAllNotifications(template string, eData m.PostStatusEventData) {
	recipients := GetPostRecipients(eData)

	for _, r := range recipients {
		sendNotification(template, r)
	}
}

func join(s1, s2 string) string {
	return s1 + "-" + s2
}

var statusTemplates = map[string]string{
	join(m.PostStatusCommitted, m.PostStatusOpen):      domain.MessageTemplateRequestFromCommittedToOpen,
	join(m.PostStatusAccepted, m.PostStatusOpen):       domain.MessageTemplateRequestFromAcceptedToOpen,
	join(m.PostStatusOpen, m.PostStatusCommitted):      domain.MessageTemplateRequestFromOpenToCommitted,
	join(m.PostStatusCommitted, m.PostStatusAccepted):  domain.MessageTemplateRequestFromCommittedToAccepted,
	join(m.PostStatusDelivered, m.PostStatusAccepted):  domain.MessageTemplateRequestFromDeliveredToAccepted,
	join(m.PostStatusReceived, m.PostStatusAccepted):   domain.MessageTemplateRequestFromReceivedToAccepted,
	join(m.PostStatusCommitted, m.PostStatusDelivered): domain.MessageTemplateRequestFromCommittedToDelivered,
	join(m.PostStatusAccepted, m.PostStatusDelivered):  domain.MessageTemplateRequestFromAcceptedToDelivered,
	join(m.PostStatusCompleted, m.PostStatusDelivered): domain.MessageTemplateRequestFromCompletedToDelivered,
	join(m.PostStatusAccepted, m.PostStatusReceived):   domain.MessageTemplateRequestFromAcceptedToReceived,
	join(m.PostStatusCompleted, m.PostStatusReceived):  domain.MessageTemplateRequestFromCompletedToReceived,
	join(m.PostStatusDelivered, m.PostStatusCompleted): domain.MessageTemplateRequestFromDeliveredToCompleted,
	join(m.PostStatusReceived, m.PostStatusCompleted):  domain.MessageTemplateRequestFromReceivedToCompleted,
	join(m.PostStatusCommitted, m.PostStatusRemoved):   domain.MessageTemplateRequestFromCommittedToRemoved,
	join(m.PostStatusAccepted, m.PostStatusRemoved):    domain.MessageTemplateRequestFromAcceptedToRemoved,
}

func requestStatusUpdatedNotifications(eData m.PostStatusEventData) {

	fromStatusTo := join(eData.OldStatus, eData.NewStatus)
	template, ok := statusTemplates[fromStatusTo]

	if !ok {
		domain.ErrLogger.Printf("unexpected status transition '%s'", fromStatusTo)
	}
	sendAllNotifications(template, eData)
}
