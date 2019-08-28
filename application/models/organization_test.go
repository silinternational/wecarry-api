package models

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"

	"testing"
)

func (ms *ModelSuite) TestFindOrgByUUID() {
	t := ms.T()
	org, _ := createOrgFixtures(t)

	type args struct {
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		want    Organization
		wantErr bool
	}{
		{
			name: "found",
			args: args{org.Uuid.String()},
			want: org,
		},
		{
			name:    "empty access token",
			args:    args{""},
			wantErr: true,
		},
		{
			name:    "near match",
			args:    args{"51b5321d-2769-48a0-908a-7af1d15083e3"},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FindOrgByUUID(test.args.uuid)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindOrgByUUID() returned an error: %v", err)
				} else if got.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", got, test.want)
				}
			}
		})
	}

	if err := DB.Destroy(&org); err != nil {
		t.Errorf("error deleting test data: %v", err)
	}
}

func (ms *ModelSuite) TestCreateOrganization() {
	t := ms.T()
	tests := []struct {
		name    string
		org     Organization
		wantErr bool
	}{
		{
			name: "full",
			org: Organization{
				Name:       "ACME",
				Uuid:       domain.GetUuid(),
				AuthType:   "saml2",
				AuthConfig: "{}",
				Url:        nulls.NewString("https://www.example.com"),
			},
			wantErr: false,
		},
		{
			name: "minimum",
			org: Organization{
				Name:       "Bits 'R' Us",
				Uuid:       domain.GetUuid(),
				AuthType:   "saml2",
				AuthConfig: "{}",
			},
			wantErr: false,
		},
		{
			name: "missing auth config",
			org: Organization{
				Name:     "Bits 'R' Us",
				Uuid:     domain.GetUuid(),
				AuthType: "saml2",
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := DB.Create(&test.org)
			if test.wantErr == true {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else if err != nil {
				t.Errorf("Unexpected error %v", err)
			} else {
				org, err := FindOrgByUUID(test.org.Uuid.String())
				if err != nil {
					t.Errorf("Couldn't find new org %v: %v", test.org.Name, err)
				}
				if org.Uuid != test.org.Uuid {
					t.Errorf("newly created org doesn't match, found %v, expected %v", org, test.org)
				}
			}

			// clean up
			if err := DB.Destroy(&test.org); err != nil {
				t.Errorf("error deleting test data: %v", err)
			}
		})
	}
}

func (ms *ModelSuite) TestValidateOrganization() {
	t := ms.T()
	tests := []struct {
		name     string
		org      Organization
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			org: Organization{
				Name:       "Bits 'R' Us",
				Uuid:       domain.GetUuid(),
				AuthType:   "saml2",
				AuthConfig: "[]",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			org: Organization{
				Uuid:       domain.GetUuid(),
				AuthType:   "saml2",
				AuthConfig: "[]",
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "missing uuid",
			org: Organization{
				Name:       "Babelfish Warehouse",
				AuthType:   "saml2",
				AuthConfig: "[]",
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing auth type",
			org: Organization{
				Name:       "Babelfish Warehouse",
				Uuid:       domain.GetUuid(),
				AuthConfig: "[]",
			},
			wantErr:  true,
			errField: "auth_type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.org.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestOrganizationFindByDomain() {
	t := ms.T()
	org, orgDomain := createOrgFixtures(t)

	type args struct {
		domain string
	}
	tests := []struct {
		name    string
		args    args
		want    Organization
		wantErr bool
	}{
		{
			name: "found",
			args: args{orgDomain.Domain},
			want: org,
		},
		{
			name:    "empty string",
			args:    args{""},
			wantErr: true,
		},
		{
			name:    "near match",
			args:    args{"example.com"},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := OrganizationFindByDomain(test.args.domain)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("OrganizationFindByDomain() returned an error: %v", err)
				} else if got.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", got, test.want)
				}
			}
		})
	}

	// delete org fixture and org domain by cascading delete
	if err := DB.Destroy(&org); err != nil {
		t.Errorf("error deleting test data: %v", err)
	}
}

func createOrgFixtures(t *testing.T) (Organization, OrganizationDomain) {
	// Load Organization test fixtures
	org := Organization{
		Name:       "ACME",
		Uuid:       domain.GetUuid(),
		AuthType:   "saml2",
		AuthConfig: "[]",
	}
	if err := DB.Create(&org); err != nil {
		t.Errorf("could not create org fixtures ... %v", err)
		t.FailNow()
	}

	// Load Organization Domains test fixtures
	orgDomain := OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "example.org",
	}
	if err := DB.Create(&orgDomain); err != nil {
		t.Errorf("could not create org domain fixtures ... %v", err)
		t.FailNow()
	}

	return org, orgDomain
}
