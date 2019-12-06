package models

import (
	"testing"

	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestOrganizationDomain_GetOrganizationUUID() {
	t := ms.T()

	org := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &org)

	orgDomain := OrganizationDomain{OrganizationID: org.ID, Domain: "example.com"}
	createFixture(ms, &orgDomain)

	tests := []struct {
		name      string
		orgDomain OrganizationDomain
		want      string
		wantErr   bool
	}{
		{
			name:      "valid",
			orgDomain: orgDomain,
			want:      org.UUID.String(),
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
			got, err := o.GetOrganizationUUID()

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, got)
		})
	}
}
