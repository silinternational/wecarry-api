package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"

	"github.com/silinternational/wecarry-api/domain"
)

type PotentialProvider struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	PostID    int       `json:"post_id" db:"post_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Post      Post      `belongs_to:"posts"`
	User      User      `belongs_to:"users"`
}

// String can be helpful for serializing the model
func (r PotentialProvider) String() string {
	jt, _ := json.Marshal(r)
	return string(jt)
}

// PotentialProviders is merely for convenience and brevity
type PotentialProviders []PotentialProvider

// String can be helpful for serializing the model
func (r PotentialProviders) String() string {
	jt, _ := json.Marshal(r)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (r *PotentialProvider) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: r.PostID, Name: "PostID"},
		&validators.IntIsPresent{Field: r.UserID, Name: "UserID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (r *PotentialProvider) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (r *PotentialProvider) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// FindByPostIDAndUserID reads a request record by the given Post ID and User ID
func (r *PotentialProviders) FindUsersByPostID(postID int) (Users, error) {
	if postID <= 0 {
		return Users{}, fmt.Errorf("error finding potential_provider, invalid id %v", postID)
	}

	if err := DB.Eager("User").Where("post_id = ?", postID).All(r); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return Users{}, fmt.Errorf("failed to find potential_provider record for post %d, %s",
				postID, err)
		}
	}
	users := make(Users, len(*r))
	for i, rc := range *r {
		users[i] = rc.User
	}

	return users, nil
}

// FindByPostIDAndUserID reads a request record by the given Post ID and User ID
//func (r *PotentialProvider) FindByPostIDAndUserID(postID, userID int) error {
//	if postID <= 0 || userID <= 0 {
//		return fmt.Errorf("error finding potential_provider, invalid id ... postID %v, userID %v",
//			postID, userID)
//	}
//
//	where := "user_id = ? AND post_id = ? AND post_type = ?"
//	if err := DB.Where(where, userID, postID, PostTypeRequest).First(r); err != nil {
//		return fmt.Errorf("failed to find potential_provider record for user %d and post %d, %s",
//			userID, postID, err)
//	}
//	return nil
//}

// Create stores the PotentialProvider data as a new record in the database.
func (r *PotentialProvider) Create() error {
	return create(r)
}

// Update writes the PotentialProvider data to an existing database record.
func (r *PotentialProvider) Update() error {
	return update(r)
}

func (r *PotentialProvider) NewWithPostUUID(postUUID string, userID int) error {
	var user User
	if err := user.FindByID(userID); err != nil {
		return err
	}

	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return err
	}

	if post.Type != PostTypeRequest {
		return fmt.Errorf("Post Type must be Request not %s", post.Type)
	}

	if post.CreatedByID == userID {
		return errors.New("PotentialProvider User must not be the Post's Receiver.")
	}

	r.PostID = post.ID
	r.UserID = user.ID

	return nil
}
