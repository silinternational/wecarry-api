package models

import (
	"time"

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
			var org Organization
			err := org.FindByUUID(test.args.uuid)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindByUUID() returned an error: %v", err)
				} else if org.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", org, test.want)
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
				var org Organization
				err := org.FindByUUID(test.org.Uuid.String())
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
			var org Organization
			err := org.FindByDomain(test.args.domain)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("OrganizationFindByDomain() returned an error: %v", err)
				} else if org.Uuid != test.want.Uuid {
					t.Errorf("found %v, expected %v", org, test.want)
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

func TestOrganization_AddLoadRemoveDomain(t *testing.T) {
	resetTables(t)

	orgFixtures := []Organization{
		{
			ID:         1,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   "na",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
		{
			ID:         2,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   "na",
			AuthConfig: "{}",
			Uuid:       domain.GetUuid(),
		},
	}
	for _, org := range orgFixtures {
		err := DB.Create(&org)
		if err != nil {
			t.Errorf("Unable to create org fixture: %s", err)
			t.FailNow()
		}
	}

	first, err := orgFixtures[0].AddDomain("first.com")
	if err != nil {
		t.Errorf("unable to add first domain to Org1: %s", err)
	} else if first.ID == 0 {
		t.Errorf("did not get error, but failed to add first domain to Org1")
	}

	second, err := orgFixtures[0].AddDomain("second.com")
	if err != nil {
		t.Errorf("unable to add second domain to Org1: %s", err)
	} else if second.ID == 0 {
		t.Errorf("did not get error, but failed to add second domain to Org1")
	}

	_, err = orgFixtures[1].AddDomain("second.com")
	if err == nil {
		t.Errorf("was to add existing domain (second.com) to Org2 but should have gotten error")
	}

	if len(orgFixtures[0].OrganizationDomains) != 0 {
		t.Errorf("hmm, didn't expect that")
	}

	err = orgFixtures[0].loadDomains()
	if err != nil {
		t.Errorf("unable to reload domains: %s", err)
	}

	if len(orgFixtures[0].OrganizationDomains) != 2 {
		t.Errorf("after reloading org domains we did not get what we expected (%v), got: %v", 2, len(orgFixtures[0].OrganizationDomains))
	}

	err = orgFixtures[0].RemoveDomain("first.com")
	if err != nil {
		t.Errorf("unable to remove domain: %s", err)
	}

	if len(orgFixtures[0].OrganizationDomains) != 1 {
		t.Errorf("org domains count after removing domain is not correct, expected %v, got: %v", 1, len(orgFixtures[0].OrganizationDomains))
	}

}
