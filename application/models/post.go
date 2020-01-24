package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
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

type PostVisibility string

const (
	PostVisibilityAll     PostVisibility = "ALL"
	PostVisibilityTrusted PostVisibility = "TRUSTED"
	PostVisibilitySame    PostVisibility = "SAME"
)

func (e PostVisibility) IsValid() bool {
	switch e {
	case PostVisibilityAll, PostVisibilityTrusted, PostVisibilitySame:
		return true
	}
	return false
}

func (e PostVisibility) String() string {
	return string(e)
}

func (e *PostVisibility) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostVisibility(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostVisibility", str)
	}
	return nil
}

func (e PostVisibility) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
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
			{status: PostStatusReceived},  // This transition is in here for later, in case one day it's not skippable
			{status: PostStatusCompleted}, // For now, `DELIVERED` is not a required step
			{status: PostStatusRemoved},
		},
		PostStatusDelivered: {
			{status: PostStatusCommitted, isBackStep: true}, // to correct a false delivery
			{status: PostStatusAccepted, isBackStep: true},  // to correct a false delivery
			{status: PostStatusCompleted},
		},
		PostStatusReceived: {
			{status: PostStatusAccepted, isBackStep: true},
			{status: PostStatusDelivered},
			{status: PostStatusCompleted},
		},
		PostStatusCompleted: {
			{status: PostStatusAccepted, isBackStep: true},  // to correct a false completion
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
	if status1 == "" {
		return false, nil
	}

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
	ID             int            `json:"id" db:"id"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	CreatedByID    int            `json:"created_by_id" db:"created_by_id"`
	Type           PostType       `json:"type" db:"type"`
	OrganizationID int            `json:"organization_id" db:"organization_id"`
	Status         PostStatus     `json:"status" db:"status"`
	Title          string         `json:"title" db:"title"`
	Size           PostSize       `json:"size" db:"size"`
	UUID           uuid.UUID      `json:"uuid" db:"uuid"`
	ReceiverID     nulls.Int      `json:"receiver_id" db:"receiver_id"`
	ProviderID     nulls.Int      `json:"provider_id" db:"provider_id"`
	Description    nulls.String   `json:"description" db:"description"`
	URL            nulls.String   `json:"url" db:"url"`
	Kilograms      float64        `json:"kilograms" db:"kilograms"`
	PhotoFileID    nulls.Int      `json:"photo_file_id" db:"photo_file_id"`
	DestinationID  int            `json:"destination_id" db:"destination_id"`
	OriginID       nulls.Int      `json:"origin_id" db:"origin_id"`
	MeetingID      nulls.Int      `json:"meeting_id" db:"meeting_id"`
	Visibility     PostVisibility `json:"visibility" db:"visibility"`

	CreatedBy    User          `belongs_to:"users"`
	Organization Organization  `belongs_to:"organizations"`
	Receiver     User          `belongs_to:"users"`
	Provider     User          `belongs_to:"users"`
	Files        PostFiles     `has_many:"post_files"`
	Histories    PostHistories `has_many:"post_histories"`
	PhotoFile    File          `belongs_to:"files"`
	Destination  Location      `belongs_to:"locations"`
	Origin       Location      `belongs_to:"locations"`
}

// PostCreatedEventData holds data needed by the New Post event listener
type PostCreatedEventData struct {
	PostID int
}

// String can be helpful for serializing the model
func (p Post) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Posts is merely for convenience and brevity
type Posts []Post

// String can be helpful for serializing the model
func (p Posts) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Create stores the Post data as a new record in the database.
func (p *Post) Create() error {
	if p.Visibility == "" {
		p.Visibility = PostVisibilitySame
	}
	return create(p)
}

// Update writes the Post data to an existing database record.
func (p *Post) Update() error {
	return update(p)
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
func (p *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: p.CreatedByID, Name: "CreatedBy"},
		&validators.StringIsPresent{Field: p.Type.String(), Name: "Type"},
		&validators.IntIsPresent{Field: p.OrganizationID, Name: "OrganizationID"},
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Size.String(), Name: "Size"},
		&validators.UUIDIsPresent{Field: p.UUID, Name: "UUID"},
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
	uuid := v.Post.UUID.String()
	if err := oldPost.FindByUUID(uuid); err != nil {
		v.Message = fmt.Sprintf("error finding existing post by UUID %s ... %v", uuid, err)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}

	if oldPost.Status == v.Post.Status {
		return
	}

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

func (p *Post) manageStatusTransition() error {
	if p.Status == "" {
		return nil
	}

	lastPostHistory := PostHistory{}
	if err := lastPostHistory.getLastForPost(*p); err != nil {
		return err
	}

	lastStatus := lastPostHistory.Status
	if p.Status == lastStatus {
		return nil
	}

	isBackStep, err := isTransitionBackStep(lastStatus, p.Status)
	if err != nil {
		return err
	}

	var pH PostHistory
	if isBackStep {
		err = pH.popForPost(*p, lastStatus)
	} else {
		err = pH.createForPost(*p)
	}

	if err != nil {
		return err
	}

	eventData := PostStatusEventData{
		OldStatus:     lastStatus,
		NewStatus:     p.Status,
		PostID:        p.ID,
		OldProviderID: *GetIntFromNullsInt(lastPostHistory.ProviderID),
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

	if err := p.manageStatusTransition(); err != nil {
		return err
	}

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

	var pH PostHistory
	if err := pH.createForPost(*p); err != nil {
		return err
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

	postFile := PostFile{PostID: p.ID, FileID: f.ID}
	if err := postFile.Create(); err != nil {
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
		if err := files[i].refreshURL(); err != nil {
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
		if err := p.Update(); err != nil {
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

	if err := p.PhotoFile.refreshURL(); err != nil {
		return nil, err
	}

	return &p.PhotoFile, nil
}

// scope query to only include posts from an organization associated with the current user
func scopeUserOrgs(cUser User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		orgs := cUser.GetOrgIDs()
		if len(orgs) == 0 {
			return q.Where("organization_id = -1")
		}
		return q.Where("organization_id IN (?)", convertSliceFromIntToInterface(orgs)...)
	}
}

// scope query to not include removed posts
func scopeNotRemoved() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status != ?", PostStatusRemoved)
	}
}

// scope query to not include removed or completed posts
func scopeNotCompleted() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status not in (?)", PostStatusRemoved, PostStatusCompleted)
	}
}

// FindByUserAndUUID finds the post identified by the given UUID if it belongs to the same organization as the
// given user and if the post has not been marked as removed.
func (p *Post) FindByUserAndUUID(ctx context.Context, user User, uuid string) error {
	return DB.Scope(scopeUserOrgs(user)).Scope(scopeNotRemoved()).
		Where("uuid = ?", uuid).First(p)
}

// FindByUser finds all posts visible to the current user. NOTE: at present, the posts are not sorted correctly; need
// to find a better way to construct the query
func (p *Posts) FindByUser(ctx context.Context, user User) error {
	q := DB.RawQuery(`
	WITH o AS (
		SELECT id FROM organizations WHERE id IN (
			SELECT organization_id FROM user_organizations WHERE user_id = ?
		)
	)
	SELECT * FROM posts WHERE
	(
		organization_id IN (SELECT id FROM o)
		OR
		organization_id IN (
			SELECT id FROM organizations WHERE id IN (
				SELECT secondary_id FROM organization_trusts WHERE primary_id IN (SELECT id FROM o)
			)
		) AND visibility IN ('ALL', 'TRUSTED')
	)
	AND status not in ('REMOVED', 'COMPLETED') ORDER BY created_at desc`, user.ID)
	if err := q.All(p); err != nil {
		return fmt.Errorf("error finding posts for user %s, %s", user.UUID.String(), err)
	}
	return nil
}

// FilterByUserTypeAndContents finds all posts belonging to the same organization as the given user,
// not marked as completed or removed and containing a certain search text.
func (p *Posts) FilterByUserTypeAndContents(ctx context.Context, user User, pType PostType, contains string) error {
	where := "type = ? and (LOWER(title) like ? or LOWER(description) like ?)"
	contains = `%` + strings.ToLower(contains) + `%`

	return DB.
		Scope(scopeUserOrgs(user)).
		Scope(scopeNotCompleted()).
		Where(where, pType, contains, contains).
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
	if p.MeetingID.Valid {
		return errors.New("Attempted to set destination on event-based post")
	}
	location.ID = p.DestinationID
	p.Destination = location
	return p.Destination.Update()
}

// SetOrigin sets the origin location fields, creating a new record in the database if necessary.
func (p *Post) SetOrigin(location Location) error {
	if p.OriginID.Valid {
		location.ID = p.OriginID.Int
		p.Origin = location
		return p.Origin.Update()
	}
	if err := location.Create(); err != nil {
		return err
	}
	p.OriginID = nulls.NewInt(location.ID)
	return p.Update()
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

	if user.ID != p.CreatedByID && !user.canEditAllPosts() {
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

	switch p.Type {
	case PostTypeRequest:
		if p.Status == PostStatusOpen && newStatus == PostStatusCommitted {
			return true
		}
		return newStatus == PostStatusDelivered && p.ProviderID.Int == user.ID
	case PostTypeOffer:
		return newStatus == PostStatusReceived && p.ReceiverID.Int == user.ID
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

// Meeting reads the meeting record, if it exists, and returns a pointer to the object.
func (p *Post) Meeting() (*Meeting, error) {
	if !p.MeetingID.Valid {
		return nil, nil
	}
	var meeting Meeting
	if err := DB.Find(&meeting, p.MeetingID); err != nil {
		return nil, err
	}

	return &meeting, nil
}
