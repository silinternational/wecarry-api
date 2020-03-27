package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/silinternational/wecarry-api/domain"
)

type RequestHistory struct {
	ID         int           `json:"id" db:"id"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
	Status     RequestStatus `json:"status" db:"status"`
	RequestID  int           `json:"request_id" db:"request_id"`
	ReceiverID nulls.Int     `json:"receiver_id" db:"receiver_id"`
	ProviderID nulls.Int     `json:"provider_id" db:"provider_id"`
	Receiver   User          `belongs_to:"users"`
}

// String can be helpful for serializing the model
func (p RequestHistory) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// RequestHistories is merely for convenience and brevity
type RequestHistories []RequestHistory

// String can be helpful for serializing the model
func (p RequestHistories) String() string {
	jt, _ := json.Marshal(p)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *RequestHistory) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *RequestHistory) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (p *RequestHistory) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Load reads the selected fields from the database
func (p *RequestHistory) Load(fields ...string) error {
	if err := DB.Load(p, fields...); err != nil {
		return fmt.Errorf("error loading data for request history %v, %s", p.ID, err)
	}

	return nil
}

// createForRequest checks if the request has a status that is different than the
// most recent of its Request History entries.  If so, it creates a new Request History
// with the Request's new status.
func (pH RequestHistory) createForRequest(request Request) error {
	err := DB.Where("request_id = ?", request.ID).Last(&pH)

	if domain.IsOtherThanNoRows(err) {
		return err
	}

	if pH.Status != request.Status {
		newPH := RequestHistory{
			Status:     request.Status,
			RequestID:  request.ID,
			ReceiverID: nulls.NewInt(request.CreatedByID),
			ProviderID: request.ProviderID,
		}

		if err := newPH.Create(); err != nil {
			return err
		}
	}

	return nil
}

// pop deletes the most recent requestHistory entry for a request
// assuming it's status matches the expected one.
func (pH RequestHistory) popForRequest(request Request, currentStatus RequestStatus) error {
	if err := DB.Where("request_id = ?", request.ID).Last(&pH); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return err
		}
		domain.ErrLogger.Printf(
			"error popping request histories for request id %v. None Found", request.ID)
		return nil
	}

	if pH.Status != currentStatus {
		domain.ErrLogger.Printf(
			"error popping request histories for request id %v. Expected newStatus %s but found %s",
			request.ID, currentStatus, pH.Status)
		return nil
	}

	if err := DB.Destroy(&pH); err != nil {
		return err
	}

	return nil
}

func (pH *RequestHistory) getLastForRequest(request Request) error {
	if err := DB.Where("request_id = ?", request.ID).Last(pH); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return fmt.Errorf("error getting last Request History for request %v ... %v", request.ID, err)
		}
	}
	return nil
}

// Create stores the RequestHistory data as a new record in the database.
func (pH *RequestHistory) Create() error {
	return create(pH)
}
