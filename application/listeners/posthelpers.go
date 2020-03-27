package listeners

import (
	"errors"
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const postTitleKey = "postTitle"

type postUser struct {
	Language string
	Nickname string
	Email    string
}

type postUsers struct {
	Receiver postUser
	Provider postUser
}

// getPostUsers returns up to two entries for the Request Requester and
// Request Provider assuming their email is not blank.
func getPostUsers(post models.Request) postUsers {

	receiver, _ := post.GetCreator()
	provider, _ := post.GetProvider()

	var recipients postUsers

	if receiver != nil {
		recipients.Receiver = postUser{
			Language: receiver.GetLanguagePreference(),
			Nickname: receiver.Nickname,
			Email:    receiver.Email,
		}
	}

	if provider != nil {
		recipients.Provider = postUser{
			Language: provider.GetLanguagePreference(),
			Nickname: provider.Nickname,
			Email:    provider.Email,
		}
	}

	return recipients
}

func getMessageForProvider(postUsers postUsers, post models.Request, template string) notifications.Message {
	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"appName":          domain.Env.AppName,
		"postURL":          domain.GetPostUIURL(post.UUID.String()),
		"postTitle":        domain.Truncate(post.Title, "...", 16),
		"postDescription":  post.Description,
		"receiverNickname": postUsers.Receiver.Nickname,
		"receiverEmail":    postUsers.Receiver.Email,
	}

	return notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    postUsers.Provider.Nickname,
		ToEmail:   postUsers.Provider.Email,
		FromEmail: domain.EmailFromAddress(nil),
	}
}

func getMessageForReceiver(postUsers postUsers, post models.Request, template string) notifications.Message {
	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"appName":          domain.Env.AppName,
		"postURL":          domain.GetPostUIURL(post.UUID.String()),
		"postTitle":        domain.Truncate(post.Title, "...", 16),
		"postDescription":  post.Description,
		"providerNickname": postUsers.Provider.Nickname,
		"providerEmail":    postUsers.Provider.Email,
	}

	return notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    postUsers.Receiver.Nickname,
		ToEmail:   postUsers.Receiver.Email,
		FromEmail: domain.EmailFromAddress(nil),
	}
}

func getPotentialProviderMessageForReceiver(
	requester models.User, providerNickname, template string, post models.Request) notifications.Message {
	data := map[string]interface{}{
		"appName":          domain.Env.AppName,
		"uiURL":            domain.Env.UIURL,
		"postURL":          domain.GetPostUIURL(post.UUID.String()),
		"postTitle":        domain.Truncate(post.Title, "...", 16),
		"providerNickname": providerNickname,
	}

	return notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    requester.GetRealName(),
		ToEmail:   requester.Email,
		FromEmail: domain.EmailFromAddress(nil),
	}
}

func sendNotificationRequestToProvider(params senderParams) {
	post := params.post
	template := params.template
	postUsers := getPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)
	msg.Subject = domain.GetTranslatedSubject(postUsers.Provider.Language, params.subject,
		map[string]string{postTitleKey: post.Title})

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestToReceiver(params senderParams) {
	post := params.post
	template := params.template

	postUsers := getPostUsers(post)

	if postUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForReceiver(postUsers, post, template)
	msg.Subject = domain.GetTranslatedSubject(postUsers.Receiver.Language, params.subject,
		map[string]string{postTitleKey: post.Title})

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedToDelivered(params senderParams) {
	sendNotificationRequestToReceiver(params)
}

func sendNotificationRequestFromAcceptedToOpen(params senderParams) {
	post := params.post
	template := params.template
	eData := params.pEventData

	postUsers := getPostUsers(post)

	oldProvider := models.User{}
	if err := oldProvider.FindByID(eData.OldProviderID); err != nil {
		domain.ErrLogger.Printf("error preparing '%s' notification for old provider id, %v ... %v",
			template, eData.OldProviderID, err)
		return
	}

	msg := getMessageForProvider(postUsers, post, template)

	msg.ToName = oldProvider.GetRealName()
	msg.ToEmail = oldProvider.Email
	msg.Subject = domain.GetTranslatedSubject(oldProvider.GetLanguagePreference(), params.subject,
		map[string]string{postTitleKey: post.Title})

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedOrDeliveredToCompleted(params senderParams) {
	sendNotificationRequestToProvider(params)
}

func sendNotificationRequestFromAcceptedToRemoved(params senderParams) {
	sendNotificationRequestToProvider(params)
}

func sendRejectionToPotentialProvider(potentialProvider models.User, post models.Request) {
	template := domain.MessageTemplatePotentialProviderRejected
	ppNickname := potentialProvider.Nickname
	ppEmail := potentialProvider.Email

	if ppNickname == "" {
		ppNickname = "Unknown User"
		ppEmail = "Missing Email"
	}

	receiver, err := post.GetCreator()
	if err != nil {
		domain.ErrLogger.Printf("error getting Request Receiver for email data, %s", err)
	}

	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"appName":          domain.Env.AppName,
		"postURL":          domain.GetPostUIURL(post.UUID.String()),
		"postTitle":        domain.Truncate(post.Title, "...", 16),
		"ppNickname":       ppNickname,
		"ppEmail":          ppEmail,
		"receiverNickname": receiver.Nickname,
	}

	subject := "Email.Subject.Request.OfferRejected"

	msg := notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    potentialProvider.GetRealName(),
		ToEmail:   ppEmail,
		FromEmail: domain.EmailFromAddress(nil),
		Subject: domain.GetTranslatedSubject(potentialProvider.GetLanguagePreference(), subject,
			map[string]string{postTitleKey: post.Title}),
	}

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to rejected potentialProvider, %s", template, err)
	}
}

func sendNotificationRequestFromOpenToAccepted(params senderParams) {
	sendNotificationRequestToProvider(params)
	post := params.post

	var providers models.PotentialProviders
	users, err := providers.FindUsersByRequestID(post, models.User{})
	if err != nil {
		domain.ErrLogger.Printf("error finding rejected potential providers for post id, %v ... %v",
			post.ID, err)
	}

	for _, u := range users {
		if u.ID != post.ProviderID.Int {
			sendRejectionToPotentialProvider(u, post)
		}
	}

}

func sendNotificationRequestFromDeliveredToAccepted(params senderParams) {
	sendNotificationRequestToReceiver(params)
}

func sendNotificationRequestFromCompletedToAcceptedOrDelivered(params senderParams) {
	sendNotificationRequestToProvider(params)
}

func sendNotificationEmpty(params senderParams) {
	domain.ErrLogger.Printf("Notification not implemented yet for %s", params.template)
}

type senderParams struct {
	template   string
	subject    string
	post       models.Request
	pEventData models.RequestStatusEventData
}

type sender struct {
	template string
	subject  string
	sender   func(senderParams) // string, string, models.Request, models.RequestStatusEventData)
}

func join(s1, s2 models.RequestStatus) string {
	return fmt.Sprintf("%s-%s", s1, s2)
}

var statusSenders = map[string]sender{
	join(models.RequestStatusAccepted, models.RequestStatusCompleted): sender{
		template: domain.MessageTemplateRequestFromAcceptedToCompleted,
		subject:  "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
		sender:   sendNotificationRequestFromAcceptedOrDeliveredToCompleted},

	join(models.RequestStatusAccepted, models.RequestStatusOpen): sender{
		template: domain.MessageTemplateRequestFromAcceptedToOpen,
		subject:  "Email.Subject.Request.FromAcceptedToOpen",
		sender:   sendNotificationRequestFromAcceptedToOpen},

	join(models.RequestStatusAccepted, models.RequestStatusReceived): sender{
		template: domain.MessageTemplateRequestFromAcceptedToCompleted,
		subject:  "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
		sender:   sendNotificationRequestFromAcceptedOrDeliveredToCompleted},

	join(models.RequestStatusAccepted, models.RequestStatusRemoved): sender{
		template: domain.MessageTemplateRequestFromAcceptedToRemoved,
		subject:  "Email.Subject.Request.FromAcceptedToRemoved",
		sender:   sendNotificationRequestFromAcceptedToRemoved},

	join(models.RequestStatusCompleted, models.RequestStatusAccepted): sender{
		template: domain.MessageTemplateRequestFromCompletedToAccepted,
		subject:  "Email.Subject.Request.FromCompletedToAcceptedOrDelivered",
		sender:   sendNotificationRequestFromCompletedToAcceptedOrDelivered},

	join(models.RequestStatusCompleted, models.RequestStatusDelivered): sender{
		template: domain.MessageTemplateRequestFromCompletedToDelivered,
		subject:  "Email.Subject.Request.FromCompletedToAcceptedOrDelivered",
		sender:   sendNotificationRequestFromCompletedToAcceptedOrDelivered},

	join(models.RequestStatusCompleted, models.RequestStatusReceived): sender{
		template: domain.MessageTemplateRequestFromCompletedToReceived,
		subject:  "",
		sender:   sendNotificationEmpty},

	join(models.RequestStatusDelivered, models.RequestStatusAccepted): sender{
		template: domain.MessageTemplateRequestFromDeliveredToAccepted,
		subject:  "Email.Subject.Request.FromDeliveredToAccepted",
		sender:   sendNotificationRequestFromDeliveredToAccepted},

	join(models.RequestStatusDelivered, models.RequestStatusCompleted): sender{
		template: domain.MessageTemplateRequestFromDeliveredToCompleted,
		subject:  "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
		sender:   sendNotificationRequestFromAcceptedOrDeliveredToCompleted},

	join(models.RequestStatusOpen, models.RequestStatusAccepted): sender{
		template: domain.MessageTemplateRequestFromOpenToAccepted,
		subject:  "Email.Subject.Request.FromOpenToAccepted",
		sender:   sendNotificationRequestFromOpenToAccepted},

	join(models.RequestStatusReceived, models.RequestStatusCompleted): sender{
		template: domain.MessageTemplateRequestFromReceivedToCompleted,
		subject:  "",
		sender:   sendNotificationEmpty},
}

func requestStatusUpdatedNotifications(post models.Request, eData models.RequestStatusEventData) {

	fromStatusTo := join(eData.OldStatus, eData.NewStatus)
	sender, ok := statusSenders[fromStatusTo]

	if !ok {
		domain.ErrLogger.Printf("unexpected status transition '%s'", fromStatusTo)
		return
	}

	params := senderParams{
		template:   notifications.GetEmailTemplate(sender.template),
		subject:    sender.subject,
		post:       post,
		pEventData: eData,
	}

	sender.sender(params)
}

func sendNewPostNotifications(post models.Request, users models.Users) {
	for i, user := range users {
		if !user.WantsRequestNotification(post) {
			continue
		}

		if err := sendNewPostNotification(user, post); err != nil {
			domain.ErrLogger.Printf("error sending post created notification (%d of %d), %s",
				i, len(users), err)
		}
	}
}

func sendNewPostNotification(user models.User, post models.Request) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	receiver, err := post.GetCreator()
	if err != nil {
		return err
	}
	receiverNickname := ""
	if receiver != nil {
		receiverNickname = receiver.Nickname
	}

	postDestination := ""
	if dest, err := post.GetDestination(); err == nil && dest != nil {
		postDestination = dest.Description
	}

	msg := notifications.Message{
		Subject: domain.GetTranslatedSubject(user.GetLanguagePreference(),
			"Email.Subject.NewRequest", map[string]string{}),
		Template:  domain.MessageTemplateNewRequest,
		ToName:    user.GetRealName(),
		ToEmail:   user.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Data: map[string]interface{}{
			"appName":          domain.Env.AppName,
			"uiURL":            domain.Env.UIURL,
			"postURL":          domain.GetPostUIURL(post.UUID.String()),
			"postTitle":        domain.Truncate(post.Title, "...", 16),
			"receiverNickname": receiverNickname,
			"postDescription":  post.Description,
			"postDestination":  postDestination,
		},
	}
	return notifications.Send(msg)
}

func sendPotentialProviderCreatedNotification(providerNickname string, requester models.User, post models.Request) error {
	template := domain.MessageTemplatePotentialProviderCreated
	msg := getPotentialProviderMessageForReceiver(requester, providerNickname, template, post)
	msg.Subject = domain.GetTranslatedSubject(requester.GetLanguagePreference(),
		"Email.Subject.Request.NewOffer", map[string]string{})

	return notifications.Send(msg)
}

func sendPotentialProviderSelfDestroyedNotification(providerNickname string, requester models.User, post models.Request) error {
	template := domain.MessageTemplatePotentialProviderSelfDestroyed
	msg := getPotentialProviderMessageForReceiver(requester, providerNickname, template, post)
	msg.Subject = domain.GetTranslatedSubject(requester.GetLanguagePreference(),
		"Email.Subject.Request.OfferRetracted", map[string]string{})
	return notifications.Send(msg)
}

func sendPotentialProviderRejectedNotification(provider models.User, requester string, post models.Request) error {
	msg := notifications.Message{
		Subject: domain.GetTranslatedSubject(provider.GetLanguagePreference(),
			"Email.Subject.Request.OfferRejected", map[string]string{}),
		Template:  domain.MessageTemplatePotentialProviderRejected,
		ToName:    provider.GetRealName(),
		ToEmail:   provider.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Data: map[string]interface{}{
			"appName":          domain.Env.AppName,
			"uiURL":            domain.Env.UIURL,
			"postURL":          domain.GetPostUIURL(post.UUID.String()),
			"postTitle":        domain.Truncate(post.Title, "...", 16),
			"receiverNickname": requester,
		},
	}
	return notifications.Send(msg)
}
