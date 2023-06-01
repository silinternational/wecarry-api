package models

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/log"
)

type UserAdminRole string

const (
	UserAdminRoleSuperAdmin UserAdminRole = "SUPERADMIN"
	UserAdminRoleSalesAdmin UserAdminRole = "SALESADMIN"
	UserAdminRoleAdmin      UserAdminRole = "ADMIN"
	UserAdminRoleUser       UserAdminRole = "USER"
)

func (e UserAdminRole) IsValid() bool {
	switch e {
	case UserAdminRoleSuperAdmin, UserAdminRoleSalesAdmin, UserAdminRoleAdmin, UserAdminRoleUser:
		return true
	}
	return false
}

func (e UserAdminRole) String() string {
	return string(e)
}

// User model
type User struct {
	ID                 int               `json:"-" db:"id"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`
	Email              string            `json:"email" db:"email"`
	FirstName          string            `json:"first_name" db:"first_name"`
	LastName           string            `json:"last_name" db:"last_name"`
	Nickname           string            `json:"nickname" db:"nickname"`
	AdminRole          UserAdminRole     `json:"admin_role" db:"admin_role"`
	UUID               uuid.UUID         `json:"uuid" db:"uuid"`
	SocialAuthProvider nulls.String      `json:"social_auth_provider" db:"social_auth_provider"`
	FileID             nulls.Int         `json:"file_id" db:"file_id"`
	AuthPhotoURL       nulls.String      `json:"auth_photo_url" db:"auth_photo_url"`
	LocationID         nulls.Int         `json:"location_id" db:"location_id"`
	Organizations      Organizations     `many_to_many:"user_organizations" order_by:"name asc" json:"-"`
	UserOrganizations  UserOrganizations `has_many:"user_organizations" json:"-"`
	UserPreferences    UserPreferences   `has_many:"user_preferences" json:"-"`
	PhotoFile          File              `belongs_to:"files" fk_id:"FileID"`
	Location           Location          `belongs_to:"locations"`
}

// String can be helpful for serializing the model
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is merely for convenience and brevity
type Users []User

// String can be helpful for serializing the model
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.FirstName, Name: "FirstName"},
		&validators.StringIsPresent{Field: u.LastName, Name: "LastName"},
		&validators.StringIsPresent{Field: u.Nickname, Name: "Nickname"},
		&validators.UUIDIsPresent{Field: u.UUID, Name: "UUID"},
		&NullsStringIsURL{Field: u.AuthPhotoURL, Name: "AuthPhotoURL"},
		&domain.StringIsVisible{Field: u.Nickname, Name: "Nickname"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// All retrieves all Users from the database.
func (u *Users) All(tx *pop.Connection) error {
	return tx.Order("nickname asc").All(u)
}

// CreateAccessToken - Create and store new UserAccessToken
func (u *User) CreateAccessToken(tx *pop.Connection, org Organization, clientID string) (string, int64, error) {
	if clientID == "" {
		return "", 0, fmt.Errorf("cannot create token with empty clientID for user %s", u.Nickname)
	}

	token, _ := getRandomToken()
	hash := HashClientIdAccessToken(clientID + token)
	expireAt := createAccessTokenExpiry()

	userAccessToken := &UserAccessToken{
		UserID:      u.ID,
		AccessToken: hash,
		ExpiresAt:   expireAt,
	}

	if org.ID > 0 {
		userOrg, err := u.FindUserOrganization(tx, org)
		if err != nil {
			return "", 0, err
		}
		userAccessToken.UserOrganizationID = nulls.NewInt(userOrg.ID)
	}

	if err := userAccessToken.Create(tx); err != nil {
		return "", 0, err
	}

	return token, expireAt.UTC().Unix(), nil
}

// CreateOrglessAccessToken - Create and store new UserAccessToken with no associated UserOrg
func (u *User) CreateOrglessAccessToken(tx *pop.Connection, clientID string) (string, int64, error) {
	return u.CreateAccessToken(tx, Organization{}, clientID)
}

func (u *User) GetOrgIDs(tx *pop.Connection) []int {
	// ignore the error and allow the user's Organizations to be an empty slice.
	_ = tx.Load(u, "Organizations")

	s := make([]int, len(u.Organizations))
	for i, v := range u.Organizations {
		s[i] = v.ID
	}

	return s
}

func (u *User) hydrateFromAuthUser(tx *pop.Connection, authUser *auth.User, authType string) error {
	newUser := true
	if u.ID != 0 {
		newUser = false
	}

	// update attributes from authUser
	u.FirstName = authUser.FirstName
	u.LastName = authUser.LastName
	u.Email = authUser.Email

	if authType != "" {
		u.SocialAuthProvider = nulls.NewString(authType)
	}

	if authUser.PhotoURL != "" {
		u.AuthPhotoURL = nulls.NewString(authUser.PhotoURL)
	}

	// if new user they will need a unique Nickname
	if newUser {
		u.Nickname = authUser.Nickname
		if err := u.uniquifyNickname(tx, getShuffledPrefixes()); err != nil {
			return err
		}
	}
	if err := u.Save(tx); err != nil {
		return errors.New("unable to save user record: " + err.Error())
	}

	if newUser {
		e := events.Event{
			Kind:    domain.EventApiUserCreated,
			Message: "Nickname: " + u.Nickname + "  UUID: " + u.UUID.String(),
			Payload: events.Payload{"user": u},
		}
		emitEvent(e)
	}
	return nil
}

func (u *User) FindOrCreateFromAuthUser(tx *pop.Connection, orgID int, authUser *auth.User) error {
	userOrgs := UserOrganizations{}
	err := userOrgs.FindByAuthEmail(tx, authUser.Email, orgID)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(userOrgs) > 1 {
		return fmt.Errorf("too many user organizations found (%v), data integrity problem", len(userOrgs))
	}

	if len(userOrgs) == 1 {
		if userOrgs[0].AuthID != authUser.UserID {
			return errors.New("a user in this organization with this email address already exists with different user id")
		}
		err = tx.Where("uuid = ?", userOrgs[0].User.UUID).First(u)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	//if err := u.hydrateFromAuthUser(tx, authUser, ""); err != nil {
	//	return err
	//}

	if len(userOrgs) == 0 {
		//userOrg := &UserOrganization{
		//	OrganizationID: orgID,
		//	UserID:         u.ID,
		//	Role:           UserOrganizationRoleUser,
		//	AuthID:         authUser.UserID,
		//	AuthEmail:      u.Email,
		//	LastLogin:      time.Now(),
		//}
		//err = userOrg.Create(tx)
		//if err != nil {
		//	return fmt.Errorf("unable to create new user_organization record: %s", err.Error())
		//}

		return fmt.Errorf("no new user accounts can be created at this time")
	}

	return nil
}

// FindOrCreateFromOrglessAuthUser creates a new User based on an auth.User and
// sets its SocialAuthProvider field so they can login again in future.
func (u *User) FindOrCreateFromOrglessAuthUser(tx *pop.Connection, authUser *auth.User, authType string) error {
	if err := tx.Where("email = ?", authUser.Email).First(u); domain.IsOtherThanNoRows(err) {
		return errors.WithStack(err)
	}

	return u.hydrateFromAuthUser(tx, authUser, authType)
}

// CanCreateOrganization returns true if the given user is allowed to create organizations
func (u *User) CanCreateOrganization() bool {
	return u.AdminRole == UserAdminRoleSuperAdmin || u.AdminRole == UserAdminRoleSalesAdmin
}

// CanCreateOrganizationTrust returns true if the given user is allowed to create an OrganizationTrust
func (u *User) CanCreateOrganizationTrust() bool {
	return u.AdminRole == UserAdminRoleSuperAdmin || u.AdminRole == UserAdminRoleSalesAdmin
}

// CanRemoveOrganizationTrust returns true if the given user is allowed to remove an OrganizationTrust
func (u *User) CanRemoveOrganizationTrust(tx *pop.Connection, orgId int) bool {
	// if user is a system admin, allow
	if u.AdminRole == UserAdminRoleSuperAdmin || u.AdminRole == UserAdminRoleSalesAdmin {
		return true
	}

	// make sure we're checking current user orgs
	if err := tx.Load(u, "UserOrganizations"); err != nil {
		return false
	}

	for _, uo := range u.UserOrganizations {
		if uo.OrganizationID == orgId && uo.Role == UserOrganizationRoleAdmin {
			return true
		}
	}

	return false
}

// CanViewOrganization returns true if the given user is allowed to view the specified organization
func (u *User) CanViewOrganization(tx *pop.Connection, orgId int) bool {
	// if user is a system admin, allow
	if u.AdminRole == UserAdminRoleSuperAdmin || u.AdminRole == UserAdminRoleSalesAdmin {
		return true
	}

	// make sure we're checking current user orgs
	if err := tx.Load(u, "UserOrganizations"); err != nil {
		return false
	}

	for _, uo := range u.UserOrganizations {
		if uo.OrganizationID == orgId && uo.Role == UserOrganizationRoleAdmin {
			return true
		}
	}

	return false
}

func (u *User) CanEditOrganization(tx *pop.Connection, orgId int) bool {
	// if user is a system admin, allow
	if u.AdminRole == UserAdminRoleSuperAdmin || u.AdminRole == UserAdminRoleSalesAdmin {
		return true
	}

	// make sure we're checking current user orgs
	if err := tx.Load(u, "UserOrganizations"); err != nil {
		return false
	}

	for _, uo := range u.UserOrganizations {
		if uo.OrganizationID == orgId && uo.Role == UserOrganizationRoleAdmin {
			return true
		}
	}

	return false
}

// canEditAllRequests indicates whether the user is allowed to edit all requests.
func (u *User) canEditAllRequests() bool {
	return u.AdminRole == UserAdminRoleSuperAdmin
}

// CanUpdateRequestStatus indicates whether the user is allowed to change the request status.
func (u *User) CanUpdateRequestStatus(request Request, newStatus RequestStatus) bool {
	if u.AdminRole == UserAdminRoleSuperAdmin {
		return true
	}

	// others can only make limited changes
	return request.canUserChangeStatus(*u, newStatus)
}

func (u *User) CanViewRequest(tx *pop.Connection, request Request) bool {
	if u.AdminRole == UserAdminRoleSuperAdmin {
		return true
	}

	if request.Visibility == RequestVisibilityAll {
		return true
	}

	// request creator can view it
	if u.ID == request.CreatedByID {
		return true
	}

	// If the user has a matching org, then yes
	uOrgIDs := u.GetOrgIDs(tx)
	for _, oID := range uOrgIDs {
		if oID == request.OrganizationID {
			return true
		}
	}

	// No matching or, so no if not open to trusted orgs
	if request.Visibility == RequestVisibilitySame {
		return false
	}

	// Check Trusted Orgs
	var orgTrust OrganizationTrust
	for _, oID := range uOrgIDs {
		err := orgTrust.FindByOrgIDs(tx, request.OrganizationID, oID)
		if err == nil {
			return true
		}
	}

	return false
}

// FindByUUID find a User with the given UUID and loads it from the database.
func (u *User) FindByUUID(tx *pop.Connection, uuid string) error {
	if uuid == "" {
		return errors.New("error: uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", uuid).First(u); err != nil {
		return fmt.Errorf("error finding user by uuid: %s", err.Error())
	}

	return nil
}

// FindByID finds a User with a given ID and loads it from the database
func (u *User) FindByID(tx *pop.Connection, id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding user: id must be a positive number")
	}

	if err := tx.Eager(eagerFields...).Find(u, id); err != nil {
		return fmt.Errorf("error finding user by id: %v, ... %v", id, err.Error())
	}

	return nil
}

// FindByEmail finds a User with a matching email
func (u *User) FindByEmail(tx *pop.Connection, email string) error {
	if err := tx.Where("email = ?", email).First(u); err != nil {
		return fmt.Errorf("error finding user by email: %s, ... %s",
			email, err.Error())
	}

	return nil
}

// FindByEmailAndSocialAuthProvider finds a User with a matching email and social_auth_provider
func (u *User) FindByEmailAndSocialAuthProvider(tx *pop.Connection, email, authProvider string) error {
	err := tx.Where("email = ? and social_auth_provider = ?", email, authProvider).First(u)
	if err != nil {
		return fmt.Errorf("error finding user by email and auth provider: %s, %s, ... %s",
			email, authProvider, err.Error())
	}

	return nil
}

// FindByIDs finds all Users associated with the given IDs and loads them from the database
func (u *Users) FindByIDs(tx *pop.Connection, ids []int) error {
	ids = domain.UniquifyIntSlice(ids)
	return tx.Where("id in (?)", ids).All(u)
}

// HashClientIdAccessToken just returns a sha256.Sum256 of the input value
func HashClientIdAccessToken(accessToken string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(accessToken)))
}

func (u *User) GetOrganizations(tx *pop.Connection) (Organizations, error) {
	if err := tx.Load(u, "Organizations"); err != nil {
		return nil, fmt.Errorf("error getting organizations for user id %v ... %v", u.ID, err)
	}

	return u.Organizations, nil
}

func (u *User) FindUserOrganization(tx *pop.Connection, org Organization) (UserOrganization, error) {
	var userOrg UserOrganization
	if err := tx.Where("user_id = ? AND organization_id = ?", u.ID, org.ID).First(&userOrg); err != nil {
		return UserOrganization{}, fmt.Errorf("association not found for user '%v' and org '%v' (%s)", u.Nickname, org.Name, err.Error())
	}

	return userOrg, nil
}

// AttachPhoto assigns a previously-stored File to this User as a profile photo
func (u *User) AttachPhoto(tx *pop.Connection, fileID string) (File, error) {
	return addFile(tx, u, fileID)
}

// RemoveFile removes an attached file from the User profile
func (u *User) RemoveFile(tx *pop.Connection) error {
	return removeFile(tx, u)
}

// GetPhotoID retrieves the UUID of the User's photo file
func (u *User) GetPhotoID(tx *pop.Connection) (*string, error) {
	if err := tx.Load(u, "PhotoFile"); err != nil {
		return nil, err
	}
	if u.FileID.Valid {
		photoID := u.PhotoFile.UUID.String()
		return &photoID, nil
	}

	return nil, nil
}

// GetPhotoURL retrieves the photo URL from the attached file
func (u *User) GetPhotoURL(tx *pop.Connection) (*string, error) {
	if err := tx.Load(u, "PhotoFile"); err != nil {
		return nil, err
	}

	if !u.FileID.Valid {
		if u.AuthPhotoURL.Valid {
			return &u.AuthPhotoURL.String, nil
		}
		url := gravatarURL(u.Email)
		return &url, nil
	}

	if err := u.PhotoFile.RefreshURL(tx); err != nil {
		return nil, err
	}
	return &u.PhotoFile.URL, nil
}

// Save wraps tx.Save() call to check for errors and operate on attached object
func (u *User) Save(tx *pop.Connection) error {
	u.Nickname = domain.RemoveUnwantedChars(u.Nickname, "-_ .,'&@")
	if err := save(tx, u); err != nil {
		appError := api.AppError{Err: err}
		if strings.Contains(err.Error(), "Nickname must have a visible character") {
			appError.Key = api.ErrorUserInvisibleNickname
			appError.Category = api.CategoryUser
		} else if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_nickname_idx"`) {
			appError.Key = api.ErrorUserDuplicateNickname
			appError.Category = api.CategoryUser
		} else {
			appError.Key = api.ErrorUserUpdate
			appError.Category = api.CategoryInternal
		}
		return &appError
	}
	return nil
}

func (u *User) uniquifyNickname(tx *pop.Connection, prefixes [30]string) error {
	simpleNN := u.Nickname
	if simpleNN == "" {
		simpleNN = u.FirstName
		if len(u.LastName) > 0 {
			simpleNN = u.FirstName + " " + u.LastName[:1]
		}
	}

	var err error

	// Use the first nickname prefix that makes it unique
	for _, p := range prefixes {
		u.Nickname = p + " " + simpleNN

		var existingUser User
		err = tx.Where("nickname = ?", u.Nickname).First(&existingUser)

		// We didn't find a match, so we're good with the current nickname
		if existingUser.Nickname == "" {
			return nil
		}

	}

	if err != nil {
		return fmt.Errorf("last error looking for unique nickname for existingUser %v ... %v", u.UUID, err)
	}

	return fmt.Errorf("failed finding unique nickname for user %s %s", u.FirstName, u.LastName)
}

// GetLocation reads the location record, if it exists, and returns the Location object.
func (u *User) GetLocation(tx *pop.Connection) (*Location, error) {
	if !u.LocationID.Valid {
		return nil, nil
	}
	location := Location{}
	if err := tx.Find(&location, u.LocationID); err != nil {
		return nil, err
	}

	return &location, nil
}

// SetLocation sets the user location fields, creating a new record in the database if necessary.
func (u *User) SetLocation(tx *pop.Connection, location Location) error {
	if u.LocationID.Valid {
		location.ID = u.LocationID.Int
		u.Location = location
		return u.Location.Update(tx)
	}
	if err := location.Create(tx); err != nil {
		return err
	}
	u.LocationID = nulls.NewInt(location.ID)
	return u.Save(tx)
}

// RemoveLocation removes the location record associated with the user
func (u *User) RemoveLocation(tx *pop.Connection) error {
	if !u.LocationID.Valid {
		return nil
	}

	if err := tx.Destroy(&Location{ID: u.LocationID.Int}); err != nil {
		return err
	}
	u.LocationID = nulls.Int{}
	// don't need to save the user because the database foreign key constraint is set to "ON DELETE SET NULL"
	return nil
}

type UnreadThread struct {
	ThreadUUID uuid.UUID
	Count      int
}

// UnreadMessageCount returns an entry for each thread that has other users' messages
// that have not yet been read by this this user.
func (u *User) UnreadMessageCount(tx *pop.Connection) ([]UnreadThread, error) {
	emptyUnreads := []UnreadThread{}

	threadPs := ThreadParticipants{}
	if err := tx.Eager("Thread").Where("user_id = ?", u.ID).All(&threadPs); err != nil {
		return emptyUnreads, err
	}

	unreads := []UnreadThread{}

	for _, tp := range threadPs {
		msgCount, err := tp.Thread.GetUnreadMessageCount(tx, u.ID, tp.LastViewedAt)
		if err != nil {
			log.Errorf("error getting count of unread messages for thread %s ... %v",
				tp.Thread.UUID, err)
			continue
		}

		if msgCount > 0 {
			unreads = append(unreads, UnreadThread{ThreadUUID: tp.Thread.UUID, Count: msgCount})
		}
	}

	return unreads, nil
}

// GetThreads finds all threads that the user is participating in.
func (u *User) GetThreads(tx *pop.Connection) (Threads, error) {
	var t Threads
	query := tx.Q().
		LeftJoin("thread_participants tp", "threads.id = tp.thread_id").
		Where("tp.user_id = ?", u.ID).
		Order("updated_at desc")
	if err := query.All(&t); err != nil {
		return nil, err
	}

	return t, nil
}

// GetThreadsForConversations finds all threads that the user is participating in.
func (u *User) GetThreadsForConversations(tx *pop.Connection) (Threads, error) {
	t := Threads{}
	query := tx.Q().
		LeftJoin("thread_participants tp", "threads.id = tp.thread_id").
		Where("tp.user_id = ?", u.ID).
		Order("updated_at desc")
	if err := query.All(&t); err != nil {
		return nil, err
	}

	for i := range t {
		err := t[i].LoadForAPI(tx, *u)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

// WantsRequestNotification answers the question "Does the user want notifications for this request?"
func (u *User) WantsRequestNotification(tx *pop.Connection, request Request) bool {
	if request.CreatedByID == u.ID {
		return false
	}

	if u.isNearRequest(tx, request) {
		return true
	}

	return u.hasMatchingWatch(tx, request)
}

func (u *User) isNearRequest(tx *pop.Connection, request Request) bool {
	if err := tx.Load(u, "Location"); err != nil {
		log.Errorf("load of user location failed, %s", err)
		return false
	}

	requestOrigin, err := request.GetOrigin(tx)
	if err != nil {
		log.Errorf("failed to get request origin, %s", err)
		return false
	}
	if requestOrigin == nil {
		return false
	}

	if u.Location.IsNear(*requestOrigin) {
		return true
	}
	return false
}

func (u *User) hasMatchingWatch(tx *pop.Connection, request Request) bool {
	watches := Watches{}
	if err := watches.FindByUser(tx, *u); err != nil {
		log.Errorf("failed to get watch list, %s", err)
		return false
	}
	for _, watch := range watches {
		if watch.matchesRequest(tx, request) {
			return true
		}
	}

	return false
}

// GetPreferences returns a StandardPreferences struct
func (u *User) GetPreferences(tx *pop.Connection) (StandardPreferences, error) {
	if err := tx.Load(u, "UserPreferences"); err != nil {
		err := errors.New("error getting user preferences ... " + err.Error())
		return StandardPreferences{}, err
	}

	dbPreferences := map[string]string{}

	// Build up a map of the User's Preferences in the database while also
	// checking that they are each allowed
	for _, uP := range u.UserPreferences {
		_, ok := allowedUserPreferenceKeys[uP.Key]
		if !ok {
			log.Errorf("the database included a user preference with an unexpected key %s", uP.Key)
			continue
		}
		dbPreferences[uP.Key] = uP.Value
	}

	finalValues := map[string]string{}

	fieldAndValidators := getPreferencesFieldsAndValidators(StandardPreferences{})
	for fieldName, fV := range fieldAndValidators {
		if value, ok := dbPreferences[fieldName]; ok {
			if fV.validator(value) {
				finalValues[fieldName] = value
			} else {
				log.Errorf("user preference %s in database not allowed ... %s", fieldName, value)
			}
		}
	}

	var finalPreferences StandardPreferences
	finalPreferences.hydrateValues(finalValues)

	return finalPreferences, nil
}

// UpdateStandardPreferences validates and updates a user's standard preferences
func (u *User) UpdateStandardPreferences(tx *pop.Connection, prefs StandardPreferences) (StandardPreferences, error) {
	if err := updateUsersStandardPreferences(tx, *u, prefs); err != nil {
		return StandardPreferences{}, err
	}

	return u.GetPreferences(tx)
}

func (u User) GetLanguagePreference(tx *pop.Connection) string {
	prefs, err := u.GetPreferences(tx)
	if err != nil || prefs.Language == "" {
		return domain.UserPreferenceLanguageEnglish
	}

	return prefs.Language
}

// GetRealName returns the real name, first and last, of the user
func (u *User) GetRealName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

// HasOrganization returns true if the user has one or more organization connections
func (u *User) HasOrganization(tx *pop.Connection) bool {
	var c Count
	err := tx.RawQuery("SELECT COUNT(*) FROM user_organizations WHERE user_id = ?", u.ID).First(&c)
	if err != nil {
		log.Errorf("error counting user organizations, user = '%s', err = %s", u.UUID, err)
		return false
	}
	if c.N == 0 {
		return false
	}
	return true
}

func (u *User) isSuperAdmin() bool {
	return u.AdminRole == UserAdminRoleSuperAdmin
}

// MeetingsAsParticipant returns all meetings in which the user is a participant
func (u *User) MeetingsAsParticipant(tx *pop.Connection) ([]Meeting, error) {
	m := Meetings{}
	if err := tx.
		Where("meeting_participants.user_id=?", u.ID).
		Join("meeting_participants", "meeting_participants.meeting_id=meetings.id").
		All(&m); err != nil {
		return m, err
	}
	return m, nil
}

func (u *User) CanCreateMeetingInvite(tx *pop.Connection, meeting Meeting) (bool, error) {
	if u.ID == meeting.CreatedByID || u.isSuperAdmin() {
		return true, nil
	}

	return meeting.isOrganizer(tx, u.ID)
}

func (u *User) CanUpdateMeeting(meeting Meeting) bool {
	return u.ID == meeting.CreatedByID || u.isSuperAdmin()
}

func (u *User) CanCreateMeetingParticipant(tx *pop.Connection, meeting Meeting) bool {
	return u.ID == meeting.CreatedByID || meeting.isVisible(tx, u.ID) || u.isSuperAdmin()
}

func (u *User) CanViewMeetingParticipants(tx *pop.Connection, meeting Meeting) (bool, error) {
	if u.ID == meeting.CreatedByID || u.isSuperAdmin() {
		return true, nil
	}

	return meeting.isOrganizer(tx, u.ID)
}

// RemovePreferences removes all of the users's preferences
func (u *User) RemovePreferences(tx *pop.Connection) error {
	if u == nil || u.ID < 1 {
		return nil
	}
	var p UserPreference
	return p.removeAll(tx, u.ID)
}

func ConvertUserPrivate(ctx context.Context, user User) (api.UserPrivate, error) {
	tx := Tx(ctx)

	output := api.UserPrivate{}
	if err := api.ConvertToOtherType(user, &output); err != nil {
		return api.UserPrivate{}, err
	}
	output.ID = user.UUID

	photoURL, err := user.GetPhotoURL(tx)
	if err != nil {
		return api.UserPrivate{}, err
	}

	if photoURL != nil {
		output.AvatarURL = nulls.NewString(*photoURL)
	}

	if user.FileID.Valid {
		// depends on the earlier call to GetPhotoURL to hydrate PhotoFile
		output.PhotoID = nulls.NewUUID(user.PhotoFile.UUID)
	}

	organizations, err := user.GetOrganizations(tx)
	if err != nil {
		return api.UserPrivate{}, err
	}
	output.Organizations = ConvertOrganizations(organizations)
	return output, nil
}

// ConvertUsers converts list of models.User to list of api.User
func ConvertUsers(ctx context.Context, users Users) (api.Users, error) {
	output := make(api.Users, len(users))
	for i := range output {
		var err error
		output[i], err = ConvertUser(ctx, users[i])
		if err != nil {
			return output, err
		}
	}
	return output, nil
}

// ConvertUsers converts models.User to api.User
func ConvertUser(ctx context.Context, user User) (api.User, error) {
	tx := Tx(ctx)

	output := api.User{}
	if err := api.ConvertToOtherType(user, &output); err != nil {
		return api.User{}, err
	}
	output.ID = user.UUID

	photoURL, err := user.GetPhotoURL(tx)
	if err != nil {
		return api.User{}, err
	}

	if photoURL != nil {
		output.AvatarURL = nulls.NewString(*photoURL)
	}

	return output, nil
}
