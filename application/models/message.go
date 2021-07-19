package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
)

type Message struct {
	ID        int       `json:"-" db:"id"`
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
	RequestUUID            string
	RequestTitle           string
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
	thread, err := m.GetThread(tx)
	if err != nil {
		return errors.New("error getting message's Thread ... " + err.Error())
	}

	err = thread.LoadRequest(tx)
	if err != nil {
		return errors.New("error getting message's Request ... " + err.Error())
	}

	// Ensure a matching threadparticipant exists
	if err := thread.ensureParticipants(tx, thread.Request, m.SentByID); err != nil {
		return err
	}

	threadP := ThreadParticipant{}

	if err := threadP.FindByThreadIDAndUserID(tx, m.ThreadID, m.SentByID); err != nil {
		domain.ErrLogger.Printf("aftercreate new message %s", err.Error())
		return nil
	}

	if err := threadP.UpdateLastViewedAt(tx, time.Now()); err != nil {
		domain.ErrLogger.Printf("aftercreate new message %s", err.Error())
		return nil
	}

	// Touch the "updatedAt" field on the thread so thread lists can easily be sorted by last activity
	if err := tx.Load(m, "Thread"); err == nil {
		if err = m.Thread.Update(tx); err != nil {
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
func (m *Message) GetSender(tx *pop.Connection) (*User, error) {
	sender := User{}
	if err := tx.Find(&sender, m.SentByID); err != nil {
		err = fmt.Errorf("error finding message sentBy user with id %v ... %v", m.SentByID, err)
		return nil, err
	}
	return &sender, nil
}

// GetThread finds and returns the Thread that this Message is attached to
func (m *Message) GetThread(tx *pop.Connection) (*Thread, error) {
	thread := Thread{}
	if err := tx.Find(&thread, m.ThreadID); err != nil {
		err = fmt.Errorf("error finding message thread id %v ... %v", m.ThreadID, err)
		return nil, err
	}
	return &thread, nil
}

// CreateFromInput a new message if authorized.
func (m *Message) CreateFromInput(tx *pop.Connection, user User, input api.MessageInput) error {
	var request Request
	if aErr := findAndValidateRequest(tx, user, input, &request); aErr != nil {
		return aErr
	}

	var thread Thread
	if input.ThreadID != nil && *input.ThreadID != "" {
		if aErr := findAndValidateThread(tx, user, input, &thread, request); aErr != nil {
			return aErr
		}
	} else {
		err := thread.CreateWithParticipants(tx, request, user)
		if err != nil {
			err = errors.New("failed to create new thread on request, " + err.Error())
			return api.NewAppError(err, api.ErrorCreateFailure, api.CategoryInternal)
		}
	}

	m.Content = input.Content
	m.ThreadID = thread.ID
	m.Thread = thread
	m.SentByID = user.ID
	if err := create(tx, m); err != nil {
		appError := api.AppError{
			Category: api.CategoryInternal,
			Key:      api.ErrorCreateFailure,
			Err:      errors.New("failed to create new message, " + err.Error()),
		}
		return &appError
	}

	return nil
}

func findAndValidateRequest(tx *pop.Connection, user User, input api.MessageInput, request *Request) *api.AppError {
	if err := request.FindByUUID(tx, input.RequestID); err != nil {
		appError := api.AppError{
			Category: api.CategoryUser,
			Key:      api.ErrorMessageBadRequestUUID,
			Err:      errors.New("error with new message: " + err.Error()),
		}
		return &appError
	}

	if isVisible, err := request.IsVisible(tx, user); err != nil {
		appError := api.AppError{
			Category: api.CategoryInternal,
			Key:      api.ErrorQueryFailure,
			Err:      errors.New("error with new message: " + err.Error()),
		}
		return &appError
	} else if !isVisible {
		appError := api.AppError{
			Category: api.CategoryForbidden,
			Key:      api.ErrorMessageRequestNotVisible,
			Err:      errors.New("user cannot create a message on request"),
		}
		return &appError
	}
	return nil
}

func findAndValidateThread(tx *pop.Connection, user User, input api.MessageInput, thread *Thread, request Request) *api.AppError {
	err := thread.FindByUUID(tx, *input.ThreadID)
	if err != nil {
		appError := api.AppError{
			Category: api.CategoryUser,
			Key:      api.ErrorMessageBadThreadUUID,
			Err:      errors.New("error with new message: " + err.Error()),
		}
		return &appError
	}
	if thread.RequestID != request.ID {
		appError := api.AppError{
			Category: api.CategoryUser,
			Key:      api.ErrorMessageThreadRequestMismatch,
			Err:      errors.New("thread is not valid for request"),
		}
		return &appError
	}
	if !thread.IsVisible(tx, user.ID) {
		appError := api.AppError{
			Category: api.CategoryForbidden,
			Key:      api.ErrorMessageThreadNotVisible,
			Err:      errors.New("user cannot create a message on thread"),
		}
		return &appError
	}

	return nil
}

// FindByID loads from DB the Message record identified by the given primary key
func (m *Message) FindByID(tx *pop.Connection, id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding message, invalid id")
	}

	var err error
	// Eager() with an empty argument list will load all fields, which is not what is intended here
	if len(eagerFields) > 0 {
		err = tx.Eager(eagerFields...).Find(m, id)
	} else {
		err = tx.Find(m, id)
	}

	if err != nil {
		return fmt.Errorf("error finding message by id, %s", err)
	}

	return tx.Find(m, id)
}

// findByUUID loads from DB the Message record identified by the given UUID
func (m *Message) findByUUID(tx *pop.Connection, id string) error {
	if id == "" {
		return errors.New("error: message uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", id).First(m); err != nil {
		return fmt.Errorf("error finding message by uuid: %s", err.Error())
	}

	return nil
}

// FindByUserAndUUID loads from DB the Message record identified by the given UUID, if the given user is allowed.
func (m *Message) FindByUserAndUUID(tx *pop.Connection, user User, id string) error {
	if err := m.findByUUID(tx, id); err != nil {
		return err
	}

	if user.AdminRole == UserAdminRoleSuperAdmin {
		return nil
	}

	var tp ThreadParticipant
	if err := tp.FindByThreadIDAndUserID(tx, m.ThreadID, user.ID); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return fmt.Errorf("error finding threadParticipant record for message %s, %s", id, err)
		}
		return fmt.Errorf("user %s has insufficient permissions to read message %s", user.UUID, id)
	}

	return nil
}

// converts models.Messages to api.Messages
func ConvertMessagesToAPIType(ctx context.Context, messages Messages) (api.Messages, error) {
	var output api.Messages
	if err := api.ConvertToOtherType(messages, &output); err != nil {
		err = errors.New("error converting messages to api.Messages: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages with their sentBy users
	for i := range output {
		sentByOutput, err := ConvertUser(ctx, messages[i].SentBy)
		if err != nil {
			err = errors.New("error converting messages sentBy to api.User: " + err.Error())
			return nil, err
		}

		output[i].ID = messages[i].UUID
		output[i].SentBy = &sentByOutput
	}

	return output, nil
}
