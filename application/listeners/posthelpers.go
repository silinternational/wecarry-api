package listeners

import (
	"errors"
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
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
func GetPostUsers(post models.Post) PostUsers {

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

func getMessageForProvider(postUsers PostUsers, post models.Post, template string) notifications.Message {
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
		Template:  template,
		Data:      data,
		ToName:    postUsers.Provider.Nickname,
		ToEmail:   postUsers.Provider.Email,
		FromEmail: domain.Env.EmailFromAddress,
	}
}

func getMessageForRequester(postUsers PostUsers, post models.Post, template string) notifications.Message {
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
		Template:  template,
		Data:      data,
		ToName:    postUsers.Requester.Nickname,
		ToEmail:   postUsers.Requester.Email,
		FromEmail: domain.Env.EmailFromAddress,
	}
}

func sendNotificationRequestFromOpenToCommitted(template string, post models.Post, eData models.PostStatusEventData) {
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
func sendNotificationRequestFromCommittedToOpen(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	oldProvider := models.User{}
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
		Template:  template,
		Data:      data,
		ToName:    postUsers.Requester.Nickname,
		ToEmail:   postUsers.Requester.Email,
		FromEmail: domain.Env.EmailFromAddress,
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

func sendNotificationRequestFromCommittedToAccepted(template string, post models.Post, eData models.PostStatusEventData) {
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

func sendNotificationRequestFromCommittedToRemoved(template string, post models.Post, eData models.PostStatusEventData) {
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

func sendNotificationRequestFromCommittedOrAcceptedToDelivered(template string, post models.Post, eData models.PostStatusEventData) {
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

func sendNotificationRequestFromAcceptedToReceived(template string, post models.Post, eData models.PostStatusEventData) {
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

func sendNotificationRequestFromAcceptedToOpen(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	oldProvider := models.User{}
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
		Template:  template,
		Data:      data,
		ToName:    oldProvider.Nickname,
		ToEmail:   oldProvider.Email,
		FromEmail: domain.Env.EmailFromAddress,
	}
	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedToRemoved(template string, post models.Post, eData models.PostStatusEventData) {
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

func sendNotificationEmpty(template string, post models.Post, eData models.PostStatusEventData) {
	domain.ErrLogger.Print("Notification not implemented yet for " + template)
}

type Sender struct {
	Template string
	Sender   func(string, models.Post, models.PostStatusEventData)
}

func join(s1, s2 string) string {
	return s1 + "-" + s2
}

var statusSenders = map[string]Sender{
	join(models.PostStatusCommitted, models.PostStatusOpen): {
		Template: domain.MessageTemplateRequestFromCommittedToOpen,
		Sender:   sendNotificationRequestFromCommittedToOpen},
	join(models.PostStatusAccepted, models.PostStatusOpen): {
		Template: domain.MessageTemplateRequestFromAcceptedToOpen,
		Sender:   sendNotificationRequestFromAcceptedToOpen},
	join(models.PostStatusOpen, models.PostStatusCommitted): {
		Template: domain.MessageTemplateRequestFromOpenToCommitted,
		Sender:   sendNotificationRequestFromOpenToCommitted},
	join(models.PostStatusCommitted, models.PostStatusAccepted): {
		Template: domain.MessageTemplateRequestFromCommittedToAccepted,
		Sender:   sendNotificationRequestFromCommittedToAccepted},
	join(models.PostStatusDelivered, models.PostStatusAccepted): {
		Template: domain.MessageTemplateRequestFromDeliveredToAccepted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusReceived, models.PostStatusAccepted): {
		Template: domain.MessageTemplateRequestFromReceivedToAccepted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusCommitted, models.PostStatusDelivered): {
		Template: domain.MessageTemplateRequestFromCommittedToDelivered,
		Sender:   sendNotificationRequestFromCommittedOrAcceptedToDelivered},
	join(models.PostStatusAccepted, models.PostStatusDelivered): {
		Template: domain.MessageTemplateRequestFromAcceptedToDelivered,
		Sender:   sendNotificationRequestFromCommittedOrAcceptedToDelivered},
	join(models.PostStatusReceived, models.PostStatusDelivered): {
		Template: domain.MessageTemplateRequestFromReceivedToDelivered,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusCompleted, models.PostStatusDelivered): {
		Template: domain.MessageTemplateRequestFromCompletedToDelivered,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusCommitted, models.PostStatusReceived): {
		Template: domain.MessageTemplateRequestFromCompletedToReceived,
		Sender:   sendNotificationRequestFromAcceptedToReceived},
	join(models.PostStatusAccepted, models.PostStatusReceived): {
		Template: domain.MessageTemplateRequestFromAcceptedToReceived,
		Sender:   sendNotificationRequestFromAcceptedToReceived},
	join(models.PostStatusCompleted, models.PostStatusReceived): {
		Template: domain.MessageTemplateRequestFromCompletedToReceived,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusDelivered, models.PostStatusCompleted): {
		Template: domain.MessageTemplateRequestFromDeliveredToCompleted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusReceived, models.PostStatusCompleted): {
		Template: domain.MessageTemplateRequestFromReceivedToCompleted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusCommitted, models.PostStatusRemoved): {
		Template: domain.MessageTemplateRequestFromCommittedToRemoved,
		Sender:   sendNotificationRequestFromCommittedToRemoved},
	join(models.PostStatusAccepted, models.PostStatusRemoved): {
		Template: domain.MessageTemplateRequestFromAcceptedToRemoved,
		Sender:   sendNotificationRequestFromAcceptedToRemoved},
}

func requestStatusUpdatedNotifications(post models.Post, eData models.PostStatusEventData) {

	fromStatusTo := join(eData.OldStatus, eData.NewStatus)
	sender, ok := statusSenders[fromStatusTo]

	if !ok {
		domain.ErrLogger.Printf("unexpected status transition '%s'", fromStatusTo)
		return
	}

	sender.Sender(notifications.GetEmailTemplate(sender.Template), post, eData)
}

func sendNewPostNotifications(post models.Post, users models.Users) {
	for i, user := range users {
		if !user.WantsPostNotification(post) {
			continue
		}

		if err := sendNewPostNotification(user, post); err != nil {
			domain.ErrLogger.Printf("error sending post created notification (%d of %d), %s",
				i, len(users), err)
			return
		}
	}
}

func sendNewPostNotification(user models.User, post models.Post) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	var template string
	if post.Type == models.PostTypeRequest {
		template = domain.MessageTemplateNewRequest
	} else if post.Type == models.PostTypeOffer {
		template = domain.MessageTemplateNewOffer
	} else {
		return fmt.Errorf("invalid post type %s", post.Type)
	}
	msg := notifications.Message{
		Template:  template,
		ToName:    user.Nickname,
		ToEmail:   user.Email,
		FromEmail: domain.Env.EmailFromAddress,
		Data: map[string]interface{}{
			"appName":   domain.Env.AppName,
			"uiURL":     domain.Env.UIURL,
			"postURL":   domain.GetPostUIURL(post.Uuid.String()),
			"postTitle": post.Title,
		},
	}
	return notifications.Send(msg)
}
