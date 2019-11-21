package models

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/events"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/domain"
)

const (
	PostRoleCreatedby string = "PostsCreated"
	PostRoleReceiving string = "PostsReceiving"
	PostRoleProviding string = "PostsProviding"
)

type User struct {
	ID                int               `json:"id" db:"id"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
	Email             string            `json:"email" db:"email"`
	FirstName         string            `json:"first_name" db:"first_name"`
	LastName          string            `json:"last_name" db:"last_name"`
	Nickname          string            `json:"nickname" db:"nickname"`
	AdminRole         nulls.String      `json:"admin_role" db:"admin_role"`
	Uuid              uuid.UUID         `json:"uuid" db:"uuid"`
	PhotoFileID       nulls.Int         `json:"photo_file_id" db:"photo_file_id"`
	PhotoURL          nulls.String      `json:"photo_url" db:"photo_url"`
	LocationID        nulls.Int         `json:"location_id" db:"location_id"`
	AccessTokens      []UserAccessToken `has_many:"user_access_tokens" json:"-"`
	Organizations     Organizations     `many_to_many:"user_organizations" order_by:"name asc" json:"-"`
	UserOrganizations UserOrganizations `has_many:"user_organizations" json:"-"`
	UserPreferences   UserPreferences   `has_many:"user_preferences" json:"-"`
	PostsCreated      Posts             `has_many:"posts" fk_id:"created_by_id" order_by:"updated_at desc"`
	PostsProviding    Posts             `has_many:"posts" fk_id:"provider_id" order_by:"updated_at desc"`
	PostsReceiving    Posts             `has_many:"posts" fk_id:"receiver_id" order_by:"updated_at desc"`
	PhotoFile         File              `belongs_to:"files"`
	Location          Location          `belongs_to:"locations"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.FirstName, Name: "FirstName"},
		&validators.StringIsPresent{Field: u.LastName, Name: "LastName"},
		&validators.StringIsPresent{Field: u.Nickname, Name: "Nickname"},
		&validators.UUIDIsPresent{Field: u.Uuid, Name: "Uuid"},
		&NullsStringIsURL{Field: u.PhotoURL, Name: "PhotoURL"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// All retrieves all Users from the database.
func (u *Users) All(selectFields ...string) error {
	return DB.Select(selectFields...).
		Order("nickname asc").
		All(u)
}

// CreateAccessToken - Create and store new UserAccessToken
func (u *User) CreateAccessToken(org Organization, clientID string) (string, int64, error) {
	if clientID == "" {
		return "", 0, fmt.Errorf("cannot create token with empty clientID for user %s", u.Nickname)
	}

	token, _ := getRandomToken()
	hash := HashClientIdAccessToken(clientID + token)
	expireAt := createAccessTokenExpiry()

	userOrg, err := u.FindUserOrganization(org)
	if err != nil {
		return "", 0, err
	}

	userAccessToken := &UserAccessToken{
		UserID:             u.ID,
		UserOrganizationID: userOrg.ID,
		AccessToken:        hash,
		ExpiresAt:          expireAt,
	}

	if err := DB.Save(userAccessToken); err != nil {
		return "", 0, err
	}

	return token, expireAt.UTC().Unix(), nil
}

func (u *User) GetOrgIDs() []int {
	// ignore the error and allow the user's Organizations to be an empty slice.
	_ = DB.Load(u, "Organizations")

	s := make([]int, len(u.Organizations))
	for i, v := range u.Organizations {
		s[i] = v.ID
	}

	return s
}

func (u *User) FindOrCreateFromAuthUser(orgID int, authUser *auth.User) error {
	var userOrgs UserOrganizations
	err := userOrgs.FindByAuthEmail(authUser.Email, orgID)
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
		err = DB.Where("uuid = ?", userOrgs[0].User.Uuid).First(u)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	newUser := true
	if u.ID != 0 {
		newUser = false
	}

	// update attributes from authUser
	u.FirstName = authUser.FirstName
	u.LastName = authUser.LastName
	u.Email = authUser.Email

	if authUser.PhotoURL != "" {
		u.PhotoURL = nulls.NewString(authUser.PhotoURL)
	} else {
		// ref: https://en.gravatar.com/site/implement/images/
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(authUser.Email))))
		u.PhotoURL = nulls.NewString(fmt.Sprintf("https://www.gravatar.com/avatar/%x.jpg?s=200&d=mp", hash))
	}

	// if new user they will need a uuid and a unique Nickname
	if newUser {
		u.Uuid = domain.GetUuid()

		u.Nickname = authUser.Nickname
		if err := u.uniquifyNickname(); err != nil {
			return err
		}
	}

	err = DB.Save(u)
	if err != nil {
		return fmt.Errorf("unable to create new user record: %s", err.Error())
	}

	if len(userOrgs) == 0 {
		userOrg := &UserOrganization{
			OrganizationID: orgID,
			UserID:         u.ID,
			Role:           UserOrganizationRoleUser,
			AuthID:         authUser.UserID,
			AuthEmail:      u.Email,
			LastLogin:      time.Now(),
		}
		err = DB.Save(userOrg)
		if err != nil {
			return fmt.Errorf("unable to create new user_organization record: %s", err.Error())
		}
	}

	if newUser {
		e := events.Event{
			Kind:    domain.EventApiUserCreated,
			Message: "Nickname: " + u.Nickname + "  Uuid: " + u.Uuid.String(),
			Payload: events.Payload{"user": u},
		}
		emitEvent(e)
	}

	// reload user
	// err = DB.Eager().Where("id = ?", u.ID).First(u)
	// if err != nil {
	// 	return fmt.Errorf("unable to reload user after update: %s", err)
	// }

	return nil
}

// CanCreateOrganization returns true if the given user is allowed to create organizations
func (u *User) CanCreateOrganization() bool {
	return u.AdminRole.String == domain.AdminRoleSuperDuperAdmin || u.AdminRole.String == domain.AdminRoleSalesAdmin
}

func (u *User) CanEditOrganization(orgId int) bool {
	// make sure we're checking current user orgs
	err := DB.Load(u, "UserOrganizations")
	if err != nil {
		return false
	}

	for _, uo := range u.UserOrganizations {
		if uo.OrganizationID == orgId && uo.Role == UserOrganizationRoleAdmin {
			return true
		}
	}

	return false
}

// CanEditAllPosts indicates whether the user is allowed to edit all posts.
func (u *User) CanEditAllPosts() bool {
	return u.AdminRole.String == domain.AdminRoleSuperDuperAdmin
}

// CanUpdatePostStatus indicates whether the user is allowed to change the post status.
func (u *User) CanUpdatePostStatus(post Post, newStatus string) bool {
	if u.AdminRole.String == domain.AdminRoleSuperDuperAdmin {
		return true
	}

	// post creator can make any status changes
	if u.ID == post.CreatedByID {
		return true
	}

	// others can only make limited changes
	return post.canUserChangeStatus(*u, newStatus)
}

// FindByUUID find a User with the given UUID and loads it from the database.
func (u *User) FindByUUID(uuid string, selectFields ...string) error {
	if uuid == "" {
		return errors.New("error: uuid must not be blank")
	}

	if err := DB.Select(selectFields...).Where("uuid = ?", uuid).First(u); err != nil {
		return fmt.Errorf("error finding user by uuid: %s", err.Error())
	}

	return nil
}

func (u *User) FindByID(id int, eagerFields ...string) error {
	if id <= 0 {
		return errors.New("error finding user: id must a positive number")
	}

	if err := DB.Eager(eagerFields...).Find(u, id); err != nil {
		return fmt.Errorf("error finding user by id: %v, ... %v", id, err.Error())
	}

	return nil
}

// HashClientIdAccessToken just returns a sha256.Sum256 of the input value
func HashClientIdAccessToken(accessToken string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(accessToken)))
}

func (u *User) GetOrganizations() ([]Organization, error) {
	if err := DB.Load(u, "Organizations"); err != nil {
		return nil, fmt.Errorf("error getting organizations for user id %v ... %v", u.ID, err)
	}

	return u.Organizations, nil
}

func (u *User) FindUserOrganization(org Organization) (UserOrganization, error) {
	var userOrg UserOrganization
	if err := DB.Where("user_id = ? AND organization_id = ?", u.ID, org.ID).First(&userOrg); err != nil {
		return UserOrganization{}, fmt.Errorf("association not found for user '%v' and org '%v' (%s)", u.Nickname, org.Name, err.Error())
	}

	return userOrg, nil
}

func (u *User) GetPosts(postRole string) ([]Post, error) {
	if err := DB.Load(u, postRole); err != nil {
		return nil, fmt.Errorf("error getting posts for user id %v ... %v", u.ID, err)
	}

	var posts Posts
	switch postRole {
	case PostRoleCreatedby:
		posts = u.PostsCreated

	case PostRoleReceiving:
		posts = u.PostsReceiving

	case PostRoleProviding:
		posts = u.PostsProviding
	}

	return posts, nil
}

// AttachPhoto assigns a previously-stored File to this User as a profile photo
func (u *User) AttachPhoto(fileID string) (File, error) {
	var f File
	if err := f.FindByUUID(fileID); err != nil {
		return f, err
	}

	u.PhotoFileID = nulls.NewInt(f.ID)
	if err := DB.Save(u); err != nil {
		return f, err
	}

	return f, nil
}

// GetPhotoURL retrieves the photo URL, either from the photo_url database field, or from the attached file
func (u *User) GetPhotoURL() (string, error) {
	if err := DB.Load(u, "PhotoFile"); err != nil {
		return "", err
	}

	url := u.PhotoURL.String
	if url == "" {
		if !u.PhotoFileID.Valid {
			return "", nil
		}
		if err := u.PhotoFile.RefreshURL(); err != nil {
			return "", err
		}
		url = u.PhotoFile.URL
	}
	return url, nil
}

// Save wraps DB.Save() call to check for errors and operate on attached object
func (u *User) Save() error {
	validationErrs, err := u.Validate(DB)
	if validationErrs != nil && validationErrs.HasAny() {
		return errors.New(FlattenPopErrors(validationErrs))
	}
	if err != nil {
		return err
	}

	return DB.Save(u)
}

func (u *User) uniquifyNickname() error {

	simpleNN := u.Nickname
	if simpleNN == "" {
		simpleNN = u.FirstName
		if len(u.LastName) > 0 {
			simpleNN = u.FirstName + u.LastName[:1]
		}
	}

	var err error

	// User the first nickname prefix that makes it unique
	for _, p := range allPrefixes() {
		u.Nickname = p + simpleNN

		var existingUser User
		err = DB.Where("nickname = ?", u.Nickname).First(&existingUser)

		// We didn't find a match, so we're good with the current nickname
		if existingUser.Nickname == "" {
			return nil
		}

	}

	if err != nil {
		return fmt.Errorf("last error looking for unique nickname for existingUser %v ... %v", u.Uuid, err)
	}

	return fmt.Errorf("failed finding unique nickname for user %s %s", u.FirstName, u.LastName)
}

// GetLocation reads the location record, if it exists, and returns the Location object.
func (u *User) GetLocation() (*Location, error) {
	if !u.LocationID.Valid {
		return nil, nil
	}
	location := Location{}
	if err := DB.Find(&location, u.LocationID); err != nil {
		return nil, err
	}

	return &location, nil
}

// SetLocation sets the user location fields, creating a new record in the database if necessary.
func (u *User) SetLocation(location Location) error {
	if u.LocationID.Valid {
		location.ID = u.LocationID.Int
		u.Location = location
		return DB.Update(&u.Location)
	}
	if err := DB.Create(&location); err != nil {
		return err
	}
	u.LocationID = nulls.NewInt(location.ID)
	return DB.Update(u)
}

type UnreadThread struct {
	ThreadUUID uuid.UUID
	Count      int
}

// UnreadMessageCount returns an entry for each thread that has other users' messages
// that have not yet been read by this this user.
func (u *User) UnreadMessageCount() ([]UnreadThread, error) {
	emptyUnreads := []UnreadThread{}

	threadPs := ThreadParticipants{}
	if err := DB.Eager("Thread").Where("user_id = ?", u.ID).All(&threadPs); err != nil {
		return emptyUnreads, err
	}

	unreads := []UnreadThread{}

	for _, tp := range threadPs {
		msgCount, err := tp.Thread.UnreadMessageCount(u.ID, tp.LastViewedAt)
		if err != nil {
			domain.ErrLogger.Printf("error getting count of unread messages for thread %s ... %v",
				tp.Thread.Uuid, err)
			continue
		}

		if msgCount > 0 {
			unreads = append(unreads, UnreadThread{ThreadUUID: tp.Thread.Uuid, Count: msgCount})
		}
	}

	return unreads, nil
}

// GetThreads finds all threads that the user is participating in.
func (u *User) GetThreads() (Threads, error) {
	var t Threads
	query := DB.Q().
		LeftJoin("thread_participants tp", "threads.id = tp.thread_id").
		Where("tp.user_id = ?", u.ID).
		Order("updated_at desc")
	if err := query.All(&t); err != nil {
		return nil, err
	}

	return t, nil
}

// WantsPostNotification answers the question "Does the user want notifications for this post?"
func (u *User) WantsPostNotification(post Post) bool {
	if post.CreatedByID == u.ID {
		return false
	}

	if err := DB.Load(u, "Location"); err != nil {
		domain.ErrLogger.Printf("load of user location failed, %s", err)
		return false
	}

	postLocation, err := post.GetLocationForNotifications()
	if err != nil {
		domain.ErrLogger.Print(err.Error())
		return false
	}

	if !u.Location.IsNear(*postLocation) {
		return false
	}

	return true
}

// GetPreferences returns a slice of matching UserPreference records
func (u *User) GetPreferences() (UserPreferences, error) {
	uPrefs := UserPreferences{}
	err := DB.Where("user_id = ?", u.ID).All(&uPrefs)

	return uPrefs, err
}

// GetPreference returns a pointer to a matching UserPreference record or if
// none is found, returns nil
func (u *User) GetPreference(key string) (*UserPreference, error) {
	uPref := UserPreference{}

	err := DB.Where("user_id = ?", u.ID).Where("key = ?", key).First(&uPref)
	if err != nil {
		if domain.IsOtherThanNoRows(err) {
			return nil, err
		}
		return nil, nil
	}

	return &uPref, nil
}

func (u *User) CreatePreference(key, value string) (UserPreference, error) {
	uPref := UserPreference{}

	if err := u.FindByUUID(u.Uuid.String(), "id"); err != nil {
		err := fmt.Errorf("can't create UserPreference - error finding user with uuid %s ... %v",
			u.Uuid.String(), err)
		return UserPreference{}, err
	}

	DB.Where("user_id = ?", u.ID).Where("key = ?", key).First(&uPref)
	if uPref.ID > 0 {
		err := fmt.Errorf("can't create UserPreference with key %s.  Already exists with id %v.", key, uPref.ID)
		return UserPreference{}, err
	}

	uPref.Uuid = domain.GetUuid()
	uPref.UserID = u.ID
	uPref.Key = key
	uPref.Value = value

	if err := DB.Save(&uPref); err != nil {
		return UserPreference{}, err
	}

	return uPref, nil
}

// UpdatePreferenceByKey will also create a new instance, if a match is not found for that user
func (u *User) UpdatePreferenceByKey(key, value string) (UserPreference, error) {

	uPref, err := u.GetPreference(key)
	if err != nil {
		return UserPreference{}, err
	}

	if uPref == nil {
		return u.CreatePreference(key, value)
	}

	if uPref.Value == value {
		return *uPref, nil
	}

	uPref.Value = value

	if err := DB.Save(uPref); err != nil {
		return UserPreference{}, err
	}

	return *uPref, nil
}

// UpdatePreferencesByKey will also create new instances for preferences that don't exist for that user
func (u *User) UpdatePreferencesByKey(keyVals [][2]string) (UserPreferences, error) {
	uPrefs := UserPreferences{}

	for _, keyVal := range keyVals {
		uP, err := u.UpdatePreferenceByKey(keyVal[0], keyVal[1])
		if err != nil {
			return UserPreferences{}, err
		}
		uPrefs = append(uPrefs, uP)
	}

	return uPrefs, nil
}
