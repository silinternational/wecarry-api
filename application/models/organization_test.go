package models

import (
	"fmt"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
	"github.com/silinternational/handcarry-api/domain"

	"reflect"
	"testing"
)

func init() {
	if err := BounceTestDB(); err != nil {
		panic(fmt.Errorf("BounceTestDB() failed: %v", err))
	}
}

func TestFindOrgByUUID(t *testing.T) {
	// Load Organization test fixtures
	orgUuidStr := "51b5321d-2769-48a0-908a-7af1d15083e2"
	orgUuid1, _ := uuid.FromString(orgUuidStr)
	orgFix := Organizations{
		{
			ID:         1,
			Name:       "ACME",
			Uuid:       orgUuid1,
			AuthType:   "saml2",
			AuthConfig: "[]",
		},
	}
	if err := CreateOrgs(orgFix); err != nil {
		t.Errorf("could not run test ... %v", err)
		return
	}

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
			args: args{orgUuidStr},
			want: orgFix[0],
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
				} else if reflect.DeepEqual(got, test.want) {
					t.Errorf("found %v, expected %v", got, test.want)
				}
			}
		})
	}

	// clean up
	resetOrganizationsTable(t)
}

func TestCreateOrganization(t *testing.T) {
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
				AuthConfig: "[]",
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
				AuthConfig: "[]",
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
			if (test.wantErr == true) && (err == nil) {
				t.Errorf("Expected an error, but did not get one")
			} else if (test.wantErr == false) && (err != nil) {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}

	// clean up
	resetOrganizationsTable(t)
}

func TestValidateOrganization(t *testing.T) {
	tests := []struct {
		name     string
		org      Organization
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			org: Organization{
				ID:         1,
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
				ID:         1,
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
				ID:         1,
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
				ID:         1,
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
