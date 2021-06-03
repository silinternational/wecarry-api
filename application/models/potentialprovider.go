package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"

	"github.com/silinternational/wecarry-api/domain"
)

type PotentialProvider struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	RequestID int       `json:"request_id" db:"request_id"`
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
		&validators.IntIsPresent{Field: p.RequestID, Name: "RequestID"},
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
// same RequestID and UserID
func (v *uniqueTogetherValidator) IsValid(errors *validate.Errors) {
	p := PotentialProvider{}
	pID := v.Object.RequestID
	uID := v.Object.UserID
	if err := DB.Where("request_id = ? and user_id = ?", pID, uID).First(&p); err != nil {
		if domain.IsOtherThanNoRows(err) {
			v.Message = "Error database for duplicate potential providers: " + err.Error()
			errors.Add(validators.GenerateKey(v.Name), v.Message)
		}
		return
	}

	// No error means we found a match
	v.Message = fmt.Sprintf("Duplicate potential provider exists with RequestID: %v and UserID: %v",
		pID, uID)
	errors.Add(validators.GenerateKey(v.Name), v.Message)
	return
}

// PotentialProviderEventData holds data needed by the event listener that deals with a single PotentialProvider
type PotentialProviderEventData struct {
	UserID    int
	RequestID int
}

// Create stores the PotentialProvider data as a new record in the database.
func (p *PotentialProvider) Create() error {
	if err := create(p); err != nil {
		return err
	}

	eventData := PotentialProviderEventData{
		UserID:    p.UserID,
		RequestID: p.RequestID,
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

// FindUsersByRequestID gets the Users associated with the Request's PotentialProviders
// This can be used without authorization by providing an empty currentUser object
//  (e.g. in the case of a notifications listener needing all the potentialProviders).
// If the currentUser is the requester or a SuperAdmin, then all potentialProvider Users are returned.
// If the currentUser is one of the potentialProviders, that User is returned.
// Otherwise, an empty slice of Users is returned.
func (p *PotentialProviders) FindUsersByRequestID(request Request, currentUser User) (Users, error) {
	if request.ID <= 0 {
		return Users{}, fmt.Errorf("error finding potential_provider, invalid id %v", request.ID)
	}

	// Default - only authorized to see self
	whereQ := fmt.Sprintf("request_id = %v and user_id = %v", request.ID, currentUser.ID)

	// See all, if authorized.
	if request.CreatedByID == currentUser.ID || currentUser.ID == 0 || currentUser.AdminRole == UserAdminRoleSuperAdmin {
		whereQ = fmt.Sprintf("request_id = %v", request.ID)
	}

	if err := DB.Eager("User").Where(whereQ).All(p); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return Users{}, fmt.Errorf("failed to find potential_provider records for request %d, %s",
				request.ID, err)
		}
	}

	users := make(Users, len(*p))
	for i, pp := range *p {
		users[i] = pp.User
	}

	return users, nil
}

// CanUserAccessPotentialProvider returns whether the current user is a SuperAdmin, the Request's creator or
// the user associated with the PotentialProvider
func (p *PotentialProvider) CanUserAccessPotentialProvider(request Request, currentUser User) bool {
	if currentUser.AdminRole == UserAdminRoleSuperAdmin || currentUser.ID == request.CreatedByID {
		return true
	}
	return p.UserID == currentUser.ID
}

// FindWithRequestUUIDAndUserUUID  finds the PotentialProvider associated with both the requestUUID and the userUUID
// No authorization checks are performed - they must be done separately
func (p *PotentialProvider) FindWithRequestUUIDAndUserUUID(requestUUID, userUUID string, currentUser User) error {
	var request Request
	if err := request.FindByUUID(requestUUID); err != nil {
		return errors.New("unable to find Request in order to find PotentialProvider: " + err.Error())
	}

	var user User
	if err := user.FindByUUID(userUUID); err != nil {
		return errors.New("unable to find User in order to find PotentialProvider: " + err.Error())
	}

	if err := DB.Where("request_id = ? AND user_id = ?", request.ID, user.ID).First(p); err != nil {
		return errors.New("unable to find PotentialProvider: " + err.Error())
	}

	if !p.CanUserAccessPotentialProvider(request, currentUser) {
		p = nil
		return errors.New("user not allowed to access PotentialProvider")
	}

	return nil
}

// NewWithRequestUUID populates a new PotentialProvider but does not save it
func (p *PotentialProvider) NewWithRequestUUID(requestUUID string, userID int) error {
	var user User
	if err := user.FindByID(userID); err != nil {
		return err
	}

	var request Request
	if err := request.FindByUUID(requestUUID); err != nil {
		return err
	}

	if request.CreatedByID == userID {
		return errors.New("PotentialProvider User must not be the Request's Receiver.")
	}

	p.RequestID = request.ID
	p.UserID = user.ID

	return nil
}

// Destroy destroys the PotentialProvider
func (p *PotentialProvider) Destroy() error {
	return DB.Destroy(p)
}

// DestroyAllWithRequestUUID Destroys all the PotentialProviders associated with a Request depending
//  on whether the current user is a SuperAdmin or the Request's creator.
func (p *PotentialProviders) DestroyAllWithRequestUUID(requestUUID string, currentUser User) error {
	var request Request
	if err := request.FindByUUID(requestUUID); err != nil {
		return errors.New("unable to find Request in order to remove PotentialProviders: " + err.Error())
	}

	if currentUser.AdminRole != UserAdminRoleSuperAdmin && currentUser.ID != request.CreatedByID {
		return fmt.Errorf("user %v has insufficient permissions to destroy PotentialProviders for Request %v",
			currentUser.ID, request.ID)
	}

	if err := DB.Where("request_id = ?", request.ID).All(p); err != nil {
		return errors.New("unable to find Request's Potential Providers in order to remove them: " + err.Error())
	}
	return DB.Destroy(p)
}
