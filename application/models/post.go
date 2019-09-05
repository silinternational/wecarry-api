package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/silinternational/handcarry-api/domain"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

const PostTypeRequest = "request"

const PostStatusUnfulfilled = "unfulfilled"

const PostSizeMedium = "medium"
const PostSizeSmall = "small"

type Post struct {
	ID             int           `json:"id" db:"id"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	CreatedByID    int           `json:"created_by_id" db:"created_by_id"`
	Type           string        `json:"type" db:"type"`
	OrganizationID int           `json:"organization_id" db:"organization_id"`
	Status         string        `json:"status" db:"status"`
	Title          string        `json:"title" db:"title"`
	Destination    nulls.String  `json:"destination" db:"destination"`
	Origin         nulls.String  `json:"origin" db:"origin"`
	Size           string        `json:"size" db:"size"`
	Uuid           uuid.UUID     `json:"uuid" db:"uuid"`
	ReceiverID     nulls.Int     `json:"receiver_id" db:"receiver_id"`
	ProviderID     nulls.Int     `json:"provider_id" db:"provider_id"`
	NeededAfter    time.Time     `json:"needed_after" db:"needed_after"`
	NeededBefore   time.Time     `json:"needed_before" db:"needed_before"`
	Category       string        `json:"category" db:"category"`
	Description    nulls.String  `json:"description" db:"description"`
	CreatedBy      User          `belongs_to:"users"`
	Organization   Organization  `belongs_to:"organizations"`
	Receiver       User          `belongs_to:"users"`
	Provider       User          `belongs_to:"users"`
	URL            nulls.String  `json:"url" db:"url"`
	Cost           nulls.Float64 `json:"cost" db:"cost"`
}

// String is not required by pop and may be deleted
func (p Post) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Posts is not required by pop and may be deleted
type Posts []Post

// String is not required by pop and may be deleted
func (p Posts) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (p *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: p.CreatedByID, Name: "CreatedBy"},
		&validators.StringIsPresent{Field: p.Type, Name: "Type"},
		&validators.IntIsPresent{Field: p.OrganizationID, Name: "OrganizationID"},
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Size, Name: "Size"},
		&validators.UUIDIsPresent{Field: p.Uuid, Name: "Uuid"},
		&validators.StringIsPresent{Field: p.Status, Name: "Status"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (p *Post) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (p *Post) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (p *Post) FindByUUID(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("error finding post: uuid must not be blank")
	}

	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := DB.Eager("CreatedBy").Where(queryString).First(p); err != nil {
		return fmt.Errorf("error finding post by uuid: %s", err.Error())
	}

	return nil
}

func (p *Post) GetCreator(fields []string) (*User, error) {
	creator := User{}
	if err := DB.Select(fields...).Find(&creator, p.CreatedByID); err != nil {
		return nil, err
	}
	return &creator, nil
}

func (p *Post) GetProvider(fields []string) (*User, error) {
	provider := User{}
	if err := DB.Select(fields...).Find(&provider, p.ProviderID); err != nil {
		return nil, nil // provider is a nullable field, so ignore any error
	}
	return &provider, nil
}

func (p *Post) GetReceiver(fields []string) (*User, error) {
	receiver := User{}
	if err := DB.Select(fields...).Find(&receiver, p.ReceiverID); err != nil {
		return nil, nil // receiver is a nullable field, so ignore any error
	}
	return &receiver, nil
}

func (p *Post) GetOrganization(fields []string) (*Organization, error) {
	organization := Organization{}
	if err := DB.Select(fields...).Find(&organization, p.OrganizationID); err != nil {
		return nil, err
	}

	return &organization, nil
}

func (p *Post) GetThreads(fields []string) ([]*Thread, error) {
	var threads []*Thread

	if err := DB.Select(fields...).Where("post_id = ?", p.ID).All(&threads); err != nil {
		return threads, fmt.Errorf("error getting threads for post id %v ... %v", p.ID, err)
	}

	return threads, nil
}

func (p *Post) GetThreadIdForUser(user User) (*string, error) {
	thread, err := FindThreadByPostIDAndUserID(p.ID, user.ID)
	if err != nil {
		return nil, err
	}

	threadUuid := thread.Uuid.String()
	if threadUuid == domain.EmptyUUID {
		return nil, nil
	}

	return &threadUuid, nil
}
