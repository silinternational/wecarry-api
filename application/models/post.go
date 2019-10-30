package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/silinternational/wecarry-api/domain"
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
	PostStatusDelivered = "DELIVERED"
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

// Create stores the Post data as a new record in the database.
func (p *Post) Create() error {
	valErrs, err := DB.ValidateAndCreate(p)
	if err != nil {
		return err
	}

	if len(valErrs.Errors) > 0 {
		vErrs := FlattenPopErrors(valErrs)
		return errors.New(vErrs)
	}

	return nil
}

// Update writes the Post data to an existing database record.
func (p *Post) Update() error {
	valErrs, err := DB.ValidateAndUpdate(p)
	if err != nil {
		return err
	}

	if len(valErrs.Errors) > 0 {
		vErrs := FlattenPopErrors(valErrs)
		return errors.New(vErrs)
	}

	return nil
}

func (p *Post) NewWithUser(pType string, currentUser User) error {
	p.Uuid = domain.GetUuid()
	p.CreatedByID = currentUser.ID
	p.Status = PostStatusOpen

	switch pType {
	case PostTypeRequest:
		p.ReceiverID = nulls.NewInt(currentUser.ID)
	case PostTypeOffer:
		p.ProviderID = nulls.NewInt(currentUser.ID)
	default:
		return errors.New("bad type for new post: " + pType)
	}

	p.Type = pType

	return nil
}

func (p *Post) SetProviderWithStatus(status string, currentUser User) {

	if p.Type == PostTypeRequest && status == PostStatusCommitted {
		p.ProviderID = nulls.NewInt(currentUser.ID)
	}
	p.Status = status
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

type createStatusValidator struct {
	Name    string
	Status  string
	Message string
}

func (v *createStatusValidator) IsValid(errors *validate.Errors) {
	if v.Status == PostStatusOpen {
		return
	}

	v.Message = fmt.Sprintf("Can only create a post with '%s' status, not '%s' status",
		PostStatusOpen, v.Status)
	errors.Add(validators.GenerateKey(v.Name), v.Message)
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *Post) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&createStatusValidator{
			Name:   "Create Status",
			Status: p.Status,
		},
	), nil
	//return validate.NewErrors(), nil
}

type updateStatusValidator struct {
	Name    string
	Post    *Post
	Context buffalo.Context
	Message string
}

func (v *updateStatusValidator) IsValid(errors *validate.Errors) {
	switch v.Post.Type {
	case PostTypeOffer:
		v.isOfferValid(errors)
	case PostTypeRequest:
		v.isRequestValid(errors)
	}
}

func (v *updateStatusValidator) isOfferValid(errors *validate.Errors) {
	v.Message = "Offer status updates not allowed at this time"
	errors.Add(validators.GenerateKey(v.Name), v.Message)
}

func (v *updateStatusValidator) isRequestValid(errors *validate.Errors) {
	oldPost := Post{}
	uuid := v.Post.Uuid.String()
	if err := oldPost.FindByUUID(uuid); err != nil {
		v.Message = fmt.Sprintf("error finding existing post by UUID %s ... %v", uuid, err)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}

	if oldPost.Status == v.Post.Status {
		return
	}

	// Ensure that the new status is compatible with the old one in terms of a transition
	// allow for doing a step in reverse, in case there was a mistake going forward and
	// also allowing for some "unofficial" interaction happening outside of the app
	okTransitions := map[string][]string{
		PostStatusOpen:      {PostStatusCommitted, PostStatusRemoved},
		PostStatusCommitted: {PostStatusOpen, PostStatusAccepted, PostStatusDelivered, PostStatusRemoved},
		PostStatusAccepted:  {PostStatusOpen, PostStatusDelivered, PostStatusReceived, PostStatusRemoved},
		PostStatusDelivered: {PostStatusAccepted, PostStatusCompleted},
		PostStatusReceived:  {PostStatusAccepted, PostStatusDelivered, PostStatusCompleted},
		PostStatusCompleted: {PostStatusDelivered, PostStatusReceived},
		PostStatusRemoved:   {},
	}

	goodStatuses, ok := okTransitions[oldPost.Status]
	if !ok {
		msg := "unexpected status '%s' on post %s"
		domain.ErrLogger.Printf(msg, oldPost.Status, oldPost.Uuid.String())
		return
	}

	if !domain.IsStringInSlice(v.Post.Status, goodStatuses) {
		errorMsg := "cannot move post %s from '%s' status to '%s' status"
		v.Message = fmt.Sprintf(errorMsg, uuid, oldPost.Status, v.Post.Status)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (p *Post) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&updateStatusValidator{
			Name: "Status",
			Post: p,
		},
	), nil
}

// PostStatusEventData holds data needed by the Post Status Updated event listener
type PostStatusEventData struct {
	OldStatus     string
	NewStatus     string
	OldProviderID int
	PostID        int
}

func (p *Post) BeforeUpdate(tx *pop.Connection) error {
	oldPost := Post{}
	if err := tx.Find(&oldPost, p.ID); err != nil {
		domain.ErrLogger.Printf("error finding original post before update - uuid %v ... %v", p.Uuid, err)
		return nil
	}

	if oldPost.Status == "" || oldPost.Status == p.Status {
		return nil
	}

	eventData := PostStatusEventData{
		OldStatus:     oldPost.Status,
		NewStatus:     p.Status,
		PostID:        p.ID,
		OldProviderID: *GetIntFromNullsInt(p.ProviderID),
	}

	e := events.Event{
		Kind:    domain.EventApiPostStatusUpdated,
		Message: "Post Status changed",
		Payload: events.Payload{"eventData": eventData},
	}

	emitEvent(e)
	return nil
}

// Make sure there is no provider on an Open Request
func (p *Post) AfterUpdate(tx *pop.Connection) error {
	if p.Type != PostTypeRequest || p.Status != PostStatusOpen {
		return nil
	}

	p.ProviderID = nulls.Int{}

	// Don't try to use DB.Update inside AfterUpdate, since that gets into an eternal loop
	if err := DB.RawQuery(
		fmt.Sprintf(`UPDATE posts set provider_id = NULL where ID = %v`, p.ID)).Exec(); err != nil {
		domain.ErrLogger.Print("error removing provider id from post ... " + err.Error())
	}

	return nil
}

func (p *Post) FindByID(id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding post: id must a positive number")
	}

	if err := DB.Eager(eagerFields...).Find(p, id); err != nil {
		return fmt.Errorf("error finding post by id: %s", err.Error())
	}

	return nil
}

func (p *Post) FindByUUID(uuid string) error {
	if uuid == "" {
		return errors.New("error finding post: uuid must not be blank")
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
func (p *Post) GetThreads(fields []string, user User) ([]Thread, error) {
	if err := DB.Load(p, "Threads"); err != nil {
		return nil, fmt.Errorf("error getting threads for post id %v ... %v", p.ID, err)
	}

	var threads []Thread
	for i, t := range p.Threads {
		if err := DB.Load(&t, "Participants"); err != nil {
			return nil, fmt.Errorf("error getting participants for thread id %v ... %v", t.ID, err)
		}

		for _, participant := range t.Participants {
			if participant.ID == user.ID {
				threads = append(threads, p.Threads[i])
				break
			}
		}
	}

	return threads, nil
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
func (p *Post) GetFiles() ([]File, error) {
	var pf []*PostFile

	if err := DB.Eager("File").Select().Where("post_id = ?", p.ID).All(&pf); err != nil {
		return nil, fmt.Errorf("error getting files for post id %d, %s", p.ID, err)
	}

	files := make([]File, len(pf))
	for i, p := range pf {
		files[i] = p.File
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

	return &p.PhotoFile, nil
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
	return DB.
		Select(selectFields...).
		Scope(scopeUserOrgs(user)).
		Scope(scopeNotRemoved()).
		Order("created_at desc").
		All(p)
}

// GetDestination reads the destination record, if it exists, and returns the Location object.
func (p *Post) GetDestination() (*Location, error) {
	if !p.DestinationID.Valid {
		return nil, nil
	}
	location := Location{}
	if err := DB.Find(&location, p.DestinationID); err != nil {
		return nil, err
	}

	return &location, nil
}

// GetOrigin reads the origin record, if it exists, and returns the Location object.
func (p *Post) GetOrigin() (*Location, error) {
	if !p.OriginID.Valid {
		return nil, nil
	}
	location := Location{}
	if err := DB.Find(&location, p.OriginID); err != nil {
		return nil, err
	}

	return &location, nil
}

// SetDestination sets the destination location fields, creating a new record in the database if necessary.
func (p *Post) SetDestination(location Location) error {
	if p.DestinationID.Valid {
		location.ID = p.DestinationID.Int
		p.Destination = location
		return DB.Update(&p.Destination)
	}

	if err := DB.Create(&location); err != nil {
		return err
	}
	p.DestinationID = nulls.NewInt(location.ID)
	return DB.Update(p)
}

// SetOrigin sets the origin location fields, creating a new record in the database if necessary.
func (p *Post) SetOrigin(location Location) error {
	if p.OriginID.Valid {
		location.ID = p.OriginID.Int
		p.Origin = location
		return DB.Update(&p.Origin)
	}
	if err := DB.Create(&location); err != nil {
		return err
	}
	p.OriginID = nulls.NewInt(location.ID)
	return DB.Update(p)
}
