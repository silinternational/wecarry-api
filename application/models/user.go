package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/silinternational/handcarry-api/domain"

	"github.com/silinternational/handcarry-api/auth"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"

	"github.com/gofrs/uuid"
)

type User struct {
	ID            int               `json:"id" db:"id"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
	Email         string            `json:"email" db:"email"`
	FirstName     string            `json:"first_name" db:"first_name"`
	LastName      string            `json:"last_name" db:"last_name"`
	Nickname      string            `json:"nickname" db:"nickname"`
	AdminRole     nulls.String      `json:"admin_role" db:"admin_role"`
	Uuid          uuid.UUID         `json:"uuid" db:"uuid"`
	AccessTokens  []UserAccessToken `has_many:"user_access_tokens"`
	Organizations Organizations     `many_to_many:"user_organizations"`
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
		&validators.IntIsPresent{Field: u.ID, Name: "ID"},
		&validators.StringIsPresent{Field: u.Email, Name: "Email"},
		&validators.StringIsPresent{Field: u.FirstName, Name: "FirstName"},
		&validators.StringIsPresent{Field: u.LastName, Name: "LastName"},
		&validators.StringIsPresent{Field: u.Nickname, Name: "Nickname"},
		&validators.UUIDIsPresent{Field: u.Uuid, Name: "Uuid"},
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

// CreateAccessToken - Create and store new UserAccessToken
func (u *User) CreateAccessToken(orgID int, clientID string) (string, int64, error) {

	token := createAccessTokenPart()
	hash := hashClientIdAccessToken(clientID + token)
	expireAt := createAccessTokenExpiry()

	userAccessToken := &UserAccessToken{
		UserID:              u.ID,
		UserOrganizationsID: orgID,
		AccessToken:         hash,
		ExpiresAt:           expireAt,
	}

	err := DB.Save(userAccessToken)
	if err != nil {
		return "", 0, err
	}

	return token, expireAt.UTC().Unix(), nil
}

func (u *User) GetOrgIDs() []interface{} {
	var ids []int
	for _, uo := range u.Organizations {
		ids = append(ids, uo.ID)
	}

	s := make([]interface{}, len(ids))
	for i, v := range ids {
		s[i] = v
	}

	return s
}

func (u *User) FindOrCreateFromAuthUser(orgID int, authUser *auth.User) error {

	userOrgs, err := UserOrganizationFindByAuthEmail(authUser.Email, orgID)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(userOrgs) > 1 {
		return fmt.Errorf("too many user organizations found (%v), data integrity problem", len(userOrgs))
	}

	if len(userOrgs) == 1 {
		if userOrgs[0].AuthID != authUser.UserID {
			return fmt.Errorf("a user in this organization with this email address already exists with different user id")
		}
		err = DB.Where("uuid = ?", userOrgs[0].User.Uuid).First(u)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var newUser bool
	if u.ID != 0 {
		newUser = false
	}

	// update attributes from authUser
	u.FirstName = authUser.FirstName
	u.LastName = authUser.LastName
	u.Email = authUser.Email
	u.Nickname = fmt.Sprintf("%s %s", authUser.FirstName, authUser.LastName[:0])

	// if new user they will need a uuid
	if newUser {
		u.Uuid = domain.GetUuid()
	}

	err = DB.Save(u)
	if err != nil {
		return fmt.Errorf("unable to create new user record: %s", err.Error())
	}

	if len(userOrgs) == 0 {
		userOrg := &UserOrganization{
			OrganizationID: orgID,
			UserID:         u.ID,
			Role:           UserOrganizationRoleMember,
			AuthID:         authUser.UserID,
			AuthEmail:      u.Email,
			LastLogin:      time.Now(),
		}
		err = DB.Save(userOrg)
		if err != nil {
			return fmt.Errorf("unable to create new user_organization record: %s", err.Error())
		}
	}

	// reload user
	// err = DB.Eager().Where("id = ?", u.ID).First(u)
	// if err != nil {
	// 	return fmt.Errorf("unable to reload user after update: %s", err)
	// }

	return nil
}

func FindUserByAccessToken(accessToken string) (User, error) {

	userAccessToken := UserAccessToken{}

	if accessToken == "" {
		return User{}, fmt.Errorf("error: access token must not be blank")
	}

	dbAccessToken := hashClientIdAccessToken(accessToken)
	// and expires_at > now()
	if err := DB.Eager().Where("access_token = ?", dbAccessToken).First(&userAccessToken); err != nil {
		return User{}, fmt.Errorf("error finding user by access token: %s", err.Error())
	}

	if userAccessToken.ID == 0 {
		return User{}, fmt.Errorf("error finding user by access token")
	}

	if userAccessToken.ExpiresAt.Before(time.Now()) {
		err := DB.Destroy(&userAccessToken)
		if err != nil {
			log.Printf("Unable to delete expired userAccessToken, id: %v", userAccessToken.ID)
		}
		return User{}, fmt.Errorf("access token has expired")
	}

	err := DB.Load(&userAccessToken.User)
	if err != nil {
		log.Printf("unable to eagerly load all associations for user: %s", err.Error())
	}

	return userAccessToken.User, nil
}

func FindUserByUUID(uuid string) (User, error) {

	if uuid == "" {
		return User{}, fmt.Errorf("error: uuid must not be blank")
	}

	user := User{}
	queryString := fmt.Sprintf("uuid = '%s'", uuid)

	if err := DB.Where(queryString).First(&user); err != nil {
		return User{}, fmt.Errorf("error finding user by uuid: %s", err.Error())
	}

	return user, nil
}

func createAccessTokenExpiry() time.Time {
	lifetime := envy.Get("ACCESS_TOKEN_LIFETIME", "28800")

	lifetimeSeconds, err := strconv.Atoi(lifetime)
	if err != nil {
		lifetimeSeconds = 28800
	}

	dtNow := time.Now()
	futureTime := dtNow.Add(time.Second * time.Duration(lifetimeSeconds))

	return futureTime
}

func createAccessTokenPart() string {
	var alphanumerics = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	tokenLength := 32
	b := make([]rune, tokenLength)
	for i := range b {
		b[i] = alphanumerics[rand.Intn(len(alphanumerics))]
	}

	accessToken := string(b)

	return accessToken
}

func hashClientIdAccessToken(accessToken string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(accessToken)))
}
