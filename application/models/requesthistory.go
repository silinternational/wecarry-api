package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
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
func (rH RequestHistory) String() string {
	jt, _ := json.Marshal(rH)
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
func (rH *RequestHistory) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (rH *RequestHistory) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (rH *RequestHistory) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Load reads the selected fields from the database
func (rH *RequestHistory) Load(tx *pop.Connection, fields ...string) error {
	if err := tx.Load(rH, fields...); err != nil {
		return fmt.Errorf("error loading data for request history %v, %s", rH.ID, err)
	}

	return nil
}

// createForRequest checks if the request has a status that is different than the
// most recent of its Request History entries.  If so, it creates a new Request History
// with the Request's new status.
func (rH RequestHistory) createForRequest(tx *pop.Connection, request Request) error {
	err := tx.Where("request_id = ?", request.ID).Last(&rH)

	if domain.IsOtherThanNoRows(err) {
		return err
	}

	if rH.Status != request.Status {
		newRH := RequestHistory{
			Status:     request.Status,
			RequestID:  request.ID,
			ReceiverID: nulls.NewInt(request.CreatedByID),
			ProviderID: request.ProviderID,
		}

		if err := newRH.Create(tx); err != nil {
			return err
		}
	}

	return nil
}

// pop deletes the most recent requestHistory entry for a request
// assuming it's status matches the expected one.
func (rH RequestHistory) popForRequest(tx *pop.Connection, request Request, currentStatus RequestStatus) error {
	if err := tx.Where("request_id = ?", request.ID).Last(&rH); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return err
		}
		log.Errorf(
			"error popping request histories for request id %v. None Found", request.ID)
		return nil
	}

	if rH.Status != currentStatus {
		log.Errorf(
			"error popping request histories for request id %v. Expected newStatus %s but found %s",
			request.ID, currentStatus, rH.Status)
		return nil
	}

	if err := tx.Destroy(&rH); err != nil {
		return err
	}

	return nil
}

func (rH *RequestHistory) getLastForRequest(tx *pop.Connection, request Request) error {
	if err := tx.Where("request_id = ?", request.ID).Last(rH); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return fmt.Errorf("error getting last Request History for request %v ... %v", request.ID, err)
		}
	}
	return nil
}

// Create stores the RequestHistory data as a new record in the database.
func (rH *RequestHistory) Create(tx *pop.Connection) error {
	return create(tx, rH)
}
