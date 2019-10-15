package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

const (
	PostTypeRequest = "REQUEST"
	PostTypeOffer   = "OFFER"

	PostSizeTiny   = "TINY"
	PostSizeSmall  = "SMALL"
	PostSizeMedium = "MEDIUM"
	PostSizeLarge  = "LARGE"
	PostSizeXlarge = "XLARGE"

	PostStatusOpen      = "OPEN"
	PostStatusCommitted = "COMMITTED"
	PostStatusAccepted  = "ACCEPTED"
	PostStatusReceived  = "RECEIVED"
	PostStatusCompleted = "COMPLETED"
	PostStatusRemoved   = "REMOVED"
)

type Post struct {
	ID             int           `json:"id" db:"id"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	CreatedByID    int           `json:"created_by_id" db:"created_by_id"`
	Type           string        `json:"type" db:"type"`
	OrganizationID int           `json:"organization_id" db:"organization_id"`
	Status         string        `json:"status" db:"status"`
	Title          string        `json:"title" db:"title"`
	Size           string        `json:"size" db:"size"`
	Uuid           uuid.UUID     `json:"uuid" db:"uuid"`
	ReceiverID     nulls.Int     `json:"receiver_id" db:"receiver_id"`
	ProviderID     nulls.Int     `json:"provider_id" db:"provider_id"`
	NeededAfter    time.Time     `json:"needed_after" db:"needed_after"`
	NeededBefore   time.Time     `json:"needed_before" db:"needed_before"`
	Category       string        `json:"category" db:"category"`
	Description    nulls.String  `json:"description" db:"description"`
	URL            nulls.String  `json:"url" db:"url"`
	Cost           nulls.Float64 `json:"cost" db:"cost"`
	PhotoFileID    nulls.Int     `json:"photo_file_id" db:"photo_file_id"`
	DestinationID  nulls.Int     `json:"destination_id" db:"destination_id"`
	OriginID       nulls.Int     `json:"origin_id" db:"origin_id"`
	CreatedBy      User          `belongs_to:"users"`
	Organization   Organization  `belongs_to:"organizations"`
	Receiver       User          `belongs_to:"users"`
	Provider       User          `belongs_to:"users"`
	Files          PostFiles     `has_many:"post_files"`
	PhotoFile      File          `belongs_to:"files"`
	Threads        Threads       `has_many:"threads"`
	Destination    Location      `belongs_to:"locations"`
	Origin         Location      `belongs_to:"locations"`
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

// GetThreads finds all threads on this post in which the given user is participating
func (p *Post) GetThreads(fields []string, user User) ([]*Thread, error) {
	if err := DB.Load(p, "Threads"); err != nil {
		return nil, fmt.Errorf("error getting threads for post id %v ... %v", p.ID, err)
	}

	var threads []*Thread
	for i, t := range p.Threads {
		if err := DB.Load(&t, "Participants"); err != nil {
			return nil, fmt.Errorf("error getting participants for thread id %v ... %v", t.ID, err)
		}

		for _, participant := range t.Participants {
			if participant.ID == user.ID {
				threads = append(threads, &(p.Threads[i]))
				break
			}
		}
	}

	return threads, nil
}

func (p *Post) GetThreadIdForUser(user User) (*string, error) {
	var thread Thread
	if err := thread.FindByPostIDAndUserID(p.ID, user.ID); err != nil {
		return nil, err
	}

	threadUuid := thread.Uuid.String()
	if threadUuid == domain.EmptyUUID {
		return nil, nil
	}

	return &threadUuid, nil
}

// AttachFile adds a previously-stored File to this Post
func (p *Post) AttachFile(fileID string) (File, error) {
	var f File
	if err := f.FindByUUID(fileID); err != nil {
		return f, err
	}

	if err := DB.Save(&PostFile{PostID: p.ID, FileID: f.ID}); err != nil {
		return f, err
	}

	return f, nil
}

// GetFiles retrieves the metadata for all of the files attached to this Post
func (p *Post) GetFiles() ([]*File, error) {
	var pf []*PostFile

	if err := DB.Eager("File").Select().Where("post_id = ?", p.ID).All(&pf); err != nil {
		return nil, fmt.Errorf("error getting files for post id %d, %s", p.ID, err)
	}

	files := make([]*File, len(pf))
	for i, p := range pf {
		files[i] = &p.File
		if err := files[i].RefreshURL(); err != nil {
			return files, err
		}
	}

	return files, nil
}

// AttachPhoto assigns a previously-stored File to this Post as its photo. Parameter `fileID` is the UUID
// of the photo to attach.
func (p *Post) AttachPhoto(fileID string) (File, error) {
	var f File
	if err := f.FindByUUID(fileID); err != nil {
		return f, nil // in case client can't recognize this error we'll fail silently
	}

	p.PhotoFileID = nulls.NewInt(f.ID)
	// if this is a new object, don't save it yet
	if p.ID != 0 {
		if err := DB.Update(p); err != nil {
			return f, err
		}
	}

	return f, nil
}

// GetPhoto retrieves the file attached as the Post photo
func (p *Post) GetPhoto() (*File, error) {
	if err := DB.Load(p, "PhotoFile"); err != nil {
		return nil, err
	}

	if !p.PhotoFileID.Valid {
		return nil, nil
	}

	if err := p.PhotoFile.RefreshURL(); err != nil {
		return nil, err
	}

	return &(p.PhotoFile), nil
}

// scope query to only include organizations for current user
func scopeUserOrgs(cUser User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		orgs := cUser.GetOrgIDs()

		// convert []int to []interface{}
		s := make([]interface{}, len(orgs))
		for i, v := range orgs {
			s[i] = v
		}

		return q.Where("organization_id IN (?)", s...)
	}
}

// scope query to not include removed posts
func scopeNotRemoved() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status != ?", PostStatusRemoved)
	}
}

// FindByUserAndUUID finds the post identified by the given UUID if it belongs to the same organization as the
// given user and if the post has not been marked as removed.
func (p *Post) FindByUserAndUUID(ctx context.Context, user User, uuid string, selectFields ...string) error {
	return DB.Select(selectFields...).Scope(scopeUserOrgs(user)).Scope(scopeNotRemoved()).
		Where("uuid = ?", uuid).First(p)
}

// FindByUser finds all posts belonging to the same organization as the given user and not marked as removed.
func (p *Posts) FindByUser(ctx context.Context, user User, selectFields ...string) error {
	return DB.Select(selectFields...).Scope(scopeUserOrgs(user)).Scope(scopeNotRemoved()).All(p)
}

func (p *Post) GetDestination() (*Location, error) {
	location := Location{}
	if err := DB.Find(&location, p.DestinationID); err != nil {
		return nil, err
	}

	return &location, nil
}

func (p *Post) GetOrigin() (*Location, error) {
	location := Location{}
	if err := DB.Find(&location, p.OriginID); err != nil {
		return nil, err
	}

	return &location, nil
}
