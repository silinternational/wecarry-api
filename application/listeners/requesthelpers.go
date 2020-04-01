package listeners

import (
	"errors"
	"fmt"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const requestTitleKey = "requestTitle"

type requestUser struct {
	Language string
	Nickname string
	Email    string
}

type requestUsers struct {
	Receiver requestUser
	Provider requestUser
}

// getRequestUsers returns up to two entries for the Request Requester and
// Request Provider assuming their email is not blank.
func getRequestUsers(request models.Request) requestUsers {

	receiver, _ := request.GetCreator()
	provider, _ := request.GetProvider()

	var recipients requestUsers

	if receiver != nil {
		recipients.Receiver = requestUser{
			Language: receiver.GetLanguagePreference(),
			Nickname: receiver.Nickname,
			Email:    receiver.Email,
		}
	}

	if provider != nil {
		recipients.Provider = requestUser{
			Language: provider.GetLanguagePreference(),
			Nickname: provider.Nickname,
			Email:    provider.Email,
		}
	}

	return recipients
}

func getMessageForProvider(requestUsers requestUsers, request models.Request, template string) notifications.Message {
	data := map[string]interface{}{
		"uiURL":              domain.Env.UIURL,
		"appName":            domain.Env.AppName,
		"requestURL":         domain.GetRequestUIURL(request.UUID.String()),
		"requestTitle":       domain.Truncate(request.Title, "...", 16),
		"requestDescription": request.Description,
		"receiverNickname":   requestUsers.Receiver.Nickname,
		"receiverEmail":      requestUsers.Receiver.Email,
	}

	return notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    requestUsers.Provider.Nickname,
		ToEmail:   requestUsers.Provider.Email,
		FromEmail: domain.EmailFromAddress(nil),
	}
}

func getMessageForReceiver(requestUsers requestUsers, request models.Request, template string) notifications.Message {
	data := map[string]interface{}{
		"uiURL":              domain.Env.UIURL,
		"appName":            domain.Env.AppName,
		"requestURL":         domain.GetRequestUIURL(request.UUID.String()),
		"requestTitle":       domain.Truncate(request.Title, "...", 16),
		"requestDescription": request.Description,
		"providerNickname":   requestUsers.Provider.Nickname,
		"providerEmail":      requestUsers.Provider.Email,
	}

	return notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    requestUsers.Receiver.Nickname,
		ToEmail:   requestUsers.Receiver.Email,
		FromEmail: domain.EmailFromAddress(nil),
	}
}

func getPotentialProviderMessageForReceiver(
	requester models.User, providerNickname, template string, request models.Request) notifications.Message {
	data := map[string]interface{}{
		"appName":          domain.Env.AppName,
		"uiURL":            domain.Env.UIURL,
		"requestURL":       domain.GetRequestUIURL(request.UUID.String()),
		"requestTitle":     domain.Truncate(request.Title, "...", 16),
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
	request := params.request
	template := params.template
	requestUsers := getRequestUsers(request)

	if requestUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForProvider(requestUsers, request, template)
	msg.Subject = domain.GetTranslatedSubject(requestUsers.Provider.Language, params.subject,
		map[string]string{requestTitleKey: request.Title})

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestToReceiver(params senderParams) {
	request := params.request
	template := params.template

	requestUsers := getRequestUsers(request)

	if requestUsers.Provider.Nickname == "" {
		domain.ErrLogger.Printf("error preparing '%s' notification - no provider", template)
		return
	}

	msg := getMessageForReceiver(requestUsers, request, template)
	msg.Subject = domain.GetTranslatedSubject(requestUsers.Receiver.Language, params.subject,
		map[string]string{requestTitleKey: request.Title})

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification, %s", template, err)
	}
}

func sendNotificationRequestFromAcceptedToDelivered(params senderParams) {
	sendNotificationRequestToReceiver(params)
}

func sendNotificationRequestFromAcceptedToOpen(params senderParams) {
	request := params.request
	template := params.template
	eData := params.pEventData

	requestUsers := getRequestUsers(request)

	oldProvider := models.User{}
	if err := oldProvider.FindByID(eData.OldProviderID); err != nil {
		domain.ErrLogger.Printf("error preparing '%s' notification for old provider id, %v ... %v",
			template, eData.OldProviderID, err)
		return
	}

	msg := getMessageForProvider(requestUsers, request, template)

	msg.ToName = oldProvider.GetRealName()
	msg.ToEmail = oldProvider.Email
	msg.Subject = domain.GetTranslatedSubject(oldProvider.GetLanguagePreference(), params.subject,
		map[string]string{requestTitleKey: request.Title})

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

func sendRejectionToPotentialProvider(potentialProvider models.User, request models.Request) {
	template := domain.MessageTemplatePotentialProviderRejected
	ppNickname := potentialProvider.Nickname
	ppEmail := potentialProvider.Email

	if ppNickname == "" {
		ppNickname = "Unknown User"
		ppEmail = "Missing Email"
	}

	receiver, err := request.GetCreator()
	if err != nil {
		domain.ErrLogger.Printf("error getting Request Receiver for email data, %s", err)
	}

	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"appName":          domain.Env.AppName,
		"requestURL":       domain.GetRequestUIURL(request.UUID.String()),
		"requestTitle":     domain.Truncate(request.Title, "...", 16),
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
			map[string]string{requestTitleKey: request.Title}),
	}

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to rejected potentialProvider, %s", template, err)
	}
}

func sendNotificationRequestFromOpenToAccepted(params senderParams) {
	sendNotificationRequestToProvider(params)
	request := params.request

	var providers models.PotentialProviders
	users, err := providers.FindUsersByRequestID(request, models.User{})
	if err != nil {
		domain.ErrLogger.Printf("error finding rejected potential providers for request id, %v ... %v",
			request.ID, err)
	}

	for _, u := range users {
		if u.ID != request.ProviderID.Int {
			sendRejectionToPotentialProvider(u, request)
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
	request    models.Request
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

	join(models.RequestStatusAccepted, models.RequestStatusDelivered): sender{
		template: domain.MessageTemplateRequestFromAcceptedToDelivered,
		subject:  "Email.Subject.Request.FromAcceptedToDelivered",
		sender:   sendNotificationRequestFromAcceptedToDelivered},

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

func requestStatusUpdatedNotifications(request models.Request, eData models.RequestStatusEventData) {

	fromStatusTo := join(eData.OldStatus, eData.NewStatus)
	sender, ok := statusSenders[fromStatusTo]

	if !ok {
		domain.ErrLogger.Printf("unexpected status transition '%s'", fromStatusTo)
		return
	}

	params := senderParams{
		template:   notifications.GetEmailTemplate(sender.template),
		subject:    sender.subject,
		request:    request,
		pEventData: eData,
	}

	sender.sender(params)
}

func sendNewRequestNotifications(request models.Request, users models.Users) {
	for i, user := range users {
		if !user.WantsRequestNotification(request) {
			continue
		}

		if err := sendNewRequestNotification(user, request); err != nil {
			domain.ErrLogger.Printf("error sending request created notification (%d of %d), %s",
				i, len(users), err)
		}
	}
}

func sendNewRequestNotification(user models.User, request models.Request) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	receiver, err := request.GetCreator()
	if err != nil {
		return err
	}
	receiverNickname := ""
	if receiver != nil {
		receiverNickname = receiver.Nickname
	}

	requestDestination := ""
	if dest, err := request.GetDestination(); err == nil && dest != nil {
		requestDestination = dest.Description
	}

	msg := notifications.Message{
		Subject: domain.GetTranslatedSubject(user.GetLanguagePreference(),
			"Email.Subject.NewRequest", map[string]string{}),
		Template:  domain.MessageTemplateNewRequest,
		ToName:    user.GetRealName(),
		ToEmail:   user.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Data: map[string]interface{}{
			"appName":            domain.Env.AppName,
			"uiURL":              domain.Env.UIURL,
			"requestURL":         domain.GetRequestUIURL(request.UUID.String()),
			"requestTitle":       domain.Truncate(request.Title, "...", 16),
			"receiverNickname":   receiverNickname,
			"requestDescription": request.Description,
			"requestDestination": requestDestination,
		},
	}
	return notifications.Send(msg)
}

func sendPotentialProviderCreatedNotification(providerNickname string, requester models.User, request models.Request) error {
	template := domain.MessageTemplatePotentialProviderCreated
	msg := getPotentialProviderMessageForReceiver(requester, providerNickname, template, request)
	msg.Subject = domain.GetTranslatedSubject(requester.GetLanguagePreference(),
		"Email.Subject.Request.NewOffer", map[string]string{})

	return notifications.Send(msg)
}

func sendPotentialProviderSelfDestroyedNotification(providerNickname string, requester models.User, request models.Request) error {
	template := domain.MessageTemplatePotentialProviderSelfDestroyed
	msg := getPotentialProviderMessageForReceiver(requester, providerNickname, template, request)
	msg.Subject = domain.GetTranslatedSubject(requester.GetLanguagePreference(),
		"Email.Subject.Request.OfferRetracted", map[string]string{})
	return notifications.Send(msg)
}

func sendPotentialProviderRejectedNotification(provider models.User, requester string, request models.Request) error {
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
			"requestURL":       domain.GetRequestUIURL(request.UUID.String()),
			"requestTitle":     domain.Truncate(request.Title, "...", 16),
			"receiverNickname": requester,
		},
	}
	return notifications.Send(msg)
}
