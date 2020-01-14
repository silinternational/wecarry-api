package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/silinternational/wecarry-api/auth"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/auth/saml"
	"github.com/silinternational/wecarry-api/domain"
)

const AuthTypeSaml = "saml"
const AuthTypeGoogle = "google"

type Organization struct {
	ID                  int                  `json:"id" db:"id"`
	CreatedAt           time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at" db:"updated_at"`
	Name                string               `json:"name" db:"name"`
	Url                 nulls.String         `json:"url" db:"url"`
	AuthType            string               `json:"auth_type" db:"auth_type"`
	AuthConfig          string               `json:"auth_config" db:"auth_config"`
	UUID                uuid.UUID            `json:"uuid" db:"uuid"`
	LogoFileID          nulls.Int            `json:"logo_file_id" db:"logo_file_id"`
	Users               Users                `many_to_many:"user_organizations" order_by:"nickname"`
	OrganizationDomains []OrganizationDomain `has_many:"organization_domains" order_by:"domain asc"`
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
		&validators.StringIsPresent{Field: o.AuthType, Name: "AuthType"},
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

func (o *Organization) GetAuthProvider() (auth.Provider, error) {

	if o.AuthType == AuthTypeSaml {
		return saml.New([]byte(o.AuthConfig))
	}

	if o.AuthType == AuthTypeGoogle {
		return google.New([]byte(o.AuthConfig))
	}

	return &auth.EmptyProvider{}, fmt.Errorf("unsupported auth provider type: %s", o.AuthType)
}

func (o *Organization) FindByUUID(uuid string) error {

	if uuid == "" {
		return errors.New("error: org uuid must not be blank")
	}

	if err := DB.Where("uuid = ?", uuid).First(o); err != nil {
		return fmt.Errorf("error finding org by uuid: %s", err.Error())
	}

	return nil
}

func (o *Organization) FindByDomain(domain string) error {
	var orgDomain OrganizationDomain
	if err := DB.Where("domain = ?", domain).First(&orgDomain); err != nil {
		return fmt.Errorf("error finding organization_domain by domain: %s", err.Error())
	}

	if err := DB.Eager().Where("id = ?", orgDomain.OrganizationID).First(o); err != nil {
		return fmt.Errorf("error finding organization by domain: %s", err.Error())
	}

	return nil
}

func (o *Organization) AddDomain(domain string) error {
	// make sure domain is not registered to another org first
	var orgDomain OrganizationDomain

	count, err := DB.Where("domain = ?", domain).Count(&orgDomain)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("this domain (%s) is already in use", domain)
	}

	orgDomain.Domain = domain
	orgDomain.OrganizationID = o.ID
	return orgDomain.Create()
}

func (o *Organization) RemoveDomain(domain string) error {
	var orgDomain OrganizationDomain
	if err := DB.Where("organization_id = ? and domain = ?", o.ID, domain).First(&orgDomain); err != nil {
		return err
	}

	return DB.Destroy(&orgDomain)
}

// Save wrap DB.Save() call to check for errors and operate on attached object
func (o *Organization) Save() error {
	return save(o)
}

func (orgs *Organizations) All() error {
	return DB.All(orgs)
}

func (orgs *Organizations) AllWhereUserIsOrgAdmin(cUser User) error {
	if cUser.AdminRole == UserAdminRoleSuperAdmin || cUser.AdminRole == UserAdminRoleSalesAdmin {
		return orgs.All()
	}

	return DB.
		Scope(scopeUserAdminOrgs(cUser)).
		Order("name asc").
		All(orgs)
}

// GetDomains finds and returns all related OrganizationDomain rows.
func (o *Organization) GetDomains() ([]OrganizationDomain, error) {
	if err := DB.Load(o, "OrganizationDomains"); err != nil {
		return nil, err
	}

	return o.OrganizationDomains, nil
}

// GetUsers finds and returns all related Users.
func (o *Organization) GetUsers() (Users, error) {
	if o.ID <= 0 {
		return nil, errors.New("invalid Organization ID")
	}

	if err := DB.Load(o, "Users"); err != nil {
		return nil, err
	}

	return o.Users, nil
}

// scope query to only include organizations that the cUser is an admin of
func scopeUserAdminOrgs(cUser User) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		var adminOrgIDs []int

		_ = DB.Load(&cUser, "UserOrganizations")

		for _, uo := range cUser.UserOrganizations {
			if uo.Role == UserOrganizationRoleAdmin {
				adminOrgIDs = append(adminOrgIDs, uo.OrganizationID)
			}
		}

		// convert []int to []interface{}
		s := make([]interface{}, len(adminOrgIDs))
		for i, v := range adminOrgIDs {
			s[i] = v
		}

		if len(s) == 0 {
			return q.Where("id = -1")
		}
		return q.Where("id IN (?)", s...)
	}
}

// LogoURL retrieves the logo URL from the attached file
func (o *Organization) LogoURL() (*string, error) {
	if o.LogoFileID.Valid {
		var file File
		if err := DB.Find(&file, o.LogoFileID); err != nil {
			return nil, fmt.Errorf("couldn't find org file %d, %s", o.LogoFileID.Int, err)
		}
		if err := file.refreshURL(); err != nil {
			return nil, fmt.Errorf("error getting logo URL, %s", err)
		}
		return &file.URL, nil
	}
	return nil, nil
}

// CreateTrust creates a Trust record linking this Organization with the organization identified by `secondaryID`
func (o *Organization) CreateTrust(secondaryID string) error {
	return nil
}

// RemoveTrust removes a Trust record between this Organization and the organization identified by `secondaryID`
func (o *Organization) RemoveTrust(secondaryID string) error {
	return nil
}

// TrustedOrganizations gets a list of connected Organizations, either primary or secondary
func (o *Organization) TrustedOrganizations() (Organizations, error) {
	t := Trusts{}
	if err := t.FindByOrgID(o.ID); domain.IsOtherThanNoRows(err) {
		return nil, err
	}
	ids := make([]interface{}, len(t))
	for i := range t {
		if o.ID == t[i].PrimaryID {
			ids[i] = t[i].SecondaryID
		} else {
			ids[i] = t[i].PrimaryID
		}
	}
	trustedOrgs := Organizations{}
	if err := DB.Where("id in (?)", ids...).All(&trustedOrgs); err != nil {
		return nil, err
	}
	return trustedOrgs, nil
}
