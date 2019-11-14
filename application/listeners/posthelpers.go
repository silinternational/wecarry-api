package listeners

import (
	"github.com/silinternational/wecarry-api/domain"
	m "github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

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

func getMessageForProvider(postUsers PostUsers, post m.Post, template string) notifications.Message {
	data := map[string]interface{}{
		"uiURL":             domain.Env.UIURL,
		"appName":           domain.Env.AppName,
		"postURL":           domain.GetPostUIURL(post.Uuid.String()),
		"postTitle":         post.Title,
		"postDescription":   post.Description,
		"requesterNickname": postUsers.Requester.Nickname,
		"requesterEmail":    postUsers.Requester.Email,
	}

	return notifications.Message{
		Template: template,
		Data:     data,
		ToName:   postUsers.Provider.Nickname,
		ToEmail:  postUsers.Provider.Email,
	}
}

func getMessageForRequester(postUsers PostUsers, post m.Post, template string) notifications.Message {
	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"appName":          domain.Env.AppName,
		"postURL":          domain.GetPostUIURL(post.Uuid.String()),
		"postTitle":        post.Title,
		"postDescription":  post.Description,
		"providerNickname": postUsers.Provider.Nickname,
		"providerEmail":    postUsers.Provider.Email,
	}

	return notifications.Message{
		Template: template,
		Data:     data,
		ToName:   postUsers.Requester.Nickname,
		ToEmail:  postUsers.Requester.Email,
	}
}

func sendNotificationRequestFromOpenToCommitted(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForRequester(postUsers, post, template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

// Until we have status auditing history, we don't know who reverted the Post to `open` status.
//  So, tell both the requester and provider about it.
func sendNotificationRequestFromCommittedToOpen(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	oldProvider := m.User{}
	if err := oldProvider.FindByID(eData.OldProviderID); err != nil {
		domain.ErrLogger.Printf("error preparing '%s' notification for old provider id, %v ... %v",
			template, eData.OldProviderID, err)
	}

	providerNickname := oldProvider.Nickname
	providerEmail := oldProvider.Email

	if providerNickname == "" {
		providerNickname = "Unknown User"
		providerEmail = "Missing Email"
	}

	// First notify requester
	data := map[string]interface{}{
		"uiURL":             domain.Env.UIURL,
		"appName":           domain.Env.AppName,
		"postURL":           domain.GetPostUIURL(post.Uuid.String()),
		"postTitle":         post.Title,
		"providerNickname":  providerNickname,
		"providerEmail":     providerEmail,
		"requesterNickname": postUsers.Requester.Nickname,
		"requesterEmail":    postUsers.Requester.Email,
	}

	msg := notifications.Message{
		Template: template,
		Data:     data,
		ToName:   postUsers.Requester.Nickname,
		ToEmail:  postUsers.Requester.Email,
	}
	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to old provider, %s", template, err)
	}

	// Now notify the old provider
	if oldProvider.Nickname == "" {
		return
	}

	msg.ToName = oldProvider.Nickname
	msg.ToEmail = oldProvider.Email

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to requester, %s", template, err)
	}
}

func sendNotificationRequestFromCommittedToAccepted(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromCommittedToRemoved(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromCommittedOrAcceptedToDelivered(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForRequester(postUsers, post, template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedToReceived(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedToOpen(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	oldProvider := m.User{}
	if err := oldProvider.FindByID(eData.OldProviderID); err != nil {
		domain.ErrLogger.Printf("error preparing '%s' notification for old provider id, %v ... %v",
			template, eData.OldProviderID, err)
		return
	}

	data := map[string]interface{}{
		"uiURL":             domain.Env.UIURL,
		"appName":           domain.Env.AppName,
		"postURL":           domain.GetPostUIURL(post.Uuid.String()),
		"postTitle":         post.Title,
		"requesterNickname": postUsers.Requester.Nickname,
		"requesterEmail":    postUsers.Requester.Email,
	}

	msg := notifications.Message{
		Template: template,
		Data:     data,
		ToName:   oldProvider.Nickname,
		ToEmail:  oldProvider.Email,
	}
	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedToRemoved(template string, post m.Post, eData m.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationEmpty(template string, post m.Post, eData m.PostStatusEventData) {
	domain.ErrLogger.Print("Notification not implemented yet for " + template)
}

type Sender struct {
	Template string
	Sender   func(string, m.Post, m.PostStatusEventData)
}

func join(s1, s2 string) string {
	return s1 + "-" + s2
}

var getT = notifications.GetEmailTemplate

var statusSenders = map[string]Sender{
	join(m.PostStatusCommitted, m.PostStatusOpen): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToOpen,
		Sender:   sendNotificationRequestFromCommittedToOpen},
	join(m.PostStatusAccepted, m.PostStatusOpen): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToOpen,
		Sender:   sendNotificationRequestFromAcceptedToOpen},
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
		Sender:   sendNotificationRequestFromCommittedOrAcceptedToDelivered},
	join(m.PostStatusAccepted, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToDelivered,
		Sender:   sendNotificationRequestFromCommittedOrAcceptedToDelivered},
	join(m.PostStatusReceived, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToDelivered,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusCompleted, m.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromCompletedToDelivered,
		Sender:   sendNotificationEmpty},
	join(m.PostStatusCommitted, m.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromCompletedToReceived,
		Sender:   sendNotificationRequestFromAcceptedToReceived},
	join(m.PostStatusAccepted, m.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToReceived,
		Sender:   sendNotificationRequestFromAcceptedToReceived},
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
		Sender:   sendNotificationRequestFromCommittedToRemoved},
	join(m.PostStatusAccepted, m.PostStatusRemoved): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToRemoved,
		Sender:   sendNotificationRequestFromAcceptedToRemoved},
}

func requestStatusUpdatedNotifications(post m.Post, eData m.PostStatusEventData) {

	fromStatusTo := join(eData.OldStatus, eData.NewStatus)
	sender, ok := statusSenders[fromStatusTo]

	if !ok {
		domain.ErrLogger.Printf("unexpected status transition '%s'", fromStatusTo)
		return
	}

	sender.Sender(getT(sender.Template), post, eData)
}
