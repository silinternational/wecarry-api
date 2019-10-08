package models

import (
	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"time"
)

// For new listener functions, register them at the end of the
// file in the RegisterListeners function

const (
	UserAccessTokensCleanupDelayMinutes = 480
)

var UserAccessTokensNextCleanupTime time.Time

func userCreated(e events.Event) {
	if e.Kind != ApiUserCreated {
		return
	}

	domain.Logger.Printf("%s User Created ... %s", domain.GetCurrentTime(), e.Message)
}

func UserAccessTokensCleanup(e events.Event) {
	if e.Kind != ApiAuthUserLoggedIn {
		return
	}

	now := time.Now()
	if !now.After(UserAccessTokensNextCleanupTime) {
		return
	}

	UserAccessTokensNextCleanupTime = now.Add(time.Duration(time.Minute * UserAccessTokensCleanupDelayMinutes))

	deleted, err := UserAccessTokensDeleteExpired()
	if err != nil {
		domain.ErrLogger.Print("Last error deleting expired user access tokens during cleanup ... " + err.Error())
	}

	domain.Logger.Printf("Deleted %v expired user access tokens during cleanup", deleted)
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	var name string
	var err error

	name = "user-created"
	_, err = events.NamedListen(name, userCreated)
	if err != nil {
		domain.ErrLogger.Print("Failed registering listener: " + name)
	}

	name = "trigger-user-access-tokens-cleanup"
	_, err = events.NamedListen(name, UserAccessTokensCleanup)
	if err != nil {
		domain.ErrLogger.Print("Failed registering listener: " + name)
	}

}
