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
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/domain"
)

type RequestStatus string

const (
	RequestStatusOpen      RequestStatus = "OPEN"
	RequestStatusAccepted  RequestStatus = "ACCEPTED"
	RequestStatusDelivered RequestStatus = "DELIVERED"
	RequestStatusReceived  RequestStatus = "RECEIVED"
	RequestStatusCompleted RequestStatus = "COMPLETED"
	RequestStatusRemoved   RequestStatus = "REMOVED"

	RequestActionReopen       = "reopen"
	RequestActionOffer        = "offer"
	RequestActionRetractOffer = "retractOffer"
	RequestActionAccept       = "accept"
	RequestActionDeliver      = "deliver"
	RequestActionReceive      = "receive"
	// RequestActionComplete     = "complete"  //  For now Receiving a Request makes it Completed
	RequestActionRemove = "remove"
)

type StatusTransitionTarget struct {
	Status           RequestStatus
	IsBackStep       bool
	isProviderAction bool
}

type RequestVisibility string

const (
	RequestVisibilityAll     RequestVisibility = "ALL"
	RequestVisibilityTrusted RequestVisibility = "TRUSTED"
	RequestVisibilitySame    RequestVisibility = "SAME"
)

func (e RequestVisibility) IsValid() bool {
	switch e {
	case RequestVisibilityAll, RequestVisibilityTrusted, RequestVisibilitySame:
		return true
	}
	return false
}

func (e RequestVisibility) String() string {
	return string(e)
}

func (e *RequestVisibility) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RequestVisibility(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RequestVisibility", str)
	}
	return nil
}

func (e RequestVisibility) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func allStatusTransitions() map[RequestStatus][]StatusTransitionTarget {
	return map[RequestStatus][]StatusTransitionTarget{
		RequestStatusOpen: {
			{Status: RequestStatusAccepted},
			{Status: RequestStatusRemoved},
		},
		RequestStatusAccepted: {
			{Status: RequestStatusOpen, IsBackStep: true}, // to correct a false acceptance
			{Status: RequestStatusDelivered, isProviderAction: true},
			{Status: RequestStatusReceived},  // This transition is in here for later, in case one day it's not skippable
			{Status: RequestStatusCompleted}, // For now, `DELIVERED` is not a required step
			{Status: RequestStatusRemoved},
		},
		RequestStatusDelivered: {
			{Status: RequestStatusAccepted, IsBackStep: true, isProviderAction: true}, // to correct a false delivery
			{Status: RequestStatusCompleted},
		},
		RequestStatusReceived: {
			{Status: RequestStatusAccepted, IsBackStep: true},
			{Status: RequestStatusDelivered},
			{Status: RequestStatusCompleted},
		},
		RequestStatusCompleted: {
			{Status: RequestStatusAccepted, IsBackStep: true},  // to correct a false completion
			{Status: RequestStatusDelivered, IsBackStep: true}, // to correct a false completion
			//	{Status: RequestStatusReceived, IsBackStep: true, isProviderAction: true}, // to correct a false completion
		},
		RequestStatusRemoved: {},
	}
}

func getNextStatusPossibilities(status RequestStatus) ([]StatusTransitionTarget, error) {
	targets, ok := allStatusTransitions()[status]
	if !ok {
		return []StatusTransitionTarget{}, errors.New("unexpected initial status - " + status.String())
	}
	return targets, nil
}

func isTransitionValid(status1, status2 RequestStatus) (bool, error) {
	targets, err := getNextStatusPossibilities(status1)
	if err != nil {
		return false, err
	}

	for _, target := range targets {
		if status2 == target.Status {
			return true, nil
		}
	}

	return false, nil
}

func isTransitionBackStep(status1, status2 RequestStatus) (bool, error) {
	if status1 == "" {
		return false, nil
	}

	targets, err := getNextStatusPossibilities(status1)
	if err != nil {
		return false, err
	}

	for _, target := range targets {
		if status2 == target.Status {
			return target.IsBackStep, nil
		}
	}
	// Not worrying about invalid transitions, since this is called by AfterUpdate
	return false, nil
}

func statusActions() map[RequestStatus]string {
	return map[RequestStatus]string{
		RequestStatusOpen:      RequestActionReopen,
		RequestStatusAccepted:  RequestActionAccept,
		RequestStatusDelivered: RequestActionDeliver,
		// RequestStatusReceived:  RequestActionReceive,  // One day we may want this back in
		RequestStatusCompleted: RequestActionReceive,
		RequestStatusRemoved:   RequestActionRemove,
	}
}

func (e RequestStatus) IsValid() bool {
	switch e {
	case RequestStatusOpen, RequestStatusAccepted, RequestStatusDelivered, RequestStatusReceived,
		RequestStatusCompleted, RequestStatusRemoved:
		return true
	}
	return false
}

func (e RequestStatus) String() string {
	return string(e)
}

func (e *RequestStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RequestStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RequestStatus", str)
	}
	return nil
}

func (e RequestStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Request struct {
	ID             int               `json:"id" db:"id"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
	CreatedByID    int               `json:"created_by_id" db:"created_by_id"`
	OrganizationID int               `json:"organization_id" db:"organization_id"`
	NeededBefore   nulls.Time        `json:"needed_before" db:"needed_before"`
	Status         RequestStatus     `json:"status" db:"status"`
	CompletedOn    nulls.Time        `json:"completed_on" db:"completed_on"`
	Title          string            `json:"title" db:"title"`
	Size           RequestSize       `json:"size" db:"size"`
	UUID           uuid.UUID         `json:"uuid" db:"uuid"`
	ProviderID     nulls.Int         `json:"provider_id" db:"provider_id"`
	Description    nulls.String      `json:"description" db:"description"`
	URL            nulls.String      `json:"url" db:"url"`
	Kilograms      nulls.Float64     `json:"kilograms" db:"kilograms"`
	FileID         nulls.Int         `json:"file_id" db:"file_id"`
	DestinationID  int               `json:"destination_id" db:"destination_id"`
	OriginID       nulls.Int         `json:"origin_id" db:"origin_id"`
	MeetingID      nulls.Int         `json:"meeting_id" db:"meeting_id"`
	Visibility     RequestVisibility `json:"visibility" db:"visibility"`

	CreatedBy    User         `belongs_to:"users"`
	Organization Organization `belongs_to:"organizations"`
	Provider     User         `belongs_to:"users"`

	Files       RequestFiles `has_many:"request_files"`
	PhotoFile   File         `belongs_to:"files" fk_id:"FileID"`
	Destination Location     `belongs_to:"locations"`
	Origin      Location     `belongs_to:"locations"`
}

// RequestCreatedEventData holds data needed by the New Request event listener
type RequestCreatedEventData struct {
	RequestID int
}

// String can be helpful for serializing the model
func (r Request) String() string {
	jp, _ := json.Marshal(r)
	return string(jp)
}

// Requests is merely for convenience and brevity
type Requests []Request

// String can be helpful for serializing the model
func (r Requests) String() string {
	jp, _ := json.Marshal(r)
	return string(jp)
}

// Create stores the Request data as a new record in the database.
func (r *Request) Create() error {
	if r.Visibility == "" {
		r.Visibility = RequestVisibilitySame
	}
	return create(r)
}

// Update writes the Request data to an existing database record.
func (r *Request) Update() error {
	return update(r)
}

func (r *Request) NewWithUser(currentUser User) error {
	r.CreatedByID = currentUser.ID
	r.Status = RequestStatusOpen
	return nil
}

// SetProviderWithStatus sets the new Status of the Request and if needed it
// also sets the ProviderID (i.e. when the new status is ACCEPTED)
func (r *Request) SetProviderWithStatus(status RequestStatus, providerID *string) error {
	if status == RequestStatusAccepted {
		if providerID == nil {
			return errors.New("provider ID must not be nil")
		}

		var user User

		if err := user.FindByUUID(*providerID); err != nil {
			return errors.New("error finding provider: " + err.Error())
		}
		r.ProviderID = nulls.NewInt(user.ID)
	}
	r.Status = status
	return nil
}

// GetPotentialProviders returns the User objects associated with the Request's
// PotentialProviders
func (r *Request) GetPotentialProviders(currentUser User) (Users, error) {
	providers := PotentialProviders{}
	users, err := providers.FindUsersByRequestID(*r, currentUser)
	return users, err
}

// DestroyPotentialProviders destroys all the PotentialProvider records
// associated with the Request if the Request's status is COMPLETED
func (r *Request) DestroyPotentialProviders(status RequestStatus, user User) error {
	if status != RequestStatusCompleted {
		return nil
	}

	var pps PotentialProviders
	return pps.DestroyAllWithRequestUUID(r.UUID.String(), user)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (r *Request) Validate(tx *pop.Connection) (*validate.Errors, error) {
	v := []validate.Validator{
		&validators.IntIsPresent{Field: r.CreatedByID, Name: "CreatedBy"},
		&validators.IntIsPresent{Field: r.OrganizationID, Name: "OrganizationID"},
		&validators.StringIsPresent{Field: r.Title, Name: "Title"},
		&validators.StringIsPresent{Field: r.Size.String(), Name: "Size"},
		&validators.UUIDIsPresent{Field: r.UUID, Name: "UUID"},
		&validators.StringIsPresent{Field: r.Status.String(), Name: "Status"},
	}

	if !r.NeededBefore.Valid {
		return validate.Validate(v...), nil
	}

	var oldRequest Request
	_ = oldRequest.FindByID(r.ID)
	if oldRequest.ID == 0 || r.NeededBefore != oldRequest.NeededBefore {
		neededBeforeDate := r.NeededBefore.Time
		v = append(v, &validators.TimeAfterTime{
			FirstName:  "NeededBefore",
			FirstTime:  neededBeforeDate,
			SecondName: "Tomorrow",
			SecondTime: time.Now().Truncate(domain.DurationDay).Add(domain.DurationDay),
			Message:    fmt.Sprintf("Request neededBefore must not be before tomorrow. Got %v", neededBeforeDate),
		})
	}

	return validate.Validate(v...), nil
}

type createStatusValidator struct {
	Name    string
	Status  RequestStatus
	Message string
}

func (v *createStatusValidator) IsValid(errors *validate.Errors) {
	if v.Status == RequestStatusOpen {
		return
	}

	v.Message = fmt.Sprintf("Can only create a request with '%s' status, not '%s' status",
		RequestStatusOpen, v.Status)
	errors.Add(validators.GenerateKey(v.Name), v.Message)
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (r *Request) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&createStatusValidator{
			Name:   "Create Status",
			Status: r.Status,
		},
	), nil
}

type updateStatusValidator struct {
	Name    string
	Request *Request
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
	oldRequest := Request{}
	uuid := v.Request.UUID.String()
	if err := oldRequest.FindByUUID(uuid); err != nil {
		v.Message = fmt.Sprintf("error finding existing request by UUID %s ... %v", uuid, err)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}

	if oldRequest.Status == v.Request.Status {
		return
	}

	isTransValid, err := isTransitionValid(oldRequest.Status, v.Request.Status)
	if err != nil {
		v.Message = fmt.Sprintf("%s on request %s", err, uuid)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
		return
	}

	if !isTransValid {
		errorMsg := "cannot move request %s from '%s' status to '%s' status"
		v.Message = fmt.Sprintf(errorMsg, uuid, oldRequest.Status, v.Request.Status)
		errors.Add(validators.GenerateKey(v.Name), v.Message)
	}
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (r *Request) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&updateStatusValidator{
			Name:    "Status",
			Request: r,
		},
	), nil
}

// RequestStatusEventData holds data needed by the Request Status Updated event listener
type RequestStatusEventData struct {
	OldStatus     RequestStatus
	NewStatus     RequestStatus
	OldProviderID int
	RequestID     int
}

func (r *Request) manageStatusTransition() error {
	if r.Status == "" {
		return nil
	}
	lastRequestHistory := RequestHistory{}
	if err := lastRequestHistory.getLastForRequest(*r); err != nil {
		return err
	}

	lastStatus := lastRequestHistory.Status
	if r.Status == lastStatus {
		return nil
	}

	isBackStep, err := isTransitionBackStep(lastStatus, r.Status)
	if err != nil {
		return err
	}

	var rH RequestHistory
	if isBackStep {
		err = rH.popForRequest(*r, lastStatus)
	} else {
		err = rH.createForRequest(*r)
	}

	if err != nil {
		return err
	}

	eventData := RequestStatusEventData{
		OldStatus:     lastStatus,
		NewStatus:     r.Status,
		RequestID:     r.ID,
		OldProviderID: *GetIntFromNullsInt(lastRequestHistory.ProviderID),
	}

	e := events.Event{
		Kind:    domain.EventApiRequestStatusUpdated,
		Message: "Request Status changed",
		Payload: events.Payload{"eventData": eventData},
	}

	emitEvent(e)

	// If completed, hydrate CompletedOn. If not completed, nullify CompletedOn
	// Don't use r.UpdateColumns, due to this being called by the AfterUpdate function
	switch r.Status {
	case RequestStatusCompleted:
		if !r.CompletedOn.Valid {
			err := DB.RawQuery(
				fmt.Sprintf(`UPDATE requests set completed_on = '%s' where ID = %v`,
					time.Now().Format(domain.DateFormat), r.ID)).Exec()
			if err != nil {
				domain.ErrLogger.Printf("unable to set Request.CompletedOn for ID: %v, %s", r.ID, err)
			}
			if err := DB.Reload(r); err != nil {
				domain.ErrLogger.Printf("unable to reload Request ID: %v, %s", r.ID, err)
			}
		}
	case RequestStatusOpen, RequestStatusAccepted, RequestStatusDelivered:
		if r.CompletedOn.Valid {
			err := DB.RawQuery(
				fmt.Sprintf(`UPDATE requests set completed_on = NULL where ID = %v`, r.ID)).Exec()
			if err != nil {
				domain.ErrLogger.Printf("unable to nullify Request.CompletedOn for ID: %v, %s", r.ID, err)
			}
			if err := DB.Reload(r); err != nil {
				domain.ErrLogger.Printf("unable to reload Request ID: %v, %s", r.ID, err)
			}
		}
	}

	return nil
}

// Make sure there is no provider on an Open Request
func (r *Request) AfterUpdate(tx *pop.Connection) error {
	if err := r.manageStatusTransition(); err != nil {
		return err
	}

	if r.Status != RequestStatusOpen {
		return nil
	}

	r.ProviderID = nulls.Int{}

	// Don't try to use DB.Update inside AfterUpdate, since that gets into an eternal loop
	if err := DB.RawQuery(
		fmt.Sprintf(`UPDATE requests set provider_id = NULL where ID = %v`, r.ID)).Exec(); err != nil {
		domain.ErrLogger.Printf("error removing provider id from request: %s", err.Error())
	}

	return nil
}

// AfterCreate is called by Pop after successful creation of the record
func (r *Request) AfterCreate(tx *pop.Connection) error {
	if r.Status != RequestStatusOpen {
		return nil
	}

	var rH RequestHistory
	if err := rH.createForRequest(*r); err != nil {
		return err
	}

	e := events.Event{
		Kind:    domain.EventApiRequestCreated,
		Message: "Request created",
		Payload: events.Payload{"eventData": RequestCreatedEventData{
			RequestID: r.ID,
		}},
	}

	emitEvent(e)
	return nil
}

func (r *Request) FindByID(id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding request: id must a positive number")
	}

	if err := DB.Eager(eagerFields...).Find(r, id); err != nil {
		return fmt.Errorf("error finding request by id: %s", err.Error())
	}

	return nil
}

func (r *Request) FindByUUID(uuid string) error {
	if uuid == "" {
		return errors.New("error finding request: uuid must not be blank")
	}

	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := DB.Eager("CreatedBy").Where(queryString).First(r); err != nil {
		return fmt.Errorf("error finding request by uuid: %s", err.Error())
	}

	return nil
}

func (r *Request) FindByUUIDForCurrentUser(uuid string, user User) error {
	if err := r.FindByUUID(uuid); err != nil {
		return err
	}

	if !user.CanViewRequest(*r) {
		return fmt.Errorf("unauthorized: user %v may not view request %v.", user.ID, r.ID)
	}

	return nil
}

func (r *Request) GetCreator() (*User, error) {
	creator := User{}
	if err := DB.Find(&creator, r.CreatedByID); err != nil {
		return nil, err
	}
	return &creator, nil
}

func (r *Request) GetProvider() (*User, error) {
	provider := User{}
	if err := DB.Find(&provider, r.ProviderID); err != nil {
		return nil, nil // provider is a nullable field, so ignore any error
	}
	return &provider, nil
}

// GetStatusTransitions finds the forward and backward transitions for the current user
func (r *Request) GetStatusTransitions(currentUser User) ([]StatusTransitionTarget, error) {
	statusOptions, err := getNextStatusPossibilities(r.Status)
	if err != nil {
		domain.ErrLogger.Printf(err.Error())
		return statusOptions, nil
	}

	finalOptions := []StatusTransitionTarget{}

	for _, o := range statusOptions {
		// User is the Creator - sees all but Provider's actions
		if currentUser.ID == r.CreatedByID && !o.isProviderAction {
			finalOptions = append(finalOptions, o)
			continue
		}
		// User is the Provider and sees only the Provider's actions
		if o.isProviderAction && r.ProviderID.Valid && currentUser.ID == r.ProviderID.Int {
			finalOptions = append(finalOptions, o)
		}
	}

	return finalOptions, nil
}

// GetPotentialProviderActions
func (r *Request) GetPotentialProviderActions(currentUser User) ([]string, error) {
	if r.Status != RequestStatusOpen || currentUser.ID == r.CreatedByID {
		return []string{}, nil
	}

	providers, err := r.GetPotentialProviders(currentUser)
	if err != nil {
		return []string{}, err
	}

	// User is not the Creator
	for _, pp := range providers {
		// If user is already one of the PotentiaProviders
		if pp.ID == currentUser.ID {
			return []string{RequestActionRetractOffer}, nil
		}
	}

	return []string{RequestActionOffer}, nil
}

func (r *Request) GetOrganization() (*Organization, error) {
	organization := Organization{}
	if err := DB.Find(&organization, r.OrganizationID); err != nil {
		return nil, err
	}

	return &organization, nil
}

// GetThreads finds all threads on this request in which the given user is participating
func (r *Request) GetThreads(user User) ([]Thread, error) {
	var threads Threads
	query := DB.Q().
		Join("thread_participants tp", "threads.id = tp.thread_id").
		Order("threads.updated_at DESC").
		Where("tp.user_id = ? AND threads.request_id = ?", user.ID, r.ID)
	if err := query.All(&threads); err != nil {
		return nil, err
	}

	return threads, nil
}

// AttachFile adds a previously-stored File to this Request
func (r *Request) AttachFile(fileID string) (File, error) {
	var f File
	if err := f.FindByUUID(fileID); err != nil {
		return f, err
	}

	requestFile := RequestFile{RequestID: r.ID, FileID: f.ID}
	if err := requestFile.Create(); err != nil {
		return f, err
	}
	if err := f.SetLinked(DB); err != nil {
		domain.ErrLogger.Printf("error marking new request file %d as linked, %s", f.ID, err)
	}

	return f, nil
}

// GetFiles retrieves the metadata for all of the files attached to this Request
func (r *Request) GetFiles() ([]File, error) {
	var rf []*RequestFile

	err := DB.Eager("File").
		Select().
		Where("request_id = ?", r.ID).
		Order("updated_at desc").
		All(&rf)
	if err != nil {
		return nil, fmt.Errorf("error getting files for request id %d, %s", r.ID, err)
	}

	files := make([]File, len(rf))
	for i, f := range rf {
		files[i] = f.File
		if err := files[i].RefreshURL(); err != nil {
			return files, err
		}
	}

	return files, nil
}

// AttachPhoto assigns a previously-stored File to this Request as its photo. Parameter `fileID` is the UUID
// of the photo to attach.
func (r *Request) AttachPhoto(fileID string) (File, error) {
	return addFile(r, fileID)
}

// RemoveFile removes an attached file from the Request
func (r *Request) RemoveFile() error {
	return removeFile(r)
}

// GetPhoto retrieves the file attached as the Request photo
func (r *Request) GetPhoto() (*File, error) {
	if err := DB.Load(r, "PhotoFile"); err != nil {
		return nil, err
	}

	if !r.FileID.Valid {
		return nil, nil
	}

	if err := r.PhotoFile.RefreshURL(); err != nil {
		return nil, err
	}

	return &r.PhotoFile, nil
}

// GetPhotoID retrieves UUID of the file attached as the Request photo
func (r *Request) GetPhotoID() (*string, error) {
	if err := DB.Load(r, "PhotoFile"); err != nil {
		return nil, err
	}

	if r.FileID.Valid {
		photoID := r.PhotoFile.UUID.String()
		return &photoID, nil
	}
	return nil, nil
}

// scope query to only include requests from an organization associated with the current user
func scopeUserOrgs(cUser User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		orgs := cUser.GetOrgIDs()
		if len(orgs) == 0 {
			return q.Where("organization_id = -1")
		}
		return q.Where("organization_id IN (?)", convertSliceFromIntToInterface(orgs)...)
	}
}

// scope query to not include removed requests
func scopeNotRemoved() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status != ?", RequestStatusRemoved)
	}
}

// scope query to not include removed or completed requests
func scopeNotCompleted() pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		return q.Where("status not in (?)", RequestStatusRemoved, RequestStatusCompleted)
	}
}

// FindByUserAndUUID finds the request identified by the given UUID if it belongs to the same organization as the
// given user and if the request has not been marked as removed.
// FIXME: This method will fail to find a shared request from a trusted Organization
//func (p *Request) FindByUserAndUUID(ctx context.Context, user User, uuid string) error {
//	return DB.Scope(scopeUserOrgs(user)).Scope(scopeNotRemoved()).
//		Where("uuid = ?", uuid).First(p)
//}

// RequestFilterParams are optional parameters to narrow the list of requests returned from a query
type RequestFilterParams struct {
	Destination *Location
	Origin      *Location
	SearchText  *string
	RequestID   *int
}

// FindByUser finds all requests visible to the current user, optionally filtered by location or search text.
func (p *Requests) FindByUser(ctx context.Context, user User, filter RequestFilterParams) error {
	if user.ID == 0 {
		return errors.New("invalid User ID in Requests.FindByUser")
	}

	if !user.HasOrganization() {
		*p = Requests{}
		return nil
	}

	selectClause := `
	WITH o AS (
		SELECT id FROM organizations WHERE id IN (
			SELECT organization_id FROM user_organizations WHERE user_id = ?
		)
	)
	SELECT * FROM requests WHERE
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

	args := []interface{}{
		user.ID, RequestVisibilityAll, RequestVisibilityTrusted, RequestStatusRemoved,
		RequestStatusCompleted,
	}

	if filter.SearchText != nil {
		selectClause = selectClause + " AND (LOWER(title) LIKE ? or LOWER(description) LIKE ?)"
		likeText := "%" + strings.ToLower(*filter.SearchText) + "%"
		args = append(args, likeText, likeText)
	}
	if filter.RequestID != nil {
		selectClause = selectClause + " AND requests.id = ?"
		args = append(args, *filter.RequestID)
	}

	requests := Requests{}
	q := DB.RawQuery(selectClause+" ORDER BY created_at desc", args...)
	if err := q.All(&requests); err != nil {
		return fmt.Errorf("error finding requests for user %s, %s", user.UUID.String(), err)
	}

	if filter.Destination != nil {
		requests = requests.FilterDestination(*filter.Destination)
	}
	if filter.Origin != nil {
		requests = requests.FilterOrigin(*filter.Origin)
	}

	*p = Requests{}
	for i := range requests {
		*p = append(*p, requests[i])
	}
	return nil
}

// GetDestination reads the destination record, if it exists, and returns the Location object.
func (r *Request) GetDestination() (*Location, error) {
	location := Location{}
	if err := DB.Find(&location, r.DestinationID); err != nil {
		return nil, err
	}

	return &location, nil
}

// GetOrigin reads the origin record, if it exists, and returns the Location object.
func (r *Request) GetOrigin() (*Location, error) {
	if !r.OriginID.Valid {
		return nil, nil
	}
	location := Location{}
	if err := DB.Find(&location, r.OriginID); err != nil {
		return nil, err
	}

	return &location, nil
}

// RemoveOrigin removes the origin from the request
func (r *Request) RemoveOrigin() error {
	if !r.OriginID.Valid {
		return nil
	}

	if err := DB.Destroy(&Location{ID: r.OriginID.Int}); err != nil {
		return err
	}
	r.OriginID = nulls.Int{}
	// don't need to save the request because the database foreign key constraint is set to "ON DELETE SET NULL"
	return nil
}

// SetDestination sets the destination location fields, creating a new record in the database if necessary.
func (r *Request) SetDestination(location Location) error {
	if r.MeetingID.Valid {
		return errors.New("Attempted to set destination on event-based request")
	}
	location.ID = r.DestinationID
	r.Destination = location
	return r.Destination.Update()
}

// SetOrigin sets the origin location fields, creating a new record in the database if necessary.
func (r *Request) SetOrigin(location Location) error {
	if r.OriginID.Valid {
		location.ID = r.OriginID.Int
		r.Origin = location
		return r.Origin.Update()
	}
	if err := location.Create(); err != nil {
		return err
	}
	r.OriginID = nulls.NewInt(location.ID)
	return r.Update()
}

// IsEditable response with true if the given user is the owner of the request or an admin,
// and it is not in a locked status.
func (r *Request) IsEditable(user User) (bool, error) {
	if user.ID <= 0 {
		return false, errors.New("user.ID must be a valid primary key")
	}

	if r.CreatedByID <= 0 {
		if err := DB.Reload(r); err != nil {
			return false, err
		}
	}

	if user.ID != r.CreatedByID && !user.canEditAllRequests() {
		return false, nil
	}

	return r.isRequestEditable(), nil
}

// isRequestEditable defines at which states can requests be edited.
func (r *Request) isRequestEditable() bool {
	switch r.Status {
	case RequestStatusOpen, RequestStatusAccepted, RequestStatusReceived, RequestStatusDelivered:
		return true
	default:
		return false
	}
}

func (r *Request) canCreatorChangeStatus(newStatus RequestStatus) bool {
	// Creator can't move off of Delivered except to Completed
	if r.Status == RequestStatusDelivered {
		return newStatus == RequestStatusCompleted
	}

	// Creator can't move from Accepted to Delivered
	return !(r.Status == RequestStatusAccepted && newStatus == RequestStatusDelivered)
}

func (r *Request) canProviderChangeStatus(newStatus RequestStatus) bool {
	if r.Status != RequestStatusCompleted && newStatus == RequestStatusDelivered {
		return true
	}
	// for cancelling a DELIVERED status
	return r.Status == RequestStatusDelivered && newStatus == RequestStatusAccepted
}

// canUserChangeStatus defines which requests statuses can be changed by which users.
// Invalid transitions are not checked here; it is left for the validator to do this.
func (r *Request) canUserChangeStatus(user User, newStatus RequestStatus) bool {
	if user.AdminRole == UserAdminRoleSuperAdmin {
		return true
	}

	if r.Status == RequestStatusCompleted {
		return false
	}

	if r.CreatedByID == user.ID {
		return r.canCreatorChangeStatus(newStatus)
	}

	if r.ProviderID.Int == user.ID {
		return r.canProviderChangeStatus(newStatus)
	}

	return false
}

// GetAudience returns a list of all of the users which have visibility to this request. As of this writing, it is
// simply the users in the organization associated with this request.
func (r *Request) GetAudience() (Users, error) {
	if r.ID <= 0 {
		return nil, errors.New("invalid request ID in GetAudience")
	}
	org, err := r.GetOrganization()
	if err != nil {
		return nil, err
	}
	users, err := org.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("unable to get request organization user list, %s", err.Error())
	}
	return users, nil
}

// Meeting reads the meeting record, if it exists, and returns a pointer to the object.
func (r *Request) Meeting() (*Meeting, error) {
	if !r.MeetingID.Valid {
		return nil, nil
	}
	var meeting Meeting
	if err := DB.Find(&meeting, r.MeetingID); err != nil {
		return nil, err
	}

	return &meeting, nil
}

// FilterDestination returns a list of all requests with a Destination near the given location. The database is not
// touched.
func (p Requests) FilterDestination(location Location) Requests {
	filtered := make(Requests, 0)
	_ = DB.Load(&p, "Destination")
	for i := range p {
		if p[i].Destination.IsNear(location) {
			filtered = append(filtered, p[i])
		}
	}
	return filtered
}

// FilterOrigin returns a list of all requests that have an Origin near the given location. The database is not touched.
func (p Requests) FilterOrigin(location Location) Requests {
	filtered := make(Requests, 0)
	_ = DB.Load(&p, "Origin")
	for i := range p {
		if p[i].Origin.IsNear(location) {
			filtered = append(filtered, p[i])
		}
	}
	return filtered
}

// IsVisible returns true if the Request is visible to the given user. Only the request ID is used in this method.
func (r *Request) IsVisible(ctx context.Context, user User) bool {
	requests := Requests{}
	if err := requests.FindByUser(ctx, user, RequestFilterParams{RequestID: &r.ID}); err != nil {
		domain.Error(domain.GetBuffaloContext(ctx), "error in Request.IsVisible, "+err.Error())
		return false
	}
	return len(requests) > 0
}

func (r *Request) GetCurrentActions(user User) ([]string, error) {
	transitions, err := r.GetStatusTransitions(user)
	if err != nil {
		return []string{}, err
	}

	allActions := statusActions()

	actions := []string{}
	for _, t := range transitions {
		if action := allActions[t.Status]; action != "" {
			actions = append(actions, action)
		}
	}

	providerActions, err := r.GetPotentialProviderActions(user)
	if err != nil {
		return actions, err
	}

	actions = append(actions, providerActions...)

	return actions, nil
}

// Creator gets the full User record of the post creator
func (r *Request) Creator() (User, error) {
	var u User
	return u, DB.Find(&u, r.CreatedByID)
}
