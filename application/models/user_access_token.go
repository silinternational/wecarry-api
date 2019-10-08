package models

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/envy"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type UserAccessToken struct {
	ID                 int              `json:"id" db:"id"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
	UserID             int              `json:"user_id" db:"user_id"`
	UserOrganizationID int              `json:"user_organization_id" db:"user_organization_id"`
	AccessToken        string           `json:"access_token" db:"access_token"`
	ExpiresAt          time.Time        `json:"expires_at" db:"expires_at"`
	User               User             `belongs_to:"users"`
	UserOrganization   UserOrganization `belongs_to:"user_organizations"`
}

// String is not required by pop and may be deleted
func (u UserAccessToken) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserAccessTokens is not required by pop and may be deleted
type UserAccessTokens []UserAccessToken

// String is not required by pop and may be deleted
func (u UserAccessTokens) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *UserAccessToken) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: u.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: u.AccessToken, Name: "AccessToken"},
		&validators.TimeIsPresent{Field: u.ExpiresAt, Name: "ExpiresAt"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *UserAccessToken) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
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
	envLifetime := envy.Get("ACCESS_TOKEN_LIFETIME", strconv.Itoa(domain.AccessTokenLifetimeSeconds))

	lifetimeSeconds, err := strconv.Atoi(envLifetime)
	if err != nil {
        // TODO Ensure this gets logged so that we know our env var is bad
		lifetimeSeconds = domain.AccessTokenLifetimeSeconds
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

// Renew extends the token expiration to the configured token lifetime
func (u *UserAccessToken) Renew() error {
	u.ExpiresAt = createAccessTokenExpiry()
	if err := DB.Update(u); err != nil {
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
		return User{}, fmt.Errorf("no user associated with access token")
	}
	return u.User, nil
}

// IsExpired checks the token expiration and returns `true` if expired. Also deletes the token from the database
// if it is expired.
func (u *UserAccessToken) IsExpired() bool {
	if u.ExpiresAt.Before(time.Now()) {
		err := DB.Destroy(u)
		if err != nil {
			log.Printf("Unable to delete expired userAccessToken, id: %v", u.ID)
		}
		return true
	}
	return false
}
