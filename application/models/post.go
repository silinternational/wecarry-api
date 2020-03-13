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

type PostStatus string

const (
	PostStatusOpen      PostStatus = "OPEN"
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
			{status: PostStatusAccepted},
			{status: PostStatusRemoved},
		},
		PostStatusAccepted: {
			{status: PostStatusOpen, isBackStep: true}, // to correct a false acceptance
			{status: PostStatusDelivered},
			{status: PostStatusReceived},  // This transition is in here for later, in case one day it's not skippable
			{status: PostStatusCompleted}, // For now, `DELIVERED` is not a required step
			{status: PostStatusRemoved},
		},
		PostStatusDelivered: {
			{status: PostStatusAccepted, isBackStep: true}, // to correct a false delivery
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
	// Not worrying about invalid transitions, since this is called by AfterUpdate
	return false, nil
}

func (e PostStatus) IsValid() bool {
	switch e {
	case PostStatusOpen, PostStatusAccepted, PostStatusDelivered, PostStatusReceived,
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
	OrganizationID int            `json:"organization_id" db:"organization_id"`
	NeededBefore   nulls.Time     `json:"needed_before" db:"needed_before"`
	Status         PostStatus     `json:"status" db:"status"`
	CompletedOn    nulls.Time     `json:"completed_on" db:"completed_on"`
	Title          string         `json:"title" db:"title"`
	Size           PostSize       `json:"size" db:"size"`
	UUID           uuid.UUID      `json:"uuid" db:"uuid"`
	ProviderID     nulls.Int      `json:"provider_id" db:"provider_id"`
	Description    nulls.String   `json:"description" db:"description"`
	URL            nulls.String   `json:"url" db:"url"`
	Kilograms      nulls.Float64  `json:"kilograms" db:"kilograms"`
	PhotoFileID    nulls.Int      `json:"photo_file_id" db:"photo_file_id"`
	DestinationID  int            `json:"destination_id" db:"destination_id"`
	OriginID       nulls.Int      `json:"origin_id" db:"origin_id"`
	MeetingID      nulls.Int      `json:"meeting_id" db:"meeting_id"`
	Visibility     PostVisibility `json:"visibility" db:"visibility"`

	CreatedBy    User         `belongs_to:"users"`
	Organization Organization `belongs_to:"organizations"`
	Provider     User         `belongs_to:"users"`

	Files       PostFiles `has_many:"post_files"`
	PhotoFile   File      `belongs_to:"files"`
	Destination Location  `belongs_to:"locations"`
	Origin      Location  `belongs_to:"locations"`
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

func (p *Post) NewWithUser(currentUser User) error {
	p.CreatedByID = currentUser.ID
	p.Status = PostStatusOpen
	return nil
}

// SetProviderWithStatus sets the new Status of the Post and if needed it
// also sets the ProviderID (i.e. when the new status is ACCEPTED)
func (p *Post) SetProviderWithStatus(status PostStatus, providerID *string) error {
	if status == PostStatusAccepted {
		if providerID == nil {
			return errors.New("provider ID must not be nil")
		}

		var user User

		if err := user.FindByUUID(*providerID); err != nil {
			return errors.New("error finding provider: " + err.Error())
		}
		p.ProviderID = nulls.NewInt(user.ID)
	}
	p.Status = status
	return nil
}

// GetPotentialProviders returns the User objects associated with the Post's
// PotentialProviders
func (p *Post) GetPotentialProviders() (Users, error) {
	providers := PotentialProviders{}
	users, err := providers.FindUsersByPostID(p.ID)
	return users, err
}

// DestroyPotentialProviders destroys all the PotentialProvider records
// associated with the Post if the Post's status is COMPLETED
func (p *Post) DestroyPotentialProviders(status PostStatus, user User) error {
	if status != PostStatusCompleted {
		return nil
	}

	var pps PotentialProviders
	return pps.DestroyAllWithPostUUID(p.UUID.String(), user)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: p.CreatedByID, Name: "CreatedBy"},
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
	tomorrow := time.Now().Truncate(domain.DurationDay).Add(domain.DurationDay)

	// If null, make it pass by pretending it's in the future
	neededBeforeDate := time.Now().Add(domain.DurationWeek)
	if p.NeededBefore.Valid {
		neededBeforeDate = p.NeededBefore.Time
	}

	return validate.Validate(
		&createStatusValidator{
			Name:   "Create Status",
			Status: p.Status,
		},
		&validators.TimeAfterTime{FirstName: "NeededBefore", FirstTime: neededBeforeDate,
			SecondName: "Tomorrow", SecondTime: tomorrow,
			Message: fmt.Sprintf("Post neededBefore must not be before tomorrow. Got %v", neededBeforeDate)},
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
	v.isRequestValid(errors)
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

	// If completed, hydrate CompletedOn. If not completed, nullify CompletedOn
	// Don't use p.UpdateColumns, due to this being called by the AfterUpdate function
	switch p.Status {
	case PostStatusCompleted:
		if !p.CompletedOn.Valid {
			err := DB.RawQuery(
				fmt.Sprintf(`UPDATE posts set completed_on = '%s' where ID = %v`,
					time.Now().Format(domain.DateFormat), p.ID)).Exec()
			if err != nil {
				domain.ErrLogger.Printf("unable to set Post.CompletedOn for ID: %v, %s", p.ID, err)
			}
			if err := DB.Reload(p); err != nil {
				domain.ErrLogger.Printf("unable to reload Post ID: %v, %s", p.ID, err)
			}
		}
	case PostStatusOpen, PostStatusAccepted, PostStatusDelivered:
		if p.CompletedOn.Valid {
			err := DB.RawQuery(
				fmt.Sprintf(`UPDATE posts set completed_on = NULL where ID = %v`, p.ID)).Exec()
			if err != nil {
				domain.ErrLogger.Printf("unable to nullify Post.CompletedOn for ID: %v, %s", p.ID, err)
			}
			if err := DB.Reload(p); err != nil {
				domain.ErrLogger.Printf("unable to reload Post ID: %v, %s", p.ID, err)
			}
		}
	}

	return nil
}

// Make sure there is no provider on an Open Request
func (p *Post) AfterUpdate(tx *pop.Connection) error {

	if err := p.manageStatusTransition(); err != nil {
		return err
	}

	if p.Status != PostStatusOpen {
		return nil
	}

	p.ProviderID = nulls.Int{}

	// Don't try to use DB.Update inside AfterUpdate, since that gets into an eternal loop
	if err := DB.RawQuery(
		fmt.Sprintf(`UPDATE posts set provider_id = NULL where ID = %v`, p.ID)).Exec(); err != nil {
		domain.ErrLogger.Printf("error removing provider id from post: %s", err.Error())
	}

	return nil
}

// AfterCreate is called by Pop after successful creation of the record
func (p *Post) AfterCreate(tx *pop.Connection) error {
	if p.Status != PostStatusOpen {
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

func (p *Post) FindByUUIDForCurrentUser(uuid string, user User) error {
	if err := p.FindByUUID(uuid); err != nil {
		return err
	}

	if !user.canViewPost(*p) {
		return fmt.Errorf("unauthorized: user %v may not view post %v.", user.ID, p.ID)
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
	if err := f.SetLinked(); err != nil {
		domain.ErrLogger.Printf("error marking new post file %d as linked, %s", f.ID, err)
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
		return f, err
	}

	oldID := p.PhotoFileID
	p.PhotoFileID = nulls.NewInt(f.ID)
	if p.ID > 0 {
		if err := DB.UpdateColumns(p, "photo_file_id"); err != nil {
			return f, err
		}
	}

	if err := f.SetLinked(); err != nil {
		domain.ErrLogger.Printf("error marking post image file %d as linked, %s", f.ID, err)
	}

	if oldID.Valid {
		oldFile := File{ID: oldID.Int}
		if err := oldFile.ClearLinked(); err != nil {
			domain.ErrLogger.Printf("error marking old post image file %d as unlinked, %s", oldFile.ID, err)
		}
	}

	return f, nil
}

// RemovePhoto removes an attached photo from the Post
func (p *Post) RemovePhoto() error {
	if p.ID < 1 {
		return fmt.Errorf("invalid Post ID %d", p.ID)
	}

	oldID := p.PhotoFileID
	p.PhotoFileID = nulls.Int{}
	if err := DB.UpdateColumns(p, "photo_file_id"); err != nil {
		return err
	}

	if !oldID.Valid {
		return nil
	}

	oldFile := File{ID: oldID.Int}
	if err := oldFile.ClearLinked(); err != nil {
		domain.ErrLogger.Printf("error marking old post photo file %d as unlinked, %s", oldFile.ID, err)
	}
	return nil
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

// GetPhotoID retrieves UUID of the file attached as the Post photo
func (p *Post) GetPhotoID() (*string, error) {
	if err := DB.Load(p, "PhotoFile"); err != nil {
		return nil, err
	}

	if p.PhotoFileID.Valid {
		photoID := p.PhotoFile.UUID.String()
		return &photoID, nil
	}
	return nil, nil
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
// FIXME: This method will fail to find a shared post from a trusted Organization
//func (p *Post) FindByUserAndUUID(ctx context.Context, user User, uuid string) error {
//	return DB.Scope(scopeUserOrgs(user)).Scope(scopeNotRemoved()).
//		Where("uuid = ?", uuid).First(p)
//}

// PostFilterParams are optional parameters to narrow the list of posts returned from a query
type PostFilterParams struct {
	Destination *Location
	Origin      *Location
	SearchText  *string
	PostID      *int
}

// FindByUser finds all posts visible to the current user, optionally filtered by location or search text.
func (p *Posts) FindByUser(ctx context.Context, user User, filter PostFilterParams) error {
	if user.ID == 0 {
		return errors.New("invalid User ID in Posts.FindByUser")
	}

	if !user.HasOrganization() {
		*p = Posts{}
		return nil
	}

	selectClause := `
	WITH o AS (
		SELECT id FROM organizations WHERE id IN (
			SELECT organization_id FROM user_organizations WHERE user_id = ?
		)
	)
	SELECT * FROM posts WHERE
	(
		organization_id IN (SELECT id FROM o)
		OR
		visibility = ?
		OR
		organization_id IN (
			SELECT id FROM organizations WHERE id IN (
				SELECT secondary_id FROM organization_trusts WHERE primary_id IN (SELECT id FROM o)
			)
		) AND visibility = ?
	)
	AND status not in (?, ?)`

	args := []interface{}{user.ID, PostVisibilityAll, PostVisibilityTrusted, PostStatusRemoved,
		PostStatusCompleted}

	if filter.SearchText != nil {
		selectClause = selectClause + " AND (LOWER(title) LIKE ? or LOWER(description) LIKE ?)"
		likeText := "%" + strings.ToLower(*filter.SearchText) + "%"
		args = append(args, likeText, likeText)
	}
	if filter.PostID != nil {
		selectClause = selectClause + " AND posts.id = ?"
		args = append(args, *filter.PostID)
	}

	posts := Posts{}
	q := DB.RawQuery(selectClause+" ORDER BY created_at desc", args...)
	if err := q.All(&posts); err != nil {
		return fmt.Errorf("error finding posts for user %s, %s", user.UUID.String(), err)
	}

	if filter.Destination != nil {
		posts = posts.FilterDestination(*filter.Destination)
	}
	if filter.Origin != nil {
		posts = posts.FilterOrigin(*filter.Origin)
	}

	*p = Posts{}
	for i := range posts {
		*p = append(*p, posts[i])
	}
	return nil
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

// RemoveOrigin removes the origin from the post
func (p *Post) RemoveOrigin() error {
	if !p.OriginID.Valid {
		return nil
	}

	if err := DB.Destroy(&Location{ID: p.OriginID.Int}); err != nil {
		return err
	}
	p.OriginID = nulls.Int{}
	// don't need to save the post because the database foreign key constraint is set to "ON DELETE SET NULL"
	return nil
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
	case PostStatusOpen, PostStatusAccepted, PostStatusReceived, PostStatusDelivered:
		return true
	default:
		return false
	}
}

func (p *Post) canCreatorChangeStatus(newStatus PostStatus) bool {
	// Creator can't move off of Delivered except to Completed
	if p.Status == PostStatusDelivered {
		return newStatus == PostStatusCompleted
	}

	// Creator can't move from Accepted to Delivered
	return !(p.Status == PostStatusAccepted && newStatus == PostStatusDelivered)
}

func (p *Post) canProviderChangeStatus(newStatus PostStatus) bool {
	if p.Status != PostStatusCompleted && newStatus == PostStatusDelivered {
		return true
	}
	// for cancelling a DELIVERED status
	return p.Status == PostStatusDelivered && newStatus == PostStatusAccepted
}

// canUserChangeStatus defines which posts statuses can be changed by which users.
// Invalid transitions are not checked here; it is left for the validator to do this.
func (p *Post) canUserChangeStatus(user User, newStatus PostStatus) bool {
	if user.AdminRole == UserAdminRoleSuperAdmin {
		return true
	}

	if p.Status == PostStatusCompleted {
		return false
	}

	if p.CreatedByID == user.ID {
		return p.canCreatorChangeStatus(newStatus)
	}

	if p.ProviderID.Int == user.ID {
		return p.canProviderChangeStatus(newStatus)
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

// FilterDestination returns a list of all posts with a Destination near the given location. The database is not
// touched.
func (p Posts) FilterDestination(location Location) Posts {
	filtered := make(Posts, 0)
	_ = DB.Load(&p, "Destination")
	for i := range p {
		if p[i].Destination.IsNear(location) {
			filtered = append(filtered, p[i])
		}
	}
	return filtered
}

// FilterOrigin returns a list of all posts that have an Origin near the given location. The database is not touched.
func (p Posts) FilterOrigin(location Location) Posts {
	filtered := make(Posts, 0)
	_ = DB.Load(&p, "Origin")
	for i := range p {
		if p[i].Origin.IsNear(location) {
			filtered = append(filtered, p[i])
		}
	}
	return filtered
}

// IsVisible returns true if the Post is visible to the given user. Only the post ID is used in this method.
func (p *Post) IsVisible(ctx context.Context, user User) bool {
	posts := Posts{}
	if err := posts.FindByUser(ctx, user, PostFilterParams{PostID: &p.ID}); err != nil {
		domain.Error(domain.GetBuffaloContext(ctx), "error in Post.IsVisible, "+err.Error())
		return false
	}
	return len(posts) > 0
}
