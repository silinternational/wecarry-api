package listeners

import (
	"errors"

	"github.com/gobuffalo/events"

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

	domain.EventApiPostStatusUpdated: {
		{
			name:     "post-status-updated-notification",
			listener: sendPostStatusUpdatedNotification,
		},
	},

	domain.EventApiPostCreated: {
		{
			name:     "post-created-notification",
			listener: sendPostCreatedNotifications,
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

func sendPostStatusUpdatedNotification(e events.Event) {
	if e.Kind != domain.EventApiPostStatusUpdated {
		return
	}

	pEData, ok := e.Payload["eventData"].(models.PostStatusEventData)
	if !ok {
		domain.ErrLogger.Printf("unable to parse Post Status Updated event payload")
		return
	}

	pid := pEData.PostID

	post := models.Post{}
	if err := post.FindByID(pid); err != nil {
		domain.ErrLogger.Printf("unable to find post from event with id %v ... %s", pid, err)
	}

	if post.Type != models.PostTypeRequest {
		return
	}

	requestStatusUpdatedNotifications(post, pEData)
}

func sendPostCreatedNotifications(e events.Event) {
	if e.Kind != domain.EventApiPostCreated {
		return
	}

	eventData, ok := e.Payload["eventData"].(models.PostCreatedEventData)
	if !ok {
		domain.ErrLogger.Printf("Post Created event payload incorrect type: %T", e.Payload["eventData"])
		return
	}

	var post models.Post
	if err := post.FindByID(eventData.PostID); err != nil {
		domain.ErrLogger.Printf("unable to find post %d from post-created event, %s", eventData.PostID, err)
	}

	users, err := post.GetAudience()
	if err != nil {
		domain.ErrLogger.Printf("unable to get post audience in event listener: %s", err.Error())
		return
	}

	sendNewPostNotifications(post, users)
}

func sendNewUserWelcome(user models.User) error {
	if user.Email == "" {
		return errors.New("'To' email address is required")
	}

	language := user.GetLanguagePreference()
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
