package listeners

import (
	"time"

	"github.com/silinternational/wecarry-api/notifications"

	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/job"
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
			listener: sendNewThreadMessageNotification,
		},
	},

	domain.EventApiPostStatusUpdated: []apiListener{
		{
			name:     "post-status-updated-notification",
			listener: sendPostStatusUpdatedNotification,
		},
	},

	domain.EventApiPostCreated: []apiListener{
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
				domain.ErrLogger.Print("Failed registering listener:", l.name, err)
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

func sendNewThreadMessageNotification(e events.Event) {
	if e.Kind != domain.EventApiMessageCreated {
		return
	}

	domain.Logger.Printf("%s Thread Message Created ... %s", domain.GetCurrentTime(), e.Message)

	id, ok := e.Payload[domain.ArgMessageID].(int)
	if !ok {
		domain.ErrLogger.Print("sendNewThreadMessageNotification: unable to read message ID from event payload")
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
		domain.ErrLogger.Print("unable to parse Post Status Updated event payload")
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
		domain.ErrLogger.Print("unable to parse Post Created event payload")
		return
	}

	post := models.Post{}
	if err := post.FindByID(eventData.PostID); err != nil {
		domain.ErrLogger.Printf("unable to find post %d from post-created event, %s", eventData.PostID, err)
		return
	}

	org, err := post.GetOrganization([]string{"id"})
	if err != nil {
		domain.ErrLogger.Print("unable to get post organization in event listener,", err.Error())
		return
	}

	users, err2 := org.GetUsers()
	if err2 != nil {
		domain.ErrLogger.Print("unable to get post organization user list in event listener,", err2.Error())
		return
	}

	for _, user := range users {
		if !user.WantsPostNotification(post) {
			continue
		}

		msg := notifications.Message{
			Template: domain.MessageTemplateNewRequest,
			ToName:   user.Nickname,
			ToEmail:  user.Email,
			Data: map[string]interface{}{
				"uiURL":     domain.Env.UIURL,
				"postURL":   domain.GetPostUIURL(post.Uuid.String()),
				"postTitle": post.Title,
			},
		}
		if err3 := notifications.Send(msg); err3 != nil {
			domain.ErrLogger.Printf("error sending post created notification, %s", err3)
		}
	}
}
