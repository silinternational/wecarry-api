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
	"github.com/silinternational/wecarry-api/auth/azureadv2"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/auth/saml"
	"github.com/silinternational/wecarry-api/domain"
)

const AuthTypeAzureAD = "azureadv2"
const AuthTypeFacebook = "facebook"
const AuthTypeGoogle = "google"
const AuthTypeLinkedIn = "linkedin"
const AuthTypeSaml = "saml"
const AuthTypeTwitter = "twitter"

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

// GetAuthProvider returns the auth provider associated with the domain of `authEmail`, if assigned, otherwise from the Organization's auth provider.
func (o *Organization) GetAuthProvider(authEmail string) (auth.Provider, error) {
	// Use type and config from organization by default
	authType := o.AuthType
	authConfig := o.AuthConfig

	// Check if organization domain has override auth config to use instead of default
	authDomain := domain.EmailDomain(authEmail)
	var orgDomain OrganizationDomain
	if err := orgDomain.FindByDomain(authDomain); err != nil {
		return &auth.EmptyProvider{}, err
	}
	if orgDomain.AuthType != "" {
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

func (o *Organization) AddDomain(domainName, authType, authConfig string) error {
	// make sure domainName is not already in use
	var orgDomain OrganizationDomain
	if err := orgDomain.FindByDomain(domainName); domain.IsOtherThanNoRows(err) {
		return err
	}
	if orgDomain.ID != 0 {
		return fmt.Errorf("this domainName (%s) is already in use", domainName)
	}

	orgDomain.Domain = domainName
	orgDomain.OrganizationID = o.ID
	orgDomain.AuthType = authType
	orgDomain.AuthConfig = authConfig
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

		s := convertSliceFromIntToInterface(adminOrgIDs)

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

// CreateTrust creates a OrganizationTrust record linking this Organization with the organization identified by `secondaryID`
func (o *Organization) CreateTrust(secondaryID string) error {
	var secondaryOrg Organization
	if err := secondaryOrg.FindByUUID(secondaryID); err != nil {
		return fmt.Errorf("CreateTrust, error finding secondary org, %s", err)
	}
	var t OrganizationTrust
	t.PrimaryID = o.ID
	t.SecondaryID = secondaryOrg.ID
	if err := t.CreateSymmetric(); err != nil {
		return fmt.Errorf("failed to create new OrganizationTrust, %s", err)
	}
	return nil
}

// RemoveTrust removes a OrganizationTrust record between this Organization and the organization identified by `secondaryID`
func (o *Organization) RemoveTrust(secondaryID string) error {
	var secondaryOrg Organization
	if err := secondaryOrg.FindByUUID(secondaryID); err != nil {
		return fmt.Errorf("RemoveTrust, error finding secondary org, %s", err)
	}
	var t OrganizationTrust
	return t.RemoveSymmetric(o.ID, secondaryOrg.ID)
}

// TrustedOrganizations gets a list of connected Organizations
func (o *Organization) TrustedOrganizations() (Organizations, error) {
	t := OrganizationTrusts{}
	if err := t.FindByOrgID(o.ID); domain.IsOtherThanNoRows(err) {
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
	if err := DB.Where("id in (?)", ids...).All(&trustedOrgs); err != nil {
		return nil, err
	}
	return trustedOrgs, nil
}

// AttachLogo assigns a previously-stored File to this Organization as its logo. Parameter `fileID` is the UUID
// of the file to attach.
func (o *Organization) AttachLogo(fileID string) (File, error) {
	var f File
	if err := f.FindByUUID(fileID); err != nil {
		err = fmt.Errorf("error finding organization logo with id %s ... %s", fileID, err)
		return f, err
	}

	oldID := o.LogoFileID
	o.LogoFileID = nulls.NewInt(f.ID)
	if o.ID > 0 {
		if err := DB.UpdateColumns(o, "logo_file_id"); err != nil {
			return f, err
		}
	}

	if err := f.SetLinked(); err != nil {
		domain.ErrLogger.Printf("error marking org logo file %d as linked, %s", f.ID, err)
	}

	if oldID.Valid {
		oldFile := File{ID: oldID.Int}
		if err := oldFile.ClearLinked(); err != nil {
			domain.ErrLogger.Printf("error marking old org logo file %d as unlinked, %s", oldFile.ID, err)
		}
	}

	return f, nil
}
