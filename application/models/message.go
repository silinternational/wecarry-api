package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/events"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo/genny/build/_fixtures/coke/models"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Message struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Uuid      uuid.UUID `json:"uuid" db:"uuid"`
	ThreadID  int       `json:"thread_id" db:"thread_id"`
	SentByID  int       `json:"sent_by_id" db:"sent_by_id"`
	Content   string    `json:"content" db:"content"`
	Thread    Thread    `belongs_to:"threads"`
	SentBy    User      `belongs_to:"users"`
}

// MessageCreatedEventData holds data needed by the Message Created event listener
type MessageCreatedEventData struct {
	MessageCreatorNickName string
	MessageCreatorEmail    string
	MessageContent         string
	PostUUID               string
	PostTitle              string
	ThreadUUID             string
	MessageRecipient       []struct{ Nickname, Email string }
}

// String is not required by pop and may be deleted
func (m Message) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Messages is not required by pop and may be deleted
type Messages []Message

// String is not required by pop and may be deleted
func (m Messages) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *Message) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: m.Uuid, Name: "Uuid"},
		&validators.IntIsPresent{Field: m.ThreadID, Name: "ThreadID"},
		&validators.IntIsPresent{Field: m.SentByID, Name: "SentByID"},
		&validators.StringIsPresent{Field: m.Content, Name: "Content"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *Message) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *Message) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// GetSender finds and returns the User that is the Sender of this Message
func (m *Message) GetSender(requestFields []string) (*User, error) {
	sender := User{}
	if err := models.DB.Select(requestFields...).Find(&sender, m.SentByID); err != nil {
		err = fmt.Errorf("error finding message sentBy user with id %v ... %v", m.SentByID, err)
		return nil, err
	}
	return &sender, nil
}

// GetThread finds and returns the Thread that this Message is attached to
func (m *Message) GetThread(requestFields []string) (*Thread, error) {
	thread := Thread{}
	if err := models.DB.Select(requestFields...).Find(&thread, m.ThreadID); err != nil {
		err = fmt.Errorf("error finding message thread id %v ... %v", m.ThreadID, err)
		return nil, err
	}
	return &thread, nil
}

// Create a new message. Sends an `EventApiMessageCreated` event.
func (m *Message) Create() error {
	if err := DB.Create(m); err != nil {
		return err
	}

	if err := DB.Load(m, "SentBy", "Thread"); err != nil {
		return err
	}

	if err := DB.Load(&m.Thread, "Participants", "Post"); err != nil {
		return err
	}

	eventData := MessageCreatedEventData{
		MessageCreatorNickName: m.SentBy.Nickname,
		MessageCreatorEmail:    m.SentBy.Email,
		MessageContent:         m.Content,
		PostUUID:               m.Thread.Post.Uuid.String(),
		PostTitle:              m.Thread.Post.Title,
		ThreadUUID:             m.Thread.Uuid.String(),
	}

	for _, tp := range m.Thread.Participants {
		if tp.ID == m.SentBy.ID {
			continue
		}
		eventData.MessageRecipient = append(eventData.MessageRecipient,
			struct{ Nickname, Email string }{Nickname: tp.Nickname, Email: tp.Email})
	}

	e := events.Event{
		Kind:    domain.EventApiMessageCreated,
		Message: "New Message from " + m.SentBy.Nickname,
		Payload: events.Payload{"eventData": eventData},
	}

	emitEvent(e)

	//
	//uiUrl := envy.Get(domain.UIURLEnv, "")
	//for _, tp := range m.Thread.Participants {
	//	if tp.ID == m.SentBy.ID {
	//		continue
	//	}
	//	e := events.Event{
	//		Kind:    domain.EventApiMessageCreated,
	//		Message: "New Message from " + m.SentBy.Nickname,
	//		Payload: events.Payload{
	//			"toName":         tp.Nickname,
	//			"toEmail":        tp.Email,
	//			"toPhone":        "",
	//			"fromName":       m.SentBy.Nickname,
	//			"fromEmail":      m.SentBy.Email,
	//			"fromPhone":      "",
	//			"postURL":        uiUrl + "/#/requests/" + m.Thread.Post.Uuid.String(),
	//			"postTitle":      m.Thread.Post.Title,
	//			"messageContent": m.Content,
	//			"threadURL":      uiUrl + "/#/messages/" + m.Thread.Uuid.String(),
	//		},
	//	}
	//	emitEvent(e)
	//}
	return nil
}
