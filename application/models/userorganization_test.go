package models

import (
	"testing"

	"github.com/silinternational/wecarry-api/domain"
)

func createUserOrganizationFixtures(ms *ModelSuite, t *testing.T) (Users, Organizations) {
	orgs := Organizations{{}, {}}
	for i := range orgs {
		orgs[i].UUID = domain.GetUUID()
		orgs[i].AuthConfig = "{}"
		createFixture(ms, &orgs[i])
	}

	uf := CreateUserFixtures(ms.DB, 2)
	users := uf.Users

	// both users are in org 0, but need user 0 to also be in org 1
	createFixture(ms, &UserOrganization{
		OrganizationID: orgs[1].ID,
		UserID:         users[0].ID,
		AuthID:         users[0].Email,
		AuthEmail:      users[0].Email,
	})

	return uf.Users, orgs
}

func (ms *ModelSuite) TestUserOrganization_FindByAuthEmail() {
	t := ms.T()
	users, _ := createUserOrganizationFixtures(ms, t)

	type args struct {
		authEmail string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "invalid email error",
			args:    args{authEmail: "not an email"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "no results",
			args:    args{authEmail: "norecords@domain.com"},
			want:    0,
			wantErr: false,
		},
		{
			name:    "single user org result",
			args:    args{authEmail: users[1].Email},
			want:    1,
			wantErr: false,
		},
		{
			name:    "two user org results",
			args:    args{authEmail: users[0].Email},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got UserOrganizations
			err := got.FindByAuthEmail(tt.args.authEmail, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByAuthEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("FindByAuthEmail() got = %v, want %v", got, tt.want)
			}
		})
	}
}
