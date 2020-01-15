package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type RequestCommitter struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	PostID    int       `json:"post_id" db:"post_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Post      Post      `belongs_to:"posts"`
	User      User      `belongs_to:"users"`
}

// String can be helpful for serializing the model
func (r RequestCommitter) String() string {
	jt, _ := json.Marshal(r)
	return string(jt)
}

// RequestCommitters is merely for convenience and brevity
type RequestCommitters []RequestCommitter

// String can be helpful for serializing the model
func (r RequestCommitters) String() string {
	jt, _ := json.Marshal(r)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (r *RequestCommitter) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: r.PostID, Name: "PostID"},
		&validators.IntIsPresent{Field: r.UserID, Name: "UserID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (r *RequestCommitter) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (r *RequestCommitter) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// FindByPostIDAndUserID reads a request record by the given Post ID and User ID
func (r *RequestCommitters) FindByPostID(postID int) error {
	if postID <= 0 {
		return fmt.Errorf("error finding request_committer, invalid id %v", postID)
	}

	if err := DB.Eager("User").Where("post_id = ?", postID).All(r); err != nil {
		return fmt.Errorf("failed to find request_committer record for post %d, %s",
			postID, err)
	}

	return nil
}

// FindByPostIDAndUserID reads a request record by the given Post ID and User ID
func (r *RequestCommitter) FindByPostIDAndUserID(postID, userID int) error {
	if postID <= 0 || userID <= 0 {
		return fmt.Errorf("error finding request_committer, invalid id ... postID %v, userID %v",
			postID, userID)
	}

	where := "user_id = ? AND post_id = ? AND post_type = ?"
	if err := DB.Where(where, userID, postID, PostTypeRequest).First(r); err != nil {
		return fmt.Errorf("failed to find request_committer record for user %d and post %d, %s",
			userID, postID, err)
	}
	return nil
}

// Create stores the RequestCommitter data as a new record in the database.
func (r *RequestCommitter) Create() error {
	return create(r)
}

// Update writes the RequestCommitter data to an existing database record.
func (r *RequestCommitter) Update() error {
	return update(r)
}

func (r *RequestCommitter) NewWithUserUUID(postID int, userUUID string) error {
	var user *User
	if err := user.FindByUUID(userUUID); err != nil {
		return err
	}

	var post *Post
	if err := post.FindByID(postID); err != nil {
		return err
	}

	r.PostID = post.ID
	r.UserID = user.ID
	return nil
}
