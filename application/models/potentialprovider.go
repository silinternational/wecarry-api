package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/events"
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
	User      User      `belongs_to:"users"`
}

// String can be helpful for serializing the model
func (p PotentialProvider) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// PotentialProviders is merely for convenience and brevity
type PotentialProviders []PotentialProvider

// String can be helpful for serializing the model
func (p PotentialProviders) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *PotentialProvider) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: p.PostID, Name: "PostID"},
		&validators.IntIsPresent{Field: p.UserID, Name: "UserID"},
		&uniqueTogetherValidator{Object: p, Name: "UniqueTogether"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *PotentialProvider) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (p *PotentialProvider) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

type uniqueTogetherValidator struct {
	Name    string
	Object  *PotentialProvider
	Message string
}

// IsValid ensures there are no other PotentialProviders with both the
// same PostID and UserID
func (v *uniqueTogetherValidator) IsValid(errors *validate.Errors) {
	p := PotentialProvider{}
	pID := v.Object.PostID
	uID := v.Object.UserID
	if err := DB.Where("post_id = ? and user_id = ?", pID, uID).First(&p); err != nil {
		if domain.IsOtherThanNoRows(err) {
			v.Message = "Error database for duplicate potential providers: " + err.Error()
			errors.Add(validators.GenerateKey(v.Name), v.Message)
		}
		return
	}

	// No error means we found a match
	v.Message = fmt.Sprintf("Duplicate potential provider exists with PostID: %v and UserID: %v",
		pID, uID)
	errors.Add(validators.GenerateKey(v.Name), v.Message)
	return
}

// PotentialProviderEventData holds data needed by the event listener that deals with a single PotentialProvider
type PotentialProviderEventData struct {
	UserID int
	PostID int
}

// Create stores the PotentialProvider data as a new record in the database.
func (p *PotentialProvider) Create() error {

	if err := create(p); err != nil {
		return err
	}

	eventData := PotentialProviderEventData{
		UserID: p.UserID,
		PostID: p.PostID,
	}

	e := events.Event{
		Kind:    domain.EventApiPotentialProviderCreated,
		Message: "Potential Provider created",
		Payload: events.Payload{"eventData": eventData},
	}

	emitEvent(e)

	return nil
}

// Update writes the PotentialProvider data to an existing database record.
func (p *PotentialProvider) Update() error {
	return update(p)
}

// FindUsersByPostID gets the Users associated with the PotentialProviders
func (p *PotentialProviders) FindUsersByPostID(postID int) (Users, error) {
	if postID <= 0 {
		return Users{}, fmt.Errorf("error finding potential_provider, invalid id %v", postID)
	}

	if err := DB.Eager("User").Where("post_id = ?", postID).All(p); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return Users{}, fmt.Errorf("failed to find potential_provider record for post %d, %s",
				postID, err)
		}
	}
	users := make(Users, len(*p))
	for i, pp := range *p {
		users[i] = pp.User
	}

	return users, nil
}

// CanUserAccessPotentialProvider returns whether the current user is a SuperAdmin, the Post's creator or
// the user associated with the PotentialProvider
func (p *PotentialProvider) CanUserAccessPotentialProvider(post Post, currentUser User) bool {
	if currentUser.AdminRole == UserAdminRoleSuperAdmin || currentUser.ID == post.CreatedByID {
		return true
	}
	return p.UserID == currentUser.ID
}

// FindWithPostUUIDAndUserUUID  finds the PotentialProvider associated with both the postUUID and the userUUID
func (p *PotentialProvider) FindWithPostUUIDAndUserUUID(postUUID, userUUID string, currentUser User) error {
	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return errors.New("unable to find Post in order to find PotentialProvider: " + err.Error())
	}

	var user User
	if err := user.FindByUUID(userUUID); err != nil {
		return errors.New("unable to find User in order to find PotentialProvider: " + err.Error())
	}

	if err := DB.Where("post_id = ? AND user_id = ?", post.ID, user.ID).First(p); err != nil {
		return errors.New("unable to find PotentialProvider: " + err.Error())
	}

	return nil
}

// NewWithPostUUID populates a new PotentialProvider but does not save it
func (p *PotentialProvider) NewWithPostUUID(postUUID string, userID int) error {
	var user User
	if err := user.FindByID(userID); err != nil {
		return err
	}

	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return err
	}

	if post.CreatedByID == userID {
		return errors.New("PotentialProvider User must not be the Post's Receiver.")
	}

	p.PostID = post.ID
	p.UserID = user.ID

	return nil
}

// Destroy destroys the PotentialProvider
func (p *PotentialProvider) Destroy() error {
	return DB.Destroy(p)
}

// DestroyAllWithPostUUID Destroys all the PotentialProviders associated with a Post depending
//  on whether the current user is a SuperAdmin or the Post's creator.
func (p *PotentialProviders) DestroyAllWithPostUUID(postUUID string, currentUser User) error {
	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return errors.New("unable to find Post in order to remove PotentialProviders: " + err.Error())
	}

	if currentUser.AdminRole != UserAdminRoleSuperAdmin && currentUser.ID != post.CreatedByID {
		return fmt.Errorf("user %v has insufficient permissions to destroy PotentialProviders for Post %v",
			currentUser.ID, post.ID)
	}

	if err := DB.Where("post_id = ?", post.ID).All(p); err != nil {
		return errors.New("unable to find Post's Potential Providers in order to remove them: " + err.Error())
	}
	return DB.Destroy(p)
}
