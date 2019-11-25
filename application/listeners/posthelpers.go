package listeners

import (
	"errors"
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

// TODO create a proper solution for this
func getUserLanguage() string {
	return "en"
}

var argAppName = map[string]string{"AppName": domain.Env.AppName}

func getTranslatedSubject(translationID, template string) string {
	subj, err := domain.TranslateWithLang(getUserLanguage(), translationID, argAppName)

	if err != nil {
		domain.ErrLogger.Printf("error translating '%s' notification subject, %s", template, err)
	}

	return subj
}

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

	requester, _ := post.GetReceiver()
	provider, _ := post.GetProvider()

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

func sendNotificationRequestFromAcceptedOrCommittedToDelivered(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForRequester(postUsers, post, template)
	msg.Subject = getTranslatedSubject("Email.Subject.Request.FromAcceptedOrCommittedToDelivered", template)

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
		Subject:   fmt.Sprintf("You are no longer expected to fulfill a certain %s request", domain.Env.AppName),
	}
	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedOrCommittedToReceived(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)
	msg.Subject = getTranslatedSubject("Email.Subject.Request.FromAcceptedOrCommittedToReceived", template)

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
	msg.Subject = getTranslatedSubject("Email.Subject.Request.FromAcceptedToRemoved", template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromCommittedToAccepted(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)
	msg.Subject = getTranslatedSubject("Email.Subject.Request.FromCommittedToAccepted", template)

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
		Subject:   getTranslatedSubject("Email.Subject.Request.FromCommittedToOpen", template),
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

func sendNotificationRequestFromCommittedToRemoved(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)
	msg.Subject = getTranslatedSubject("Email.Subject.Request.FromCommittedToRemoved", template)

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromOpenToCommitted(template string, post models.Post, eData models.PostStatusEventData) {
	postUsers := GetPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForRequester(postUsers, post, template)
	msg.Subject = getTranslatedSubject("Email.Subject.Request.FromOpenToCommitted", template)

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

func join(s1, s2 models.PostStatus) string {
	return fmt.Sprintf("%s-%s", s1, s2)
}

var statusSenders = map[string]Sender{
	join(models.PostStatusAccepted, models.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToDelivered,
		Sender:   sendNotificationRequestFromAcceptedOrCommittedToDelivered},
	join(models.PostStatusAccepted, models.PostStatusOpen): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToOpen,
		Sender:   sendNotificationRequestFromAcceptedToOpen},
	join(models.PostStatusAccepted, models.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToReceived,
		Sender:   sendNotificationRequestFromAcceptedOrCommittedToReceived},
	join(models.PostStatusAccepted, models.PostStatusRemoved): Sender{
		Template: domain.MessageTemplateRequestFromAcceptedToRemoved,
		Sender:   sendNotificationRequestFromAcceptedToRemoved},
	join(models.PostStatusCommitted, models.PostStatusAccepted): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToAccepted,
		Sender:   sendNotificationRequestFromCommittedToAccepted},
	join(models.PostStatusCommitted, models.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToDelivered,
		Sender:   sendNotificationRequestFromAcceptedOrCommittedToDelivered},
	join(models.PostStatusCommitted, models.PostStatusOpen): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToOpen,
		Sender:   sendNotificationRequestFromCommittedToOpen},
	join(models.PostStatusCommitted, models.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToReceived,
		Sender:   sendNotificationRequestFromAcceptedOrCommittedToReceived},
	join(models.PostStatusCommitted, models.PostStatusRemoved): Sender{
		Template: domain.MessageTemplateRequestFromCommittedToRemoved,
		Sender:   sendNotificationRequestFromCommittedToRemoved},
	join(models.PostStatusCompleted, models.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromCompletedToDelivered,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusCompleted, models.PostStatusReceived): Sender{
		Template: domain.MessageTemplateRequestFromCompletedToReceived,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusDelivered, models.PostStatusAccepted): Sender{
		Template: domain.MessageTemplateRequestFromDeliveredToAccepted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusDelivered, models.PostStatusCompleted): Sender{
		Template: domain.MessageTemplateRequestFromDeliveredToCompleted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusOpen, models.PostStatusCommitted): Sender{
		Template: domain.MessageTemplateRequestFromOpenToCommitted,
		Sender:   sendNotificationRequestFromOpenToCommitted},
	join(models.PostStatusReceived, models.PostStatusAccepted): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToAccepted,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusReceived, models.PostStatusDelivered): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToDelivered,
		Sender:   sendNotificationEmpty},
	join(models.PostStatusReceived, models.PostStatusCompleted): Sender{
		Template: domain.MessageTemplateRequestFromReceivedToCompleted,
		Sender:   sendNotificationEmpty},
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
		}
	}
}

func sendNewPostNotification(user models.User, post models.Post) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	newPostTemplates := map[string]string{
		models.PostTypeRequest.String(): domain.MessageTemplateNewRequest,
		models.PostTypeOffer.String():   domain.MessageTemplateNewOffer,
	}

	msg := notifications.Message{
		Template:  newPostTemplates[post.Type.String()],
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
