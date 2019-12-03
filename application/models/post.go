package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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

type PostType string

const (
	PostTypeRequest PostType = "REQUEST"
	PostTypeOffer   PostType = "OFFER"
)

func (e PostType) IsValid() bool {
	switch e {
	case PostTypeRequest, PostTypeOffer:
		return true
	}
	return false
}

func (e PostType) String() string {
	return string(e)
}

func (e *PostType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostType", str)
	}
	return nil
}

func (e PostType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PostStatus string

const (
	PostStatusOpen      PostStatus = "OPEN"
	PostStatusCommitted PostStatus = "COMMITTED"
	PostStatusAccepted  PostStatus = "ACCEPTED"
	PostStatusDelivered PostStatus = "DELIVERED"
	PostStatusReceived  PostStatus = "RECEIVED"
	PostStatusCompleted PostStatus = "COMPLETED"
	PostStatusRemoved   PostStatus = "REMOVED"
)

type statusTransitionTarget struct {
	status     PostStatus
	isBackStep bool
}

func getStatusTransitions() map[PostStatus][]statusTransitionTarget {
	return map[PostStatus][]statusTransitionTarget{
		PostStatusOpen: {
			{status: PostStatusCommitted},
			{status: PostStatusRemoved},
		},
		PostStatusCommitted: {
			{status: PostStatusOpen, isBackStep: true},
			{status: PostStatusAccepted},
			{status: PostStatusDelivered},
			{status: PostStatusRemoved},
		},
		PostStatusAccepted: {
			{status: PostStatusOpen},
			{status: PostStatusCommitted, isBackStep: true}, // to correct a false acceptance
			{status: PostStatusDelivered},
			{status: PostStatusReceived},
			{status: PostStatusRemoved},
		},
		PostStatusDelivered: {
			{status: PostStatusCommitted, isBackStep: true}, // to correct a false delivery
			{status: PostStatusAccepted, isBackStep: true},  // to correct a false delivery
			{status: PostStatusCompleted},
		},
		PostStatusReceived: {
			{status: PostStatusAccepted, isBackStep: true},
			{status: PostStatusDelivered, isBackStep: true},
			{status: PostStatusCompleted},
		},
		PostStatusCompleted: {
			{status: PostStatusDelivered, isBackStep: true}, // to correct a false completion
			{status: PostStatusReceived, isBackStep: true},  // to correct a false completion
		},
		PostStatusRemoved: {},
	}
}

func isTransitionValid(status1, status2 PostStatus) (bool, error) {
	transitions := getStatusTransitions()
	targets, ok := transitions[status1]
	if !ok {
		return false, errors.New("unexpected initial status - " + status1.String())
	}

	for _, target := range targets {
		if status2 == target.status {
			return true, nil
		}
	}

	return false, nil
}

func isTransitionBackStep(status1, status2 PostStatus) (bool, error) {
	transitions := getStatusTransitions()
	targets, ok := transitions[status1]
	if !ok {
		return false, errors.New("unexpected initial status - " + status1.String())
	}

	for _, target := range targets {
		if status2 == target.status {
			return target.isBackStep, nil
		}
	}

	return false, fmt.Errorf("invalid status transition from %s to %s", status1, status2)
}

func (e PostStatus) IsValid() bool {
	switch e {
	case PostStatusOpen, PostStatusCommitted, PostStatusAccepted, PostStatusDelivered, PostStatusReceived,
		PostStatusCompleted, PostStatusRemoved:
		return true
	}
	return false
}

func (e PostStatus) String() string {
	return string(e)
}

func (e *PostStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostStatus", str)
	}
	return nil
}

func (e PostStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PostSize string

const (
	PostSizeTiny   PostSize = "TINY"
	PostSizeSmall  PostSize = "SMALL"
	PostSizeMedium PostSize = "MEDIUM"
	PostSizeLarge  PostSize = "LARGE"
	PostSizeXlarge PostSize = "XLARGE"
)

func (e PostSize) String() string {
	return string(e)
}

type Post struct {
	ID             int           `json:"id" db:"id"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
	CreatedByID    int           `json:"created_by_id" db:"created_by_id"`
	Type           PostType      `json:"type" db:"type"`
	OrganizationID int           `json:"organization_id" db:"organization_id"`
	Status         PostStatus    `json:"status" db:"status"`
	Title          string        `json:"title" db:"title"`
	Size           PostSize      `json:"size" db:"size"`
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
	DestinationID  int           `json:"destination_id" db:"destination_id"`
	OriginID       nulls.Int     `json:"origin_id" db:"origin_id"`
	CreatedBy      User          `belongs_to:"users"`
	Organization   Organization  `belongs_to:"organizations"`
	Receiver       User          `belongs_to:"users"`
	Provider       User          `belongs_to:"users"`
	Files          PostFiles     `has_many:"post_files"`
	Histories      PostHistories `has_many:"post_histories"`
	PhotoFile      File          `belongs_to:"files"`
	Destination    Location      `belongs_to:"locations"`
	Origin         Location      `belongs_to:"locations"`
}

// PostCreatedEventData holds data needed by the New Post event listener
type PostCreatedEventData struct {
	PostID int
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
	if p.Uuid.Version() == 0 {
		p.Uuid = domain.GetUuid()
	}

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

func (p *Post) NewWithUser(pType PostType, currentUser User) error {
	p.CreatedByID = currentUser.ID
	p.Status = PostStatusOpen

	switch pType {
	case PostTypeRequest:
		p.ReceiverID = nulls.NewInt(currentUser.ID)
	case PostTypeOffer:
		p.ProviderID = nulls.NewInt(currentUser.ID)
	default:
		return errors.New("bad type for new post: " + pType.String())
	}

	p.Type = pType

	return nil
}

func (p *Post) SetProviderWithStatus(status PostStatus, currentUser User) {

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
		&validators.StringIsPresent{Field: p.Type.String(), Name: "Type"},
		&validators.IntIsPresent{Field: p.OrganizationID, Name: "OrganizationID"},
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Size.String(), Name: "Size"},
		&validators.UUIDIsPresent{Field: p.Uuid, Name: "Uuid"},
		&validators.StringIsPresent{Field: p.Status.String(), Name: "Status"},
	), nil
}

type createStatusValidator struct {
	Name    string
	Status  PostStatus
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
	//okTransitions := map[PostStatus][]PostStatus{
	//	PostStatusOpen:      {PostStatusCommitted, PostStatusRemoved},
	//	PostStatusCommitted: {PostStatusOpen, PostStatusAccepted, PostStatusDelivered, PostStatusRemoved},
	//	PostStatusAccepted:  {PostStatusOpen, PostStatusDelivered, PostStatusReceived, PostStatusRemoved},
	//	PostStatusDelivered: {PostStatusAccepted, PostStatusCompleted},
	//	PostStatusReceived:  {PostStatusAccepted, PostStatusDelivered, PostStatusCompleted},
	//	PostStatusCompleted: {PostStatusDelivered, PostStatusReceived},
	//	PostStatusRemoved:   {},
	//}

	isTransValid, err := isTransitionValid(oldPost.Status, v.Post.Status)
	if err != nil {
		v.Message = fmt.Sprintf("%s on post %s", err, uuid)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
		return
	}

	if !isTransValid {
		errorMsg := "cannot move post %s from '%s' status to '%s' status"
		v.Message = fmt.Sprintf(errorMsg, uuid, oldPost.Status, v.Post.Status)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}
}

// IsStatusInSlice iterates over a slice of PostStatus, looking for the given
// status. If found, true is returned. Otherwise, false is returned.
func IsStatusInSlice(needle PostStatus, haystack []PostStatus) bool {
	for _, hs := range haystack {
		if needle == hs {
			return true
		}
	}

	return false
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
	OldStatus     PostStatus
	NewStatus     PostStatus
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

	isBackStep, err := isTransitionBackStep(oldPost.Status, p.Status)
	if err != nil {
		return err
	}

	if isBackStep {
		err = p.popHistory(oldPost.Status)
	} else {
		err = p.createNewHistory()
	}

	if err != nil {
		return err
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

// AfterCreate is called by Pop after successful creation of the record
func (p *Post) AfterCreate(tx *pop.Connection) error {
	if p.Type != PostTypeRequest || p.Status != PostStatusOpen {
		return nil
	}

	e := events.Event{
		Kind:    domain.EventApiPostCreated,
		Message: "Post created",
		Payload: events.Payload{"eventData": PostCreatedEventData{
			PostID: p.ID,
		}},
	}

	emitEvent(e)
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

func (p *Post) GetCreator() (*User, error) {
	creator := User{}
	if err := DB.Find(&creator, p.CreatedByID); err != nil {
		return nil, err
	}
	return &creator, nil
}

func (p *Post) GetProvider() (*User, error) {
	provider := User{}
	if err := DB.Find(&provider, p.ProviderID); err != nil {
		return nil, nil // provider is a nullable field, so ignore any error
	}
	return &provider, nil
}

func (p *Post) GetReceiver() (*User, error) {
	receiver := User{}
	if err := DB.Find(&receiver, p.ReceiverID); err != nil {
		return nil, nil // receiver is a nullable field, so ignore any error
	}
	return &receiver, nil
}

func (p *Post) GetOrganization() (*Organization, error) {
	organization := Organization{}
	if err := DB.Find(&organization, p.OrganizationID); err != nil {
		return nil, err
	}

	return &organization, nil
}

// GetThreads finds all threads on this post in which the given user is participating
func (p *Post) GetThreads(user User) ([]Thread, error) {
	var threads Threads
	query := DB.Q().
		Join("thread_participants tp", "threads.id = tp.thread_id").
		Order("threads.updated_at DESC").
		Where("tp.user_id = ? AND threads.post_id = ?", user.ID, p.ID)
	if err := query.All(&threads); err != nil {
		return nil, err
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

	err := DB.Eager("File").
		Select().
		Where("post_id = ?", p.ID).
		Order("updated_at desc").
		All(&pf)
	if err != nil {
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

// scope query to only include posts from an organization associated with the current user
func scopeUserOrgs(cUser User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		orgs := cUser.GetOrgIDs()

		// convert []int to []interface{}
		s := make([]interface{}, len(orgs))
		for i, v := range orgs {
			s[i] = v
		}

		if len(s) == 0 {
			return q.Where("organization_id = -1")
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
func (p *Post) FindByUserAndUUID(ctx context.Context, user User, uuid string) error {
	return DB.Scope(scopeUserOrgs(user)).Scope(scopeNotRemoved()).
		Where("uuid = ?", uuid).First(p)
}

// FindByUser finds all posts belonging to the same organization as the given user and not marked as removed.
func (p *Posts) FindByUser(ctx context.Context, user User) error {
	return DB.
		Scope(scopeUserOrgs(user)).
		Scope(scopeNotRemoved()).
		Order("created_at desc").
		All(p)
}

// GetDestination reads the destination record, if it exists, and returns the Location object.
func (p *Post) GetDestination() (*Location, error) {
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
	location.ID = p.DestinationID
	p.Destination = location
	return DB.Update(&p.Destination)
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

// IsEditable response with true if the given user is the owner of the post or an admin,
// and it is not in a locked status.
func (p *Post) IsEditable(user User) (bool, error) {
	if user.ID <= 0 {
		return false, errors.New("user.ID must be a valid primary key")
	}

	if p.CreatedByID <= 0 {
		if err := DB.Reload(p); err != nil {
			return false, err
		}
	}

	if user.ID != p.CreatedByID && !user.CanEditAllPosts() {
		return false, nil
	}

	return p.isPostEditable(), nil
}

// isPostEditable defines at which states can posts be edited.
func (p *Post) isPostEditable() bool {
	switch p.Status {
	case PostStatusOpen, PostStatusCommitted, PostStatusAccepted, PostStatusReceived, PostStatusDelivered:
		return true
	default:
		return false
	}
}

// canUserChangeStatus defines which posts statuses can be changed by which users.
// Invalid transitions are not checked here; it is left for the validator to do this.
func (p *Post) canUserChangeStatus(user User, newStatus PostStatus) bool {
	if user.AdminRole == UserAdminRoleSuperAdmin {
		return true
	}

	if p.CreatedByID == user.ID {
		return true
	}

	if newStatus == PostStatusCommitted {
		return true
	}

	if p.ProviderID.Int != user.ID && p.ReceiverID.Int != user.ID {
		return false
	}

	if p.Type == PostTypeRequest && newStatus == PostStatusDelivered {
		return true
	}

	if p.Type == PostTypeOffer && newStatus == PostStatusReceived {
		return true
	}

	return false
}

// GetAudience returns a list of all of the users which have visibility to this post. As of this writing, it is
// simply the users in the organization associated with this post.
func (p *Post) GetAudience() (Users, error) {
	if p.ID <= 0 {
		return nil, errors.New("invalid post ID in GetAudience")
	}
	org, err := p.GetOrganization()
	if err != nil {
		return nil, err
	}
	users, err := org.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("unable to get post organization user list, %s", err.Error())
	}
	return users, nil
}

// GetLocationForNotifications gets the location most suitable for basic notifications. Specifically,
// the origin for requests, and the destination for offers.
func (p *Post) GetLocationForNotifications() (*Location, error) {
	var postLocation Location
	switch p.Type {
	case PostTypeRequest:
		if err := DB.Load(p, "Origin"); err != nil {
			return nil, fmt.Errorf("loading post origin failed, %s", err)
		}
		postLocation = p.Origin

	case PostTypeOffer:
		if err := DB.Load(p, "Destination"); err != nil {
			return nil, fmt.Errorf("loading post destination failed, %s", err)
		}
		postLocation = p.Destination
	}
	return &postLocation, nil
}

// createNewHistory checks if the post has a status that is different than the
// most recent of its Post History entries.  If so, it creates a new Post History
// with the Post's new status.
func (p *Post) createNewHistory() error {
	var oldPH PostHistory

	err := DB.Where("post_id = ?", p.ID).Last(&oldPH)

	if domain.IsOtherThanNoRows(err) {
		return err
	}

	if oldPH.Status != p.Status {
		newPH := PostHistory{
			Status:     p.Status,
			PostID:     p.ID,
			ReceiverID: p.ReceiverID,
			ProviderID: p.ProviderID,
		}

		if err := DB.Create(&newPH); err != nil {
			return err
		}
	}

	return nil
}

func (p *Post) popHistory(currentStatus PostStatus) error {
	var oldPH PostHistory

	if err := DB.Where("post_id = ?", p.ID).Last(&oldPH); err != nil {
		if domain.IsOtherThanNoRows(err) {
			return err
		}
		domain.ErrLogger.Printf(
			"error popping post histories for post id %v. None Found", p.ID)
		return nil
	}

	if oldPH.Status != currentStatus {
		domain.ErrLogger.Printf(
			"error popping post histories for post id %v. Expected newStatus %s but found %s",
			p.ID, currentStatus, oldPH.Status)
		return nil
	}

	if err := DB.Destroy(&oldPH); err != nil {
		return err
	}

	return nil
}
