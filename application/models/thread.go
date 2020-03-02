package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/silinternational/wecarry-api/domain"
)

type Thread struct {
	ID           int       `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	UUID         uuid.UUID `json:"uuid" db:"uuid"`
	PostID       int       `json:"post_id" db:"post_id"`
	Post         Post      `belongs_to:"posts"`
	Participants Users     `many_to_many:"thread_participants"`
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
		&validators.IntIsPresent{Field: t.PostID, Name: "PostID"},
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

// All retrieves all Threads from the database.
func (t *Threads) All() error {
	return DB.Order("updated_at desc").All(t)
}

func (t *Thread) FindByUUID(uuid string) error {
	if uuid == "" {
		return errors.New("error: thread uuid must not be blank")
	}

	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := DB.Where(queryString).First(t); err != nil {
		return fmt.Errorf("error finding thread by uuid: %s", err.Error())
	}

	return nil
}

func (t *Thread) GetPost() (*Post, error) {
	if t.PostID <= 0 {
		if err := t.FindByUUID(t.UUID.String()); err != nil {
			return nil, err
		}
	}
	post := Post{}
	if err := DB.Find(&post, t.PostID); err != nil {
		return nil, fmt.Errorf("error loading post %v %s", t.PostID, err)
	}
	return &post, nil
}

func (t *Thread) Messages() ([]Message, error) {
	var messages []Message
	if err := DB.Where("thread_id = ?", t.ID).All(&messages); err != nil {
		return messages, fmt.Errorf("error getting messages for thread id %v ... %v", t.ID, err)
	}

	return messages, nil
}

func (t *Thread) GetParticipants() ([]User, error) {
	var users []User
	var threadParticipants []*ThreadParticipant

	if err := DB.Where("thread_id = ?", t.ID).Order("id asc").All(&threadParticipants); err != nil {
		return users, fmt.Errorf("error reading from thread_participants table %v ... %v", t.ID, err)
	}

	for _, tp := range threadParticipants {
		u := User{}

		if err := DB.Find(&u, tp.UserID); err != nil {
			return users, fmt.Errorf("error finding users on thread %v ... %v", t.ID, err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (t *Thread) CreateWithParticipants(postUUID string, user User) error {
	if user.ID <= 0 {
		return fmt.Errorf("error creating thread, invalid user ID %v", user.ID)
	}

	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return err
	}

	thread := Thread{
		PostID: post.ID,
	}

	if err := thread.Create(); err != nil {
		err = fmt.Errorf("error saving new thread for message: %v", err.Error())
		return err
	}

	*t = thread
	return t.ensureParticipants(post, user.ID)
}

func (t *Thread) ensureParticipants(post Post, userID int) error {
	threadParticipants, err := t.GetParticipants()
	if domain.IsOtherThanNoRows(err) {
		err = errors.New("error getting threadParticipants for thread: " + err.Error())
		return err
	}

	if err := t.createParticipantIfNeeded(threadParticipants, post.CreatedByID); err != nil {
		return err
	}

	if userID == post.CreatedByID {
		return nil
	}

	return t.createParticipantIfNeeded(threadParticipants, userID)
}

func (t *Thread) createParticipantIfNeeded(tpUsers Users, userID int) error {
	for _, tPU := range tpUsers {
		if tPU.ID == userID {
			return nil
		}
	}

	newTP := ThreadParticipant{}
	newTP.ThreadID = t.ID
	newTP.UserID = userID
	if err := newTP.Create(); err != nil {
		return fmt.Errorf("error creating threadParticipant on thread ID: %v ... %v", t.ID, err)
	}
	return nil
}

// GetLastViewedAt gets the last viewed time for the given user on the thread
func (t *Thread) GetLastViewedAt(user User) (*time.Time, error) {
	var tp ThreadParticipant
	if err := tp.FindByThreadIDAndUserID(t.ID, user.ID); err != nil {
		return nil, err
	}
	lastViewedAt := tp.LastViewedAt
	return &lastViewedAt, nil
}

// UpdateLastViewedAt sets the last viewed time for the given user on the thread
func (t *Thread) UpdateLastViewedAt(userID int, time time.Time) error {
	var tp ThreadParticipant

	if err := tp.FindByThreadIDAndUserID(t.ID, userID); err != nil {
		return err
	}

	return tp.UpdateLastViewedAt(time)
}

// Load reads the selected fields from the database
func (t *Thread) Load(fields ...string) error {
	if err := DB.Load(t, fields...); err != nil {
		return fmt.Errorf("error loading data for thread %s, %s", t.UUID.String(), err)
	}

	return nil
}

// UnreadMessageCount returns the number of messages on this thread that the current
//  user has not created and for which the CreatedAt value is after the lastViewedAt value
func (t *Thread) UnreadMessageCount(userID int, lastViewedAt time.Time) (int, error) {
	count := 0
	if userID <= 0 {
		return count, fmt.Errorf("error in UnreadMessageCount, invalid id %v", userID)
	}

	msgs, err := t.Messages()
	if err != nil {
		return count, err
	}

	for _, m := range msgs {
		if m.SentByID != userID && m.CreatedAt.After(lastViewedAt) {
			count++
		}
	}

	return count, nil
}

// Create stores the Thread data as a new record in the database.
func (t *Thread) Create() error {
	return create(t)
}

// Update writes the Thread data to an existing database record.
func (t *Thread) Update() error {
	return update(t)
}
