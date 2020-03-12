package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/domain"
)

type Message struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UUID      uuid.UUID `json:"uuid" db:"uuid"`
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
	MessageRecipients      []struct{ Nickname, Email string }
}

// String can be helpful for serializing the model
func (m Message) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Messages is merely for convenience and brevity
type Messages []Message

// String can be helpful for serializing the model
func (m Messages) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (m *Message) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: m.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: m.ThreadID, Name: "ThreadID"},
		&validators.IntIsPresent{Field: m.SentByID, Name: "SentByID"},
		&validators.StringIsPresent{Field: m.Content, Name: "Content"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (m *Message) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (m *Message) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// AfterCreate updates the LastViewedAt value on the associated ThreadParticipant and the UpdatedAt on the Thread to
// the current time. It also ensures the associated ThreadParticipant records exist, and emits an EventApiMessageCreated
// event.
func (m *Message) AfterCreate(tx *pop.Connection) error {

	thread, err := m.GetThread()
	if err != nil {
		return errors.New("error getting message's Thread ... " + err.Error())
	}

	post, err := thread.GetPost()
	if err != nil {
		return errors.New("error getting message's Post ... " + err.Error())
	}

	// Ensure a matching threadparticipant exists
	if err := thread.ensureParticipants(*post, m.SentByID); err != nil {
		return err
	}

	threadP := ThreadParticipant{}

	if err := threadP.FindByThreadIDAndUserID(m.ThreadID, m.SentByID); err != nil {
		domain.ErrLogger.Printf("aftercreate new message %s", err.Error())
		return nil
	}

	if err := threadP.UpdateLastViewedAt(time.Now()); err != nil {
		domain.ErrLogger.Printf("aftercreate new message %s", err.Error())
		return nil
	}

	// Touch the "updatedAt" field on the thread so thread lists can easily be sorted by last activity
	if err := DB.Load(m, "Thread"); err == nil {
		if err = m.Thread.Update(); err != nil {
			domain.Logger.Print("failed to save thread on message create,", err.Error())
		}
	}

	e := events.Event{
		Kind:    domain.EventApiMessageCreated,
		Message: "New Message Created",
		Payload: events.Payload{domain.ArgMessageID: m.ID},
	}

	emitEvent(e)

	return nil
}

// GetSender finds and returns the User that is the Sender of this Message
func (m *Message) GetSender() (*User, error) {
	sender := User{}
	if err := DB.Find(&sender, m.SentByID); err != nil {
		err = fmt.Errorf("error finding message sentBy user with id %v ... %v", m.SentByID, err)
		return nil, err
	}
	return &sender, nil
}

// GetThread finds and returns the Thread that this Message is attached to
func (m *Message) GetThread() (*Thread, error) {
	thread := Thread{}
	if err := DB.Find(&thread, m.ThreadID); err != nil {
		err = fmt.Errorf("error finding message thread id %v ... %v", m.ThreadID, err)
		return nil, err
	}
	return &thread, nil
}

// Create a new message if authorized.
func (m *Message) Create(ctx context.Context, postUUID string, threadUUID *string, content string) error {
	user := CurrentUser(ctx)

	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return errors.New("failed to find post, " + err.Error())
	}
	if !post.IsVisible(ctx, user) {
		return errors.New("user cannot create a message on post")
	}

	var thread Thread
	if threadUUID != nil {
		err := thread.FindByUUID(*threadUUID)
		if err != nil {
			return errors.New("failed to find thread, " + err.Error())
		}
		if thread.PostID != post.ID {
			return errors.New("thread is not valid for post")
		}
		if !thread.IsVisible(user.ID) {
			return errors.New("user cannot create a message on thread")
		}
	} else {
		err := thread.CreateWithParticipants(post, user)
		if err != nil {
			return errors.New("failed to create new thread on post, " + err.Error())
		}
	}

	m.Content = content
	m.ThreadID = thread.ID
	m.SentByID = user.ID
	if err := create(m); err != nil {
		return errors.New("failed to crate new message, " + err.Error())
	}

	return nil
}

// FindByID loads from DB the Message record identified by the given primary key
func (m *Message) FindByID(id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding message, invalid id")
	}

	var err error
	// Eager() with an empty argument list will load all fields, which is not what is intended here
	if len(eagerFields) > 0 {
		err = DB.Eager(eagerFields...).Find(m, id)
	} else {
		err = DB.Find(m, id)
	}

	if err != nil {
		return fmt.Errorf("error finding message by id, %s", err)
	}

	return DB.Find(m, id)
}

// findByUUID loads from DB the Message record identified by the given UUID
func (m *Message) findByUUID(id string) error {
	if id == "" {
		return errors.New("error: message uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", id).First(m); err != nil {
		return fmt.Errorf("error finding message by uuid: %s", err.Error())
	}

	return nil
}

// FindByUserAndUUID loads from DB the Message record identified by the given UUID, if the given user is allowed.
func (m *Message) FindByUserAndUUID(user User, id string) error {
	if err := m.findByUUID(id); err != nil {
		return err
	}

	if user.AdminRole == UserAdminRoleSuperAdmin {
		return nil
	}

	var tp ThreadParticipant
	if err := tp.FindByThreadIDAndUserID(m.ThreadID, user.ID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return fmt.Errorf("error finding threadParticipant record for message %s, %s", id, err)
		}
		return fmt.Errorf("user %s has insufficient permissions to read message %s", user.UUID, id)
	}

	return nil
}
