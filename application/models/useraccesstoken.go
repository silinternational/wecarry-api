package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/pkg/errors"
	"github.com/silinternational/wecarry-api/domain"
)

type UserAccessToken struct {
	ID                 int              `json:"id" db:"id"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
	UserID             int              `json:"user_id" db:"user_id"`
	UserOrganizationID nulls.Int        `json:"user_organization_id" db:"user_organization_id"`
	AccessToken        string           `json:"access_token" db:"access_token"`
	ExpiresAt          time.Time        `json:"expires_at" db:"expires_at"`
	User               User             `belongs_to:"users"`
	UserOrganization   UserOrganization `belongs_to:"user_organizations"`
}

// String can be helpful for serializing the model
func (u UserAccessToken) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserAccessTokens is merely for convenience and brevity
type UserAccessTokens []UserAccessToken

// String can be helpful for serializing the model
func (u UserAccessTokens) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (u *UserAccessToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: u.AccessToken, Name: "AccessToken"},
		&validators.TimeIsPresent{Field: u.ExpiresAt, Name: "ExpiresAt"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (u *UserAccessToken) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (u *UserAccessToken) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (u *UserAccessToken) DeleteByBearerToken(bearerToken string) error {
	if err := u.FindByBearerToken(bearerToken); err != nil {
		return err
	}
	return DB.Destroy(u)
}

func (u *UserAccessToken) FindByBearerToken(bearerToken string) error {
	if err := DB.Eager().Where("access_token = ?", HashClientIdAccessToken(bearerToken)).First(u); err != nil {
		l := len(bearerToken)
		if l > 5 {
			l = 5
		}
		return fmt.Errorf("failed to find access token '%s...', %s", bearerToken[0:l], err)
	}

	return nil
}

// GetOrganization returns the Organization of the UserOrganization of the UserAccessToken
func (u *UserAccessToken) GetOrganization() (Organization, error) {
	if u.UserOrganization.ID <= 0 {
		if err := DB.Load(u, "UserOrganization"); err != nil {
			return Organization{}, fmt.Errorf("error loading user organization for user access token id %v ... %v",
				u.ID, err)
		}
	}

	uOrg := u.UserOrganization

	if uOrg.OrganizationID <= 0 {
		return Organization{}, fmt.Errorf("user access token id %v has no organization", u.ID)
	}

	if uOrg.Organization.ID <= 0 {
		if err := DB.Load(&uOrg, "Organization"); err != nil {
			return Organization{}, fmt.Errorf("error loading user organization for user access token id %v ... %v",
				u.ID, err)
		}
	}

	return uOrg.Organization, nil
}

func createAccessTokenExpiry() time.Time {
	dtNow := time.Now()
	futureTime := dtNow.Add(time.Second * time.Duration(domain.Env.AccessTokenLifetimeSeconds))

	return futureTime
}

// Renew extends the token expiration to the configured token lifetime
func (u *UserAccessToken) Renew() error {
	u.ExpiresAt = createAccessTokenExpiry()
	if err := u.Update(); err != nil {
		return fmt.Errorf("error renewing access token, %s", err)
	}
	return nil
}

// GetUser returns the User associated with this access token
func (u *UserAccessToken) GetUser() (User, error) {
	if err := DB.Load(u, "User"); err != nil {
		return User{}, err
	}
	if u.User.ID <= 0 {
		return User{}, errors.New("no user associated with access token")
	}
	return u.User, nil
}

// DeleteIfExpired checks the token expiration and returns `true` if expired. Also deletes
// the token from the database if it is expired.
func (u *UserAccessToken) DeleteIfExpired() (bool, error) {
	if u.ExpiresAt.Before(time.Now()) {
		err := DB.Destroy(u)
		if err != nil {
			return true, fmt.Errorf("unable to delete expired userAccessToken, id: %v", u.ID)
		}
		return true, nil
	}
	return false, nil
}

// DeleteExpired removes all expired UserAccessToken records
func (u *UserAccessTokens) DeleteExpired() (int, error) {
	var c Count
	err := DB.RawQuery("SELECT COUNT(*) FROM user_access_tokens WHERE expires_at < ?", time.Now()).First(&c)
	if err != nil || c.N == 0 {
		return 0, err
	}

	err = DB.RawQuery("DELETE FROM user_access_tokens WHERE expires_at < ?", time.Now()).Exec()
	if err != nil {
		return 0, err
	}

	return c.N, nil
}

// Create stores the UserAccessToken data as a new record in the database.
func (u *UserAccessToken) Create() error {
	return create(u)
}

// Update writes the UserAccessToken data to an existing database record.
func (u *UserAccessToken) Update() error {
	return update(u)
}
