package models

import (
	"github.com/gofrs/uuid"
	"reflect"
	"testing"
)

func TestFindOrgByUUID(t *testing.T) {
	if err := BounceTestDB(); err != nil {
		t.Errorf("BounceTestDB() failed: %v", err)
		return
	}

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
}

func Test(t *testing.T) {
}
