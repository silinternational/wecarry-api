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
	Uuid         uuid.UUID `json:"uuid" db:"uuid"`
	PostID       int       `json:"post_id" db:"post_id"`
	Post         Post      `belongs_to:"posts"`
	Messages     Messages  `has_many:"messages"`
	Participants Users     `many_to_many:"thread_participants"`
}

// String is not required by pop and may be deleted
func (t Thread) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Threads is not required by pop and may be deleted
type Threads []Thread

// String is not required by pop and may be deleted
func (t Threads) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Thread) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: t.Uuid, Name: "Uuid"},
		&validators.IntIsPresent{Field: t.PostID, Name: "PostID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Thread) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Thread) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// All retrieves all Threads from the database.
func (t *Threads) All(selectFields ...string) error {
	return DB.Select(selectFields...).All(t)
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

func (t *Thread) GetPost(selectFields []string) (*Post, error) {
	if t.PostID <= 0 {
		return nil, fmt.Errorf("error: PostID must be positive, got %v", t.PostID)
	}
	post := Post{}
	if err := DB.Select(selectFields...).Find(&post, t.PostID); err != nil {
		return nil, fmt.Errorf("error loading post %v %s", t.PostID, err)
	}
	return &post, nil
}

func (t *Thread) GetMessages(selectFields []string) ([]*Message, error) {
	var messages []*Message
	if err := DB.Select(selectFields...).Where("thread_id = ?", t.ID).All(&messages); err != nil {
		return messages, fmt.Errorf("error getting messages for thread id %v ... %v", t.ID, err)
	}

	return messages, nil
}

func (t *Thread) GetParticipants(selectFields []string) ([]*User, error) {
	var users []*User
	var threadParticipants []*ThreadParticipant

	if err := DB.Where("thread_id = ?", t.ID).Order("id asc").All(&threadParticipants); err != nil {
		return users, fmt.Errorf("error reading from thread_participants table %v ... %v", t.ID, err)
	}

	for _, tp := range threadParticipants {
		u := User{}

		if err := DB.Select(selectFields...).Find(&u, tp.UserID); err != nil {
			return users, fmt.Errorf("error finding users on thread %v ... %v", t.ID, err)
		}
		users = append(users, &u)
	}
	return users, nil
}

func (t *Thread) CreateWithParticipants(postUuid string, user User) error {
	var post Post
	if err := post.FindByUUID(postUuid); err != nil {
		return err
	}

	participants := Users{user}

	// Ensure Post Creator is one of the participants
	if post.CreatedBy.ID != 0 && post.CreatedBy.ID != user.ID {
		participants = append(participants, post.CreatedBy)
	}

	thread := Thread{
		PostID:       post.ID,
		Uuid:         domain.GetUuid(),
		Participants: participants,
	}

	if err := DB.Save(&thread); err != nil {
		err = fmt.Errorf("error saving new thread for message: %v", err.Error())
		return err
	}

	*t = thread
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
func (t *Thread) UpdateLastViewedAt(user User, time time.Time) error {
	var tp ThreadParticipant

	if err := tp.FindByThreadIDAndUserID(t.ID, user.ID); err != nil {
		return err
	}

	return tp.UpdateLastViewedAt(time)
}

// Load reads the selected fields from the database
func (t *Thread) Load(fields ...string) error {
	if err := DB.Load(t, fields...); err != nil {
		return fmt.Errorf("error loading data for thread %s, %s", t.Uuid.String(), err)
	}

	return nil
}

func (t *Thread) UnreadMessageCount(lastViewedAt time.Time) (int, error) {
	count := 0

	msgs, err := t.GetMessages([]string{"created_at"})
	if err != nil {
		return count, err
	}

	for _, m := range msgs {
		if m.CreatedAt.After(lastViewedAt) {
			count++
		}
	}

	return count, nil
}
