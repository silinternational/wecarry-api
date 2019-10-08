package eventers

import (
	"github.com/gobuffalo/events"
	"log"
)

const (
	ApiUserCreated = "api:user:created"
)

/////////////////////////////////////////////////////////////////////////////////////////
//
//  This file is for all the event emitting functions used throughout the app
//
//    Here is an example of how to use them in a model  ...
//
//    eventers.UserCreated("Nickname: " + u.Nickname + "  Uuid: " + u.Uuid.String())
//
/////////////////////////////////////////////////////////////////////////////////////////

// UserCreated emits an ApiUserCreated event
func UserCreated(message string) {
	e := events.Event{
		Kind:    ApiUserCreated,
		Message: message,
	}
	if err := events.Emit(e); err != nil {
		log.Printf("error emitting event %s ... %v", e.Kind, err)
	}
}
