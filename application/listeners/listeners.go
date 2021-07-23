package listeners

import (
	"context"
	"errors"
	"time"

	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/cache"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
	"github.com/silinternational/wecarry-api/marketing"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

type apiListener struct {
	name     string
	listener func(events.Event)
}

//
// Register new listener functions here.  Remember, though, that these groupings just
// describe what we want.  They don't make it happen this way. The listeners
// themselves still need to verify the event kind
//
var apiListeners = map[string][]apiListener{
	domain.EventApiUserCreated: {
		{
			name:     "user-created-logger",
			listener: userCreatedLogger,
		},
		{
			name:     "user-created-send-welcome-message",
			listener: userCreatedSendWelcomeMessage,
		},
		{
			name:     "user-created-add-to-marketing-list",
			listener: userCreatedAddToMarketingList,
		},
	},

	domain.EventApiMessageCreated: {
		{
			name:     "send-new-message-notification",
			listener: sendNewThreadMessageNotification,
		},
	},

	domain.EventApiRequestStatusUpdated: {
		{
			name:     "request-status-updated-notification",
			listener: sendRequestStatusUpdatedNotification,
		},
	},

	domain.EventApiRequestCreated: {
		{
			name:     "request-created-notification",
			listener: sendRequestCreatedNotifications,
		},
		{
			name:     "request-created-cache",
			listener: cacheRequestCreatedListener,
		},
	},

	domain.EventApiRequestUpdated: {
		{
			name:     "request-updated-cache",
			listener: cacheRequestUpdatedListener,
		},
	},

	domain.EventApiPotentialProviderCreated: {
		{
			name:     "potentialprovider-created-notification",
			listener: potentialProviderCreated,
		},
	},

	domain.EventApiPotentialProviderSelfDestroyed: {
		{
			name:     "potentialprovider-self-destroyed-notification",
			listener: potentialProviderSelfDestroyed,
		},
	},

	domain.EventApiPotentialProviderRejected: {
		{
			name:     "potentialprovider-rejected-notification",
			listener: potentialProviderRejected,
		},
	},

	domain.EventApiMeetingInviteCreated: {
		{
			name:     "meeting-invite-created",
			listener: meetingInviteCreated,
		},
	},
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	for _, listeners := range apiListeners {
		for _, l := range listeners {
			_, err := events.NamedListen(l.name, l.listener)
			if err != nil {
				domain.ErrLogger.Printf("Failed registering listener: %s, err: %s", l.name, err.Error())
			}
		}
	}
}

func userCreatedLogger(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	domain.Logger.Printf("User Created: %s", e.Message)
}

func userCreatedSendWelcomeMessage(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	user, ok := e.Payload["user"].(*models.User)
	if !ok {
		domain.ErrLogger.Printf("Failed to get User from event payload for sending welcome message. Event message: %s", e.Message)
		return
	}

	if err := sendNewUserWelcome(*user); err != nil {
		domain.ErrLogger.Printf("Failed to send new user welcome to %s. Error: %s",
			user.UUID.String(), err)
	}
}

func userCreatedAddToMarketingList(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	user, ok := e.Payload["user"].(*models.User)
	if !ok {
		domain.ErrLogger.Printf(
			"Failed to get User from event payload for adding to marketing list. Event message: %s", e.Message)
		return
	}

	// ensure env vars are present
	if domain.Env.MailChimpAPIKey == "" {
		domain.ErrLogger.Printf("missing required env var for MAILCHIMP_API_KEY. need to add %s to list", user.Email)
		return
	}
	if domain.Env.MailChimpListID == "" {
		domain.ErrLogger.Printf("missing required env var for MAILCHIMP_LIST_ID. need to add %s to list", user.Email)
		return
	}
	if domain.Env.MailChimpUsername == "" {
		domain.ErrLogger.Printf("missing required env var for MAILCHIMP_USERNAME. need to add %s to list", user.Email)
		return
	}

	err := marketing.AddUserToList(*user, domain.Env.MailChimpAPIBaseURL, domain.Env.MailChimpListID,
		domain.Env.MailChimpUsername, domain.Env.MailChimpAPIKey)
	if err != nil {
		domain.ErrLogger.Printf("error calling marketing.AddUserToList when trying to add %s: %s",
			user.Email, err.Error())
	}
}

func sendNewThreadMessageNotification(e events.Event) {
	if e.Kind != domain.EventApiMessageCreated {
		return
	}

	domain.Logger.Printf("%s Thread Message Created ... %s", domain.GetCurrentTime(), e.Message)

	id, ok := e.Payload[domain.ArgMessageID].(int)
	if !ok {
		domain.ErrLogger.Printf("sendNewThreadMessageNotification: unable to read message ID from event payload")
		return
	}

	if err := job.SubmitDelayed(job.NewThreadMessage, domain.NewMessageNotificationDelay,
		map[string]interface{}{domain.ArgMessageID: id}); err != nil {
		domain.ErrLogger.Printf("error starting 'New Message' job, %s", err)
	}
}

func sendRequestStatusUpdatedNotification(e events.Event) {
	if e.Kind != domain.EventApiRequestStatusUpdated {
		return
	}

	pEData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.RequestStatusEventData)
	if !ok {
		domain.ErrLogger.Printf("unable to parse Request Status Updated event payload")
		return
	}

	pid := pEData.RequestID

	request := models.Request{}
	if err := request.FindByID(models.DB, pid); err != nil {
		domain.ErrLogger.Printf("unable to find request from event with id %v ... %s", pid, err)
	}

	requestStatusUpdatedNotifications(request, pEData)
}

func sendRequestCreatedNotifications(e events.Event) {
	if e.Kind != domain.EventApiRequestCreated {
		return
	}

	eventData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.RequestCreatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Request Created event payload incorrect type: %T", e.Payload[domain.EventPayloadKeyEventData])
		return
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from request-created event, %s", eventData.RequestID, err)
	}

	users, err := request.GetAudience(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to get request audience in event listener: %s", err.Error())
		return
	}

	sendNewRequestNotifications(request, users)
}

func cacheRequestCreatedListener(e events.Event) {
	if e.Kind != domain.EventApiRequestCreated {
		return
	}

	eventData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.RequestCreatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Request Created event payload incorrect type: %T", e.Payload[domain.EventPayloadKeyEventData])
		return
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from request-created event, %s", eventData.RequestID, err)
	}

	ctx := context.Background()
	cache.CacheRebuildOnNewRequest(ctx, request)
}

func cacheRequestUpdatedListener(e events.Event) {
	if e.Kind != domain.EventApiRequestUpdated {
		return
	}

	eventData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.RequestUpdatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Request Updated event payload incorrect type: %T", e.Payload[domain.EventPayloadKeyEventData])
		return
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from request-created event, %s", eventData.RequestID, err)
	}

	ctx := context.Background()
	cache.CacheRebuildOnChangedRequest(ctx, request)
}

func potentialProviderCreated(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderCreated {
		return
	}

	eventData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload[domain.EventPayloadKeyEventData])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(models.DB, eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from PotentialProvider event, %s", eventData.RequestID, err)
	}

	creator, err := request.Creator(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to find request %d creator from PotentialProvider event, %s",
			eventData.RequestID, err)
	}

	sendPotentialProviderCreatedNotification(potentialProvider.Nickname, creator, request)
}

func potentialProviderSelfDestroyed(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderSelfDestroyed {
		return
	}

	eventData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload[domain.EventPayloadKeyEventData])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(models.DB, eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from PotentialProvider event, %s", eventData.RequestID, err)
	}

	creator, err := request.Creator(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to find request %d creator from PotentialProvider event, %s",
			eventData.RequestID, err)
	}

	sendPotentialProviderSelfDestroyedNotification(potentialProvider.Nickname, creator, request)
}

func potentialProviderRejected(e events.Event) {
	if e.Kind != domain.EventApiPotentialProviderRejected {
		return
	}

	eventData, ok := e.Payload[domain.EventPayloadKeyEventData].(models.PotentialProviderEventData)
	if !ok {
		domain.ErrLogger.Printf("PotentialProvider event payload incorrect type: %T", e.Payload[domain.EventPayloadKeyEventData])
		return
	}

	var potentialProvider models.User
	if err := potentialProvider.FindByID(models.DB, eventData.UserID); err != nil {
		domain.ErrLogger.Printf("unable to find PotentialProvider User %d, %s", eventData.UserID, err)
	}

	var request models.Request
	if err := request.FindByID(models.DB, eventData.RequestID); err != nil {
		domain.ErrLogger.Printf("unable to find request %d from PotentialProvider event, %s", eventData.RequestID, err)
	}

	creator, err := request.Creator(models.DB)
	if err != nil {
		domain.ErrLogger.Printf("unable to find request %d creator from PotentialProvider event, %s",
			eventData.RequestID, err)
	}

	sendPotentialProviderRejectedNotification(potentialProvider, creator.Nickname, request)
}

func sendNewUserWelcome(user models.User) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	language := user.GetLanguagePreference(models.DB)
	subject := domain.GetTranslatedSubject(language, "Email.Subject.Welcome", map[string]string{})

	msg := notifications.Message{
		Template:  domain.MessageTemplateNewUserWelcome,
		ToName:    user.GetRealName(),
		ToEmail:   user.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Subject:   subject,
		Data: map[string]interface{}{
			"appName":      domain.Env.AppName,
			"uiURL":        domain.Env.UIURL,
			"supportEmail": domain.Env.SupportEmail,
			"userEmail":    user.Email,
			"firstName":    user.FirstName,
		},
	}
	return notifications.Send(msg)
}

func meetingInviteCreated(e events.Event) {
	if e.Kind != domain.EventApiMeetingInviteCreated {
		return
	}

	id, err := getID(e.Payload)
	if err != nil {
		domain.ErrLogger.Printf("meeting invite ID not found in payload, %s", err)
		return
	}

	foundMeetingInvite := false
	var invite models.MeetingInvite
	var findErr error

	i := 1
	for ; i <= domain.Env.ListenerMaxRetries; i++ {
		findErr = invite.FindByID(models.DB, id, "Meeting", "Inviter")
		if findErr == nil {
			foundMeetingInvite = true
			break
		}
		time.Sleep(getDelayDuration(i * i))
	}
	domain.Logger.Printf("listener meetingInviteCreated required %d retries with delay %d", i-1, domain.Env.ListenerDelayMilliseconds)

	if !foundMeetingInvite {
		domain.ErrLogger.Printf("failed to find MeetingInvite in meetingInviteCreated, %s", findErr)
		return
	}

	if err = sendMeetingInvite(invite); err != nil {
		domain.ErrLogger.Printf("unable to send invite %d in meetingInviteCreated event, %s", invite.ID, err)
	}
}

// sendMeetingInvite sends an email to the invitee. The MeetingInvite must have its Meeting and Inviter hydrated.
func sendMeetingInvite(invite models.MeetingInvite) error {
	if invite.Email == "" {
		return errors.New("'To' email address is required")
	}

	language := invite.Inviter.GetLanguagePreference(models.DB)
	subject := domain.GetTranslatedSubject(language, "Email.Subject.MeetingInvite",
		map[string]string{"MeetingName": invite.Meeting.Name})

	msg := notifications.Message{
		Template:  domain.MessageTemplateInvite,
		ToEmail:   invite.Email,
		FromEmail: domain.EmailFromAddress(nil),
		Subject:   subject,
		Data: map[string]interface{}{
			"appName":      domain.Env.AppName,
			"uiURL":        domain.Env.UIURL,
			"supportEmail": domain.Env.SupportEmail,
			"inviterName":  invite.Inviter.FirstName,
			"eventName":    invite.Meeting.Name,
			"inviteURL":    invite.InviteURL(),
		},
	}
	return notifications.Send(msg)
}

// getDelayDuration is a helper function to calculate delay in milliseconds before processing event
func getDelayDuration(multiplier int) time.Duration {
	return time.Duration(domain.Env.ListenerDelayMilliseconds) * time.Millisecond * time.Duration(multiplier)
}

func getID(p events.Payload) (int, error) {
	i, ok := p[domain.EventPayloadKeyId]
	if !ok {
		return 0, errors.New("id not in event payload")
	}

	id, ok := i.(int)
	if !ok {
		return 0, errors.New("ID is not an int")
	}

	return id, nil
}
