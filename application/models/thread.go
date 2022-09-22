package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
)

type Thread struct {
	ID           int       `json:"-" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	UUID         uuid.UUID `json:"uuid" db:"uuid"`
	RequestID    int       `json:"request_id" db:"request_id"`
	Request      Request   `belongs_to:"requests"`
	Participants Users     `many_to_many:"thread_participants" order_by:"id asc"`
	Messages     Messages  `json:"messages" db:"-"`

	LastViewedAt       *time.Time `json:"last_viewed_at" db:"-"`
	UnreadMessageCount int        `json:"unread_message_count" db:"-"`
}

// String can be helpful for serializing the model
func (t Thread) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Threads is merely for convenience and brevity
type Threads []Thread

// String can be helpful for serializing the model
func (t Threads) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (t *Thread) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: t.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: t.RequestID, Name: "RequestID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (t *Thread) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (t *Thread) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (t *Thread) FindByUUID(tx *pop.Connection, uuid string) error {
	if uuid == "" {
		return errors.New("error: thread uuid must not be blank")
	}

	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := tx.Where(queryString).First(t); err != nil {
		return fmt.Errorf("error finding thread by uuid: %s", err.Error())
	}

	return nil
}

func (t *Thread) LoadRequest(tx *pop.Connection, eagerFields ...string) error {
	if t.RequestID <= 0 {
		if err := t.FindByUUID(tx, t.UUID.String()); err != nil {
			return err
		}
	}
	request := Request{}
	// If no eagerFields, then don't use Eager at all, otherwise it uses Eager on all of them

	if len(eagerFields) > 0 {
		if err := tx.EagerPreload(eagerFields...).Find(&request, t.RequestID); err != nil {
			return fmt.Errorf("error loading (preloading) request %v %s", t.RequestID, err)
		}
	} else {
		if err := tx.Find(&request, t.RequestID); err != nil {
			return fmt.Errorf("error loading request %v %s", t.RequestID, err)
		}
	}
	t.Request = request
	return nil
}

func (t *Thread) LoadMessages(tx *pop.Connection, eagerFields ...string) error {
	var messages []Message
	// If no eagerFields, then don't use Eager at all, otherwise it uses Eager on all of them

	var q *pop.Query
	if len(eagerFields) > 0 {
		q = tx.EagerPreload(eagerFields...).Where("thread_id = ?", t.ID)
	} else {
		q = tx.Where("thread_id = ?", t.ID)
	}
	if err := q.All(&messages); err != nil {
		return fmt.Errorf("error getting messages for thread id %v ... %v", t.ID, err)
	}

	t.Messages = messages
	return nil
}

func (t *Thread) LoadParticipants(tx *pop.Connection) error {
	if err := tx.Load(t, "Participants"); err != nil {
		return fmt.Errorf("error loading threads participants %v ... %v", t.ID, err)
	}

	return nil
}

func (t *Thread) LoadThreadParticipants(tx *pop.Connection) error {
	if err := tx.Load(t, "ThreadParticipants"); err != nil {
		return fmt.Errorf("error loading threads thread_participants %v ... %v", t.ID, err)
	}

	return nil
}

func (t *Thread) CreateWithParticipants(tx *pop.Connection, request Request, user User) error {
	if user.ID <= 0 {
		return fmt.Errorf("error creating thread, invalid user ID %v", user.ID)
	}

	thread := Thread{
		RequestID: request.ID,
	}

	if err := thread.Create(tx); err != nil {
		err = fmt.Errorf("error saving new thread for message: %v", err.Error())
		return err
	}

	*t = thread
	return t.ensureParticipants(tx, request, user.ID)
}

func (t *Thread) ensureParticipants(tx *pop.Connection, request Request, userID int) error {
	err := t.LoadParticipants(tx)
	if domain.IsOtherThanNoRows(err) {
		err = errors.New("error getting threadParticipants for thread: " + err.Error())
		return err
	}

	if err := t.createParticipantIfNeeded(tx, t.Participants, request.CreatedByID); err != nil {
		return err
	}

	if userID == request.CreatedByID {
		return nil
	}

	return t.createParticipantIfNeeded(tx, t.Participants, userID)
}

func (t *Thread) createParticipantIfNeeded(tx *pop.Connection, tpUsers Users, userID int) error {
	for _, tPU := range tpUsers {
		if tPU.ID == userID {
			return nil
		}
	}

	newTP := ThreadParticipant{}
	newTP.ThreadID = t.ID
	newTP.UserID = userID
	if err := newTP.Create(tx); err != nil {
		return fmt.Errorf("error creating threadParticipant on thread ID: %v ... %v", t.ID, err)
	}
	return nil
}

// GetLastViewedAt gets the last viewed time for the given user on the thread
func (t *Thread) GetLastViewedAt(tx *pop.Connection, user User) (*time.Time, error) {
	var tp ThreadParticipant
	if err := tp.FindByThreadIDAndUserID(tx, t.ID, user.ID); err != nil {
		return nil, err
	}

	lastViewedAt := tp.LastViewedAt
	return &lastViewedAt, nil
}

// UpdateLastViewedAt sets the last viewed time for the given user on the thread
func (t *Thread) UpdateLastViewedAt(tx *pop.Connection, userID int, time time.Time) error {
	var tp ThreadParticipant

	if err := tp.FindByThreadIDAndUserID(tx, t.ID, userID); err != nil {
		return err
	}

	return tp.UpdateLastViewedAt(tx, time)
}

func (t *Thread) LoadForAPI(tx *pop.Connection, user User) error {
	err := t.LoadMessages(tx, "SentBy")
	if err != nil {
		return errors.New("error loading thread messages: " + err.Error())
	}

	err = t.LoadParticipants(tx)
	if err != nil {
		return errors.New("error loading participants: " + err.Error())
	}

	err = t.LoadRequest(tx, "CreatedBy")
	if err != nil {
		return errors.New("error loading thread request: " + err.Error())
	}

	lastViewed, err := t.GetLastViewedAt(tx, user)
	if err != nil {
		return errors.New("error getting thread lastViewedAt: " + err.Error())
	}

	t.LastViewedAt = lastViewed

	count, err := t.GetUnreadMessageCount(tx, user.ID, *lastViewed)
	if err != nil {
		return errors.New("error getting thread unreadMessageCount: " + err.Error())
	}

	t.UnreadMessageCount = count

	return nil
}

// Load reads the selected fields from the database
func (t *Thread) Load(tx *pop.Connection, fields ...string) error {
	if err := tx.Load(t, fields...); err != nil {
		return fmt.Errorf("error loading data for thread %s, %s", t.UUID.String(), err)
	}

	return nil
}

// GetUnreadMessageCount returns the number of messages on this thread that the current
//  user has not created and for which the CreatedAt value is after the lastViewedAt value
func (t *Thread) GetUnreadMessageCount(tx *pop.Connection, userID int, lastViewedAt time.Time) (int, error) {
	count := 0
	if userID <= 0 {
		return count, fmt.Errorf("error in GetUnreadMessageCount, invalid id %v", userID)
	}

	if len(t.Messages) == 0 {
		err := t.LoadMessages(tx)
		if err != nil {
			return count, err
		}
	}

	for _, m := range t.Messages {
		if m.SentByID != userID && m.CreatedAt.After(lastViewedAt) {
			count++
		}
	}

	return count, nil
}

// Create stores the Thread data as a new record in the database.
func (t *Thread) Create(tx *pop.Connection) error {
	return create(tx, t)
}

// Update writes the Thread data to an existing database record.
func (t *Thread) Update(tx *pop.Connection) error {
	return update(tx, t)
}

// IsVisible returns true if and only if the given user is already a participant of the message thread.
func (t *Thread) IsVisible(tx *pop.Connection, userID int) bool {
	if userID < 1 {
		return false
	}
	if err := t.LoadParticipants(tx); err == nil {
		for _, user := range t.Participants {
			if user.ID == userID {
				return true
			}
		}
	}
	return false
}

// converts models.Threads to api.Threads
func ConvertThreadsToAPIType(ctx context.Context, threads Threads) (api.Threads, error) {
	var output api.Threads
	if err := api.ConvertToOtherType(threads, &output); err != nil {
		err = errors.New("error converting threads to api.threads: " + err.Error())
		return nil, err
	}

	// Hydrate the thread's messages, participants
	for i := range output {
		messagesOutput, err := ConvertMessagesToAPIType(ctx, threads[i].Messages)
		if err != nil {
			return nil, err
		}
		output[i].Messages = &messagesOutput

		// Not converting Participants, since that happens automatically  above and
		// because it doesn't have nested related objects
		for j := range output[i].Participants {
			output[i].Participants[j].ID = threads[i].Participants[j].UUID
		}

		requestOutput, err := ConvertRequest(ctx, threads[i].Request)
		if err != nil {
			return nil, err
		}

		output[i].Request = &requestOutput
		output[i].ID = threads[i].UUID
	}

	return output, nil
}

func ConvertThread(ctx context.Context, thread Thread) (api.Thread, error) {
	var output api.Thread
	if err := api.ConvertToOtherType(thread, &output); err != nil {
		err = errors.New("error converting thread to api.thread: " + err.Error())
		return api.Thread{}, err
	}

	// Hydrate the thread's messages, participants

	messagesOutput, err := ConvertMessagesToAPIType(ctx, thread.Messages)
	if err != nil {
		return api.Thread{}, err
	}
	output.Messages = &messagesOutput

	// Not converting Participants, since that happens automatically  above and
	// because it doesn't have nested related objects
	for i := range output.Participants {
		output.Participants[i].ID = thread.Participants[i].UUID
	}

	if thread.Request.ID > 0 {
		requestOutput, err := ConvertRequest(ctx, thread.Request)
		if err != nil {
			return api.Thread{}, err
		}

		output.Request = &requestOutput
	} else {
		output.Request = nil
	}

	output.ID = thread.UUID
	return output, nil
}
