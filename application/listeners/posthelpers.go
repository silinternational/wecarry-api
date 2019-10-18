package listeners

import (
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

type PostMsgRecipient struct {
	nickname string
	email    string
}

// GetPostRecipients returns up to two entries for the Post Requestor and
// Post Provider assuming their email is not blank.
func GetPostRecipients(eData models.PostStatusEventData) []PostMsgRecipient {
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

func sendAllNotifications(template string, eData models.PostStatusEventData) {
	recipients := GetPostRecipients(eData)

	for _, r := range recipients {
		sendNotification(template, r)
	}
}

func RequestNewStatusOpen(eData models.PostStatusEventData) {

	switch eData.OldStatus {
	case models.PostStatusCommitted: // Reverted. No longer committed
		sendAllNotifications(domain.MessageTemplateRequestFromCommittedToOpen, eData)

	case models.PostStatusAccepted: // Reverted. No longer accepted or committed
		sendAllNotifications(domain.MessageTemplateRequestFromAcceptedToOpen, eData)
	}
}

func RequestNewStatusCommitted(eData models.PostStatusEventData) {

	// Just keeping to the pattern, even though there is only one option now
	switch eData.OldStatus {
	case models.PostStatusOpen: // Advanced - normal progression
		sendAllNotifications(domain.MessageTemplateRequestFromOpenToCommitted, eData)
	}
}

func RequestNewStatusAccepted(eData models.PostStatusEventData) {

	switch eData.OldStatus {
	case models.PostStatusCommitted: // Advanced - normal progression
		sendAllNotifications(domain.MessageTemplateRequestFromCommittedToAccepted, eData)

	case models.PostStatusDelivered: // Reverted - just fixing a mis-click?
		sendAllNotifications(domain.MessageTemplateRequestFromDeliveredToAccepted, eData)

	case models.PostStatusReceived: // Reverted - just fixing a mis-click?
		sendAllNotifications(domain.MessageTemplateRequestFromReceivedToAccepted, eData)
	}
}

func RequestNewStatusDelivered(eData models.PostStatusEventData) {
	switch eData.OldStatus {
	case models.PostStatusCommitted: // Advanced - skipped Accepted
		sendAllNotifications(domain.MessageTemplateRequestFromCommittedToDelivered, eData)

	case models.PostStatusAccepted: // Advanced - normal progression
		sendAllNotifications(domain.MessageTemplateRequestFromAcceptedToDelivered, eData)

	case models.PostStatusCompleted: // Reverted - just fixing a mis-click?
		sendAllNotifications(domain.MessageTemplateRequestFromCompletedToDelivered, eData)
	}
}

func RequestNewStatusReceived(eData models.PostStatusEventData) {
	switch eData.OldStatus {
	case models.PostStatusAccepted: // Advanced - normal progression
		sendAllNotifications(domain.MessageTemplateRequestFromAcceptedToReceived, eData)

	case models.PostStatusCompleted: // Reverted - just fixing a mis-click?
		sendAllNotifications(domain.MessageTemplateRequestFromCompletedToReceived, eData)
	}
}

func RequestStatusCompleted(eData models.PostStatusEventData) {
	switch eData.OldStatus {
	case models.PostStatusDelivered: // Advanced - normal progression
		sendAllNotifications(domain.MessageTemplateRequestFromDeliveredToCompleted, eData)

	case models.PostStatusReceived: // Advanced - normal progression
		sendAllNotifications(domain.MessageTemplateRequestFromReceivedToCompleted, eData)
	}
}

func RequestNewStatusRemoved(eData models.PostStatusEventData) {
	switch eData.OldStatus {
	case models.PostStatusOpen: // Cancelling the request
		sendAllNotifications(domain.MessageTemplateRequestFromOpenToRemoved, eData)

	case models.PostStatusCommitted: // Cancelling the request
		sendAllNotifications(domain.MessageTemplateRequestFromCommittedToRemoved, eData)

	case models.PostStatusAccepted: // Cancelling the request
		sendAllNotifications(domain.MessageTemplateRequestFromAcceptedToRemoved, eData)
	}
}
