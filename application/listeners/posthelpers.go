package listeners

import (
	"github.com/silinternational/wecarry-api/domain"
	m "github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

// The intention is to allow the tests to override this with their own Send function
var ntfSend = notifications.Send

type PostUser struct {
	Nickname string
	Email    string
}

type PostUsers struct {
	Requester PostUser
	Provider  PostUser
}

// GetPostUsers returns up to two entries for the Post Requester and
// Post Provider assuming their email is not blank.
func GetPostUsers(post m.Post) PostUsers {

	requester, _ := post.GetReceiver([]string{"Email", "Nickname"})
	provider, _ := post.GetProvider([]string{"Email", "Nickname"})

	var recipients PostUsers

	if requester != nil {
		recipients.Requester = PostUser{Nickname: requester.Nickname, Email: requester.Email}
	}

	if provider != nil {
		recipients.Provider = PostUser{Nickname: provider.Nickname, Email: provider.Email}
	}

	return recipients
}

func sendNotificationRequestFromOpenToCommitted(template string, post m.Post) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"postURL":          domain.GetPostUIURL(post.Uuid.String()),
		"postTitle":        post.Title,
		"providerNickname": postUsers.Provider.Nickname,
		"providerEmail":    postUsers.Provider.Email,
	}

	msg := notifications.Message{
		Template: template,
		Data:     data,
		ToName:   postUsers.Requester.Nickname,
		ToEmail:  postUsers.Requester.Email,
	}
	if err := ntfSend(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromCommittedToAccepted(template string, post m.Post) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	data := map[string]interface{}{
		"uiURL":             domain.Env.UIURL,
		"postURL":           domain.GetPostUIURL(post.Uuid.String()),
		"postTitle":         post.Title,
		"postDescription":   post.Description,
		"requesterNickname": postUsers.Requester.Nickname,
	}

	msg := notifications.Message{
		Template: template,
		Data:     data,
		ToName:   postUsers.Provider.Nickname,
		ToEmail:  postUsers.Provider.Email,
	}
	if err := ntfSend(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationEmpty(template string, post m.Post) {
	domain.ErrLogger.Print("Notification not implemented yet for " + template)
}

type Sender struct {
	Template string
	Sender   func(string, m.Post)
}

func join(s1, s2 string) string {
	return s1 + "-" + s2
}

var statusSenders = map[string]Sender{
	join(m.PostStatusCommitted, m.PostStatusOpen): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToOpen,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusAccepted, m.PostStatusOpen): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToOpen,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusOpen, m.PostStatusCommitted): Sender{
		Template: domain.MessageTemplateRequestFromOpenToCommitted,
		Sender:   sendNotificationRequestFromOpenToCommitted},
	join(m.PostStatusCommitted, m.PostStatusAccepted): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToAccepted,
		Sender:   sendNotificationRequestFromCommittedToAccepted},
	join(m.PostStatusDelivered, m.PostStatusAccepted): Sender{
		Template: domain.MessageTemplateRequestFromDeliveredToAccepted,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusReceived, m.PostStatusAccepted): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToAccepted,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusCommitted, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToDelivered,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusAccepted, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToDelivered,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusReceived, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToDelivered,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusCompleted, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromCompletedToDelivered,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusAccepted, m.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToReceived,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusCompleted, m.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromCompletedToReceived,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusDelivered, m.PostStatusCompleted): Sender{
		Template: domain.MessageTemplateRequestFromDeliveredToCompleted,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusReceived, m.PostStatusCompleted): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToCompleted,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusCommitted, m.PostStatusRemoved): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToRemoved,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusAccepted, m.PostStatusRemoved): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToRemoved,
		Sender:   sendNotificationEmpty},
}

func requestStatusUpdatedNotifications(post m.Post, eData m.PostStatusEventData) {

	fromStatusTo := join(eData.OldStatus, eData.NewStatus)
	sender, ok := statusSenders[fromStatusTo]

	if !ok {
		domain.ErrLogger.Printf("unexpected status transition '%s'", fromStatusTo)
		return
	}

	sender.Sender(sender.Template, post)
}
