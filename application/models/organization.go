package models

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/auth/azureadv2"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/auth/saml"
	"github.com/silinternational/wecarry-api/domain"
)

type AuthType string

const (
	// AuthTypeAzureAD : Microsoft Azure AD (Office 365)
	AuthTypeAzureAD AuthType = "AZUREADV2"
	// AuthTypeDefault : Default to Organization's AuthType (only valid on OrganizationDomain)
	AuthTypeDefault AuthType = "DEFAULT"
	// AuthTypeGoogle : Google OAUTH 2.0
	AuthTypeGoogle AuthType = "GOOGLE"
	// AuthTypeSaml : SAML 2.0
	AuthTypeSaml AuthType = "SAML"
)

func (e AuthType) IsValid() bool {
	switch e {
	case AuthTypeAzureAD, AuthTypeDefault, AuthTypeGoogle, AuthTypeSaml:
		return true
	}
	return false
}

func (e AuthType) String() string {
	return string(e)
}

func (e *AuthType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AuthType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AuthType", str)
	}
	return nil
}

func (e AuthType) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}

// Organization subscribed to the App. Provides privacy controls for visibility of Requests and Meetings, and specifies
// authentication for associated users.
// swagger:model
type Organization struct {
	// ----- Database-only fields
	ID         int          `json:"-" db:"id"`
	CreatedAt  time.Time    `json:"-" db:"created_at"`
	UpdatedAt  time.Time    `json:"-" db:"updated_at"`
	Url        nulls.String `json:"-" db:"url"`
	AuthType   AuthType     `json:"-" db:"auth_type"`
	AuthConfig string       `json:"-" db:"auth_config"`
	FileID     nulls.Int    `json:"-" db:"file_id"`
	Users      Users        `json:"-" many_to_many:"user_organizations" order_by:"nickname"`

	// ----- Database & API fields

	// unique identifier for the Organization
	// swagger:strfmt uuid4
	// example: 63d5b060-1460-4348-bdf0-ad03c105a8d5
	UUID uuid.UUID `json:"id" db:"uuid"`

	// Organization name, limited to 255 characters
	Name string `json:"name" db:"name"`
}

// String is used to serialize error extras
func (o Organization) String() string {
	ju, _ := json.Marshal(o)
	return string(ju)
}

type Organizations []Organization

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (o *Organization) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: o.Name, Name: "Name"},
		&validators.StringIsPresent{Field: o.AuthType.String(), Name: "AuthType"},
		&validators.UUIDIsPresent{Field: o.UUID, Name: "UUID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (o *Organization) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (o *Organization) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// GetAuthProvider returns the auth provider associated with the domain of `authEmail`, if assigned, otherwise from the Organization's auth provider.
func (o *Organization) GetAuthProvider(tx *pop.Connection, authEmail string) (auth.Provider, error) {
	// Use type and config from organization by default
	authType := AuthType(strings.ToUpper(o.AuthType.String()))
	authConfig := o.AuthConfig

	// Check if organization domain has override auth config to use instead of default
	authDomain := domain.EmailDomain(authEmail)
	var orgDomain OrganizationDomain
	if err := orgDomain.FindByDomain(tx, authDomain); err != nil {
		return &auth.EmptyProvider{}, err
	}
	if orgDomain.AuthType != "" && orgDomain.AuthType != AuthTypeDefault {
		authType = orgDomain.AuthType
		authConfig = orgDomain.AuthConfig
	}

	switch authType {
	case AuthTypeAzureAD:
		return azureadv2.New([]byte(authConfig))
	case AuthTypeGoogle:
		return google.New(
			struct{ Key, Secret string }{
				Key:    domain.Env.GoogleKey,
				Secret: domain.Env.GoogleSecret,
			},
			[]byte(authConfig),
		)
	case AuthTypeSaml:
		return saml.New([]byte(authConfig))

	}

	return &auth.EmptyProvider{}, fmt.Errorf("unsupported auth provider type: %s", o.AuthType)
}

func (o *Organization) FindByUUID(tx *pop.Connection, uuid string) error {
	if uuid == "" {
		return errors.New("error: org uuid must not be blank")
	}

	if err := tx.Where("uuid = ?", uuid).First(o); err != nil {
		return fmt.Errorf("error finding org by uuid: %s", err.Error())
	}

	return nil
}

func (o *Organization) FindByDomain(tx *pop.Connection, domain string) error {
	var orgDomain OrganizationDomain
	if err := tx.Where("domain = ?", domain).First(&orgDomain); err != nil {
		return fmt.Errorf("error finding organization_domain by domain: %s", err.Error())
	}

	if err := tx.Eager().Where("id = ?", orgDomain.OrganizationID).First(o); err != nil {
		return fmt.Errorf("error finding organization by domain: %s", err.Error())
	}

	return nil
}

func (o *Organization) AddDomain(tx *pop.Connection, domainName string, authType AuthType, authConfig string) error {
	// make sure domainName is not already in use
	var orgDomain OrganizationDomain
	if err := orgDomain.FindByDomain(tx, domainName); domain.IsOtherThanNoRows(err) {
		return err
	}
	if orgDomain.ID != 0 {
		return fmt.Errorf("this domainName (%s) is already in use", domainName)
	}

	orgDomain.Domain = domainName
	orgDomain.OrganizationID = o.ID
	orgDomain.AuthType = authType
	orgDomain.AuthConfig = authConfig
	return orgDomain.Create(tx)
}

func (o *Organization) RemoveDomain(tx *pop.Connection, domain string) error {
	var orgDomain OrganizationDomain
	if err := tx.Where("organization_id = ? and domain = ?", o.ID, domain).First(&orgDomain); err != nil {
		return err
	}

	return tx.Destroy(&orgDomain)
}

// Save wrap tx.Save() call to check for errors and operate on attached object
func (o *Organization) Save(tx *pop.Connection) error {
	return save(tx, o)
}

func (o *Organizations) All(tx *pop.Connection) error {
	return tx.All(o)
}

func (o *Organizations) AllWhereUserIsOrgAdmin(tx *pop.Connection, cUser User) error {
	if cUser.AdminRole == UserAdminRoleSuperAdmin || cUser.AdminRole == UserAdminRoleSalesAdmin {
		return o.All(tx)
	}

	return tx.
		Scope(scopeUserAdminOrgs(tx, cUser)).
		Order("name asc").
		All(o)
}

// Domains finds and returns all related OrganizationDomain rows.
func (o *Organization) Domains(tx *pop.Connection) ([]OrganizationDomain, error) {
	var domains OrganizationDomains
	if err := tx.Where("organization_id=?", o.ID).Order("domain asc").All(&domains); err != nil {
		return nil, err
	}

	return domains, nil
}

// GetUsers finds and returns all related Users.
func (o *Organization) GetUsers(tx *pop.Connection) (Users, error) {
	if o.ID <= 0 {
		return nil, errors.New("invalid Organization ID")
	}

	if err := tx.Load(o, "Users"); err != nil {
		return nil, err
	}

	return o.Users, nil
}

// scope query to only include organizations that the cUser is an admin of
func scopeUserAdminOrgs(tx *pop.Connection, cUser User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		var adminOrgIDs []int

		_ = tx.Load(&cUser, "UserOrganizations")

		for _, uo := range cUser.UserOrganizations {
			if uo.Role == UserOrganizationRoleAdmin {
				adminOrgIDs = append(adminOrgIDs, uo.OrganizationID)
			}
		}

		s := convertSliceFromIntToInterface(adminOrgIDs)

		if len(s) == 0 {
			return q.Where("id = -1")
		}
		return q.Where("id IN (?)", s...)
	}
}

// LogoURL retrieves the logo URL from the attached file
func (o *Organization) LogoURL(tx *pop.Connection) (*string, error) {
	if o.FileID.Valid {
		var file File
		tx := tx
		if err := tx.Find(&file, o.FileID); err != nil {
			return nil, fmt.Errorf("couldn't find org file %d, %s", o.FileID.Int, err)
		}
		if err := file.RefreshURL(tx); err != nil {
			return nil, fmt.Errorf("error getting logo URL, %s", err)
		}
		return &file.URL, nil
	}
	return nil, nil
}

// CreateTrust creates a OrganizationTrust record linking this Organization with the organization identified by `secondaryID`
func (o *Organization) CreateTrust(tx *pop.Connection, secondaryID string) error {
	var secondaryOrg Organization
	if err := secondaryOrg.FindByUUID(tx, secondaryID); err != nil {
		return fmt.Errorf("CreateTrust, error finding secondary org, %s", err)
	}
	var t OrganizationTrust
	t.PrimaryID = o.ID
	t.SecondaryID = secondaryOrg.ID
	if err := t.CreateSymmetric(tx); err != nil {
		return fmt.Errorf("failed to create new OrganizationTrust, %s", err)
	}
	return nil
}

// RemoveTrust removes a OrganizationTrust record between this Organization and the organization identified by `secondaryID`
func (o *Organization) RemoveTrust(tx *pop.Connection, secondaryID string) error {
	var secondaryOrg Organization
	if err := secondaryOrg.FindByUUID(tx, secondaryID); err != nil {
		return fmt.Errorf("RemoveTrust, error finding secondary org, %s", err)
	}
	var t OrganizationTrust
	return t.RemoveSymmetric(tx, o.ID, secondaryOrg.ID)
}

// TrustedOrganizations gets a list of connected Organizations
func (o *Organization) TrustedOrganizations(tx *pop.Connection) (Organizations, error) {
	t := OrganizationTrusts{}
	if err := t.FindByOrgID(tx, o.ID); domain.IsOtherThanNoRows(err) {
		return nil, err
	}
	if len(t) < 1 {
		return Organizations{}, nil
	}
	ids := make([]interface{}, len(t))
	for i := range t {
		ids[i] = t[i].SecondaryID
	}
	trustedOrgs := Organizations{}
	if err := tx.Where("id in (?)", ids...).All(&trustedOrgs); err != nil {
		return nil, err
	}
	return trustedOrgs, nil
}

// AttachLogo assigns a previously-stored File to this Organization as its logo. Parameter `fileID` is the UUID
// of the file to attach.
func (o *Organization) AttachLogo(tx *pop.Connection, fileID string) (File, error) {
	return addFile(tx, o, fileID)
}

// RemoveFile removes an attached file from the Request
func (o *Organization) RemoveFile(tx *pop.Connection) error {
	return removeFile(tx, o)
}

// FindByIDs finds all Organizations associated with the given IDs and loads them from the database
func (o *Organizations) FindByIDs(tx *pop.Connection, ids []int) error {
	ids = domain.UniquifyIntSlice(ids)
	return tx.Where("id in (?)", ids).All(o)
}
