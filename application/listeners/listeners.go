package listeners

import (
	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"time"
)

const (
	UserAccessTokensCleanupDelayMinutes = 480
)

var UserAccessTokensNextCleanupTime time.Time

//
// Register new listener functions here
//
var apiListeners = []apiListener{
	{
		name:     "user-created",
		listener: userCreated,
	},
	{
		name:     "trigger-user-access-tokens-cleanup",
		listener: userAccessTokensCleanup,
	},
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	for _, a := range apiListeners {
		_, err := events.NamedListen(a.name, a.listener)
		if err != nil {
			domain.ErrLogger.Print("Failed registering listener: " + a.name)
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
		domain.ErrLogger.Print("Last error deleting expired user access tokens during cleanup ... " + err.Error())
	}

	domain.Logger.Printf("Deleted %v expired user access tokens during cleanup", deleted)
}

func userCreated(e events.Event) {
	if e.Kind != domain.EventApiUserCreated {
		return
	}

	domain.Logger.Printf("%s User Created ... %s", domain.GetCurrentTime(), e.Message)
}

type apiListener struct {
	name     string
	listener func(events.Event)
}
