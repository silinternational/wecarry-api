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

// getPostUsers returns up to two entries for the Post Requester and
// Post Provider assuming their email is not blank.
func getPostUsers(post models.Post) postUsers {

	receiver, _ := post.GetReceiver()
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

func getMessageForProvider(postUsers postUsers, post models.Post, template string) notifications.Message {
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

func getMessageForReceiver(postUsers postUsers, post models.Post, template string) notifications.Message {
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

func sendNotificationRequestFromAcceptedOrCommittedToDelivered(params senderParams) {
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

func sendRejectionToCommitter(committer models.User, post models.Post) {
	template := domain.MessageTemplateCommittmentNotAccepted
	committerNickname := committer.Nickname
	committerEmail := committer.Email

	if committerNickname == "" {
		committerNickname = "Unknown User"
		committerEmail = "Missing Email"
	}

	data := map[string]interface{}{
		"uiURL":             domain.Env.UIURL,
		"appName":           domain.Env.AppName,
		"postURL":           domain.GetPostUIURL(post.UUID.String()),
		"postTitle":         domain.Truncate(post.Title, "...", 16),
		"committerNickname": committerNickname,
		"committerEmail":    committerEmail,
	}

	subject := "Email.Subject.Request.CommittmentRejected"

	msg := notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    committer.GetRealName(),
		ToEmail:   committerEmail,
		FromEmail: domain.EmailFromAddress(nil),
		Subject: domain.GetTranslatedSubject(committer.GetLanguagePreference(), subject,
			map[string]string{postTitleKey: post.Title}),
	}

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to rejected committer, %s", template, err)
	}
}

func sendNotificationRequestFromOpenToAccepted(params senderParams) {
	sendNotificationRequestToProvider(params)

	post := params.post

	var committers models.RequestCommitters
	if err := committers.FindByPostID(post.ID); err != nil {
		domain.ErrLogger.Printf("error finding rejected providers for post id, %v ... %v",
			post.ID, err)
	}

	for _, c := range committers {
		sendRejectionToCommitter(c.User, post)
	}

}

func sendNotificationRequestFromCommittedToCommitted(params senderParams) {
	sendNotificationRequestToReceiver(params)

}

// Until we have status auditing history, we don't know who reverted the Post to `open` status.
//  So, tell both the receiver and provider about it.
func sendNotificationRequestFromCommittedToOpen(params senderParams) {
	post := params.post
	template := params.template
	eData := params.pEventData

	postUsers := getPostUsers(post)

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

	// First notify receiver
	data := map[string]interface{}{
		"uiURL":            domain.Env.UIURL,
		"appName":          domain.Env.AppName,
		"postURL":          domain.GetPostUIURL(post.UUID.String()),
		"postTitle":        domain.Truncate(post.Title, "...", 16),
		"providerNickname": providerNickname,
		"providerEmail":    providerEmail,
		"receiverNickname": postUsers.Receiver.Nickname,
		"receiverEmail":    postUsers.Receiver.Email,
	}

	msg := notifications.Message{
		Template:  template,
		Data:      data,
		ToName:    postUsers.Receiver.Nickname,
		ToEmail:   postUsers.Receiver.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Subject: domain.GetTranslatedSubject(postUsers.Receiver.Language, params.subject,
			map[string]string{postTitleKey: post.Title}),
	}

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to old provider, %s", template, err)
	}

	// Now notify the old provider
	if oldProvider.Nickname == "" {
		return
	}

	msg.ToName = oldProvider.GetRealName()
	msg.ToEmail = oldProvider.Email
	msg.Subject = domain.GetTranslatedSubject(oldProvider.GetLanguagePreference(), params.subject,
		map[string]string{postTitleKey: post.Title})

	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending '%s' notification to requester, %s", template, err)
	}
}

func sendNotificationRequestFromCommittedToRemoved(params senderParams) {
	sendNotificationRequestToProvider(params)
}

func sendNotificationRequestFromDeliveredToAccepted(params senderParams) {
	sendNotificationRequestToReceiver(params)
}

func sendNotificationRequestFromDeliveredToCommitted(params senderParams) {
	sendNotificationRequestToReceiver(params)
}

func sendNotificationRequestFromOpenToCommitted(params senderParams) {
	sendNotificationRequestToReceiver(params)
}

func sendNotificationRequestFromCompletedToAcceptedOrDelivered(params senderParams) {
	sendNotificationRequestToProvider(params)
}

func sendNotificationEmpty(params senderParams) {
	domain.ErrLogger.Print("Notification not implemented yet for " + params.template)
}

type senderParams struct {
	template   string
	subject    string
	post       models.Post
	pEventData models.PostStatusEventData
}

type sender struct {
	template string
	subject  string
	sender   func(senderParams) // string, string, models.Post, models.PostStatusEventData)
}

func join(s1, s2 models.PostStatus) string {
	return fmt.Sprintf("%s-%s", s1, s2)
}

var statusSenders = map[string]sender{
	join(models.PostStatusAccepted, models.PostStatusCompleted): sender{
		template: domain.MessageTemplateRequestFromAcceptedToCompleted,
		subject:  "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
		sender:   sendNotificationRequestFromAcceptedOrDeliveredToCompleted},

	join(models.PostStatusAccepted, models.PostStatusOpen): sender{
		template: domain.MessageTemplateRequestFromAcceptedToOpen,
		subject:  "Email.Subject.Request.FromAcceptedToOpen",
		sender:   sendNotificationRequestFromAcceptedToOpen},

	join(models.PostStatusAccepted, models.PostStatusReceived): sender{
		template: domain.MessageTemplateRequestFromAcceptedToCompleted,
		subject:  "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
		sender:   sendNotificationRequestFromAcceptedOrDeliveredToCompleted},

	join(models.PostStatusAccepted, models.PostStatusRemoved): sender{
		template: domain.MessageTemplateRequestFromAcceptedToRemoved,
		subject:  "Email.Subject.Request.FromAcceptedToRemoved",
		sender:   sendNotificationRequestFromAcceptedToRemoved},

	join(models.PostStatusCompleted, models.PostStatusAccepted): sender{
		template: domain.MessageTemplateRequestFromCompletedToAccepted,
		subject:  "Email.Subject.Request.FromCompletedToAcceptedOrDelivered",
		sender:   sendNotificationRequestFromCompletedToAcceptedOrDelivered},

	join(models.PostStatusCompleted, models.PostStatusDelivered): sender{
		template: domain.MessageTemplateRequestFromCompletedToDelivered,
		subject:  "Email.Subject.Request.FromCompletedToAcceptedOrDelivered",
		sender:   sendNotificationRequestFromCompletedToAcceptedOrDelivered},

	join(models.PostStatusCompleted, models.PostStatusReceived): sender{
		template: domain.MessageTemplateRequestFromCompletedToReceived,
		subject:  "",
		sender:   sendNotificationEmpty},

	join(models.PostStatusDelivered, models.PostStatusAccepted): sender{
		template: domain.MessageTemplateRequestFromDeliveredToAccepted,
		subject:  "Email.Subject.Request.FromDeliveredToAccepted",
		sender:   sendNotificationRequestFromDeliveredToAccepted},

	join(models.PostStatusDelivered, models.PostStatusCompleted): sender{
		template: domain.MessageTemplateRequestFromDeliveredToCompleted,
		subject:  "Email.Subject.Request.FromAcceptedOrDeliveredToCompleted",
		sender:   sendNotificationRequestFromAcceptedOrDeliveredToCompleted},

	join(models.PostStatusOpen, models.PostStatusAccepted): sender{
		template: domain.MessageTemplateRequestFromOpenToAccepted,
		subject:  "Email.Subject.Request.FromOpenToAccepted",
		sender:   sendNotificationRequestFromOpenToAccepted},

	join(models.PostStatusReceived, models.PostStatusCompleted): sender{
		template: domain.MessageTemplateRequestFromReceivedToCompleted,
		subject:  "",
		sender:   sendNotificationEmpty},
}

func requestStatusUpdatedNotifications(post models.Post, eData models.PostStatusEventData) {

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

	receiver, err := post.GetReceiver()
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
		Template:  newPostTemplates[post.Type.String()],
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
