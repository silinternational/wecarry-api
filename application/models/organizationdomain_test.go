package models

import (
	"testing"
)

func (ms *ModelSuite) TestOrganizationDomain_Organization() {
	t := ms.T()

	org := createOrganizationFixtures(ms.DB, 1)[0]

	orgDomain := OrganizationDomain{OrganizationID: org.ID, Domain: "example.com"}
	createFixture(ms, &orgDomain)

	tests := []struct {
		name      string
		orgDomain OrganizationDomain
		want      Organization
		wantErr   bool
	}{
		{
			name:      "valid",
			orgDomain: orgDomain,
			want:      org,
		},
		{
			name:      "error",
			orgDomain: OrganizationDomain{},
			wantErr:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := test.orgDomain
			got, err := o.Organization()

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			ms.Equal(test.want.ID, got.ID)
			ms.Equal(test.want.UUID, got.UUID)
			ms.Equal(test.want.Name, got.Name)
			ms.Equal(test.want.Url, got.Url)
			ms.Equal(test.want.FileID, got.FileID)
			ms.Equal(test.want.AuthType, got.AuthType)
			ms.Equal(test.want.AuthConfig, got.AuthConfig)
		})
	}
}
