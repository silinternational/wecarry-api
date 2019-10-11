package listeners

import (
	"time"

	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/silinternational/wecarry-api/notifications"
)

const (
	UserAccessTokensCleanupDelayMinutes = 480
)

var UserAccessTokensNextCleanupTime time.Time

//
// Register new listener functions here.  Remember, though, that these groupings just
// describe what we want.  They don't make it happen this way. The listeners
// themselves still need to verify the event kind
//
var apiListeners = map[string][]apiListener{
	domain.EventApiUserCreated: []apiListener{
		{
			name:     "user-created",
			listener: userCreated,
		},
	},

	domain.EventApiAuthUserLoggedIn: []apiListener{
		{
			name:     "trigger-user-access-tokens-cleanup",
			listener: userAccessTokensCleanup,
		},
	},

	domain.EventApiMessageCreated: []apiListener{
		{
			name:     "send-new-message-notification",
			listener: sendNewMessageNotification,
		},
	},
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	for _, listeners := range apiListeners {
		for _, l := range listeners {
			_, err := events.NamedListen(l.name, l.listener)
			if err != nil {
				domain.ErrLogger.Print("Failed registering listener: " + l.name)
			}
		}
	}
}

func userAccessTokensCleanup(e events.Event) {
	if e.Kind != domain.EventApiAuthUserLoggedIn {
		return
	}

	now := time.Now()
	if !now.After(UserAccessTokensNextCleanupTime) {
		return
	}

	UserAccessTokensNextCleanupTime = now.Add(time.Duration(time.Minute * UserAccessTokensCleanupDelayMinutes))

	var uats models.UserAccessTokens
	deleted, err := uats.DeleteExpired()
	if err != nil {
		domain.ErrLogger.Printf("%s Last error deleting expired user access tokens during cleanup ... %v",
			domain.GetCurrentTime(), err)
	}

	domain.Logger.Printf("%s Deleted %v expired user access tokens during cleanup", domain.GetCurrentTime(), deleted)
}

func userCreated(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	domain.Logger.Printf("%s User Created ... %s", domain.GetCurrentTime(), e.Message)
}

func sendNewMessageNotification(e events.Event) {
	if e.Kind != domain.EventApiMessageCreated {
		return
	}
	domain.Logger.Printf("%s Message Created ... %s", domain.GetCurrentTime(), e.Message)

	data := map[string]interface{}{
		"postURL":        getPayload(e, "postURL"),
		"postTitle":      getPayload(e, "postTitle"),
		"messageContent": getPayload(e, "messageContent"),
		"sentByNickname": getPayload(e, "fromName"),
		"threadURL":      getPayload(e, "threadURL"),
	}

	msg := notifications.Message{
		Template:  domain.MessageTemplateNewMessage,
		Data:      data,
		FromName:  getPayload(e, "fromName").(string),
		FromEmail: getPayload(e, "fromEmail").(string),
		FromPhone: getPayload(e, "fromPhone").(string),
		ToName:    getPayload(e, "toName").(string),
		ToEmail:   getPayload(e, "toEmail").(string),
		ToPhone:   getPayload(e, "toPhone").(string),
	}
	if err := notifications.Send(msg); err != nil {
		domain.ErrLogger.Printf("error sending 'New Message' notification, %s", err)
	}
}

func getPayload(event events.Event, name string) interface{} {
	p, err := event.Payload.Pluck(name)
	if err != nil {
		domain.ErrLogger.Printf("error retrieving payload %s from %s event, %s", name, event.Kind, err)
	}
	return p
}

type apiListener struct {
	name     string
	listener func(events.Event)
}
