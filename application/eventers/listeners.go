package eventers

import (
	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"
	"log"
)

// For new listener functions, register them at the end of the
// file in the RegisterListeners function

func userCreated(e events.Event) {
	if e.Kind != ApiUserCreated {
		return
	}
	now := domain.GetCurrentTime()
	log.Printf("%s User Created ... %s", now, e.Message)
}

// RegisterListeners registers all the listeners to be used by the app
func RegisterListeners() {
	var name string
	var err error

	name = "user-created"
	_, err = events.NamedListen(name, userCreated)
	if err != nil {
		log.Print("Failed registering listener: " + name)
	}

}
