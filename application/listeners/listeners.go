package listeners

import (
	"time"

	"github.com/silinternational/wecarry-api/job"

	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

const (
	UserAccessTokensCleanupDelayMinutes = 480
)

var UserAccessTokensNextCleanupTime time.Time

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

	id, ok := e.Payload["id"].(int)
	if !ok {
		domain.ErrLogger.Print("sendNewMessageNotification: unable to read message ID from event payload")
		return
	}

	if err := job.Submit("new_message", map[string]interface{}{"message_id": id}); err != nil {
		domain.ErrLogger.Printf("error starting 'New Message' job, %s", err)
	}
}
