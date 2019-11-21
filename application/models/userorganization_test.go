package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

func createUserOrganizationFixtures(ms *ModelSuite, t *testing.T) {
	// reset db tables

	singleUuid := domain.GetUuid()
	twoUuid := domain.GetUuid()

	users := []User{
		{
			Email:     "single@domain.com",
			FirstName: "Single",
			LastName:  "Result",
			Nickname:  "Single",
			AdminRole: UserAdminRoleUser,
			Uuid:      singleUuid,
		},
		{
			Email:     "two@domain.com",
			FirstName: "Two",
			LastName:  "Results",
			Nickname:  "Two",
			AdminRole: UserAdminRoleUser,
			Uuid:      twoUuid,
		},
	}
	for i := range users {
		createFixture(ms, &users[i])
	}

	org1uuid := domain.GetUuid()
	org2uuid := domain.GetUuid()

	orgs := []Organization{
		{
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   "",
			AuthConfig: "{}",
			Uuid:       org1uuid,
		},
		{
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   "",
			AuthConfig: "{}",
			Uuid:       org2uuid,
		},
	}
	for i := range orgs {
		createFixture(ms, &orgs[i])
	}

	userOrgs := []UserOrganization{
		{
			OrganizationID: orgs[0].ID,
			UserID:         users[0].ID,
			AuthID:         "one",
			AuthEmail:      "single@domain.com",
		},
		{
			OrganizationID: orgs[0].ID,
			UserID:         users[1].ID,
			AuthID:         "two",
			AuthEmail:      "two@domain.com",
		},
		{
			OrganizationID: orgs[1].ID,
			UserID:         users[1].ID,
			AuthID:         "two",
			AuthEmail:      "two@domain.com",
		},
	}
	for i := range userOrgs {
		createFixture(ms, &userOrgs[i])
	}
}

func (ms *ModelSuite) TestFindByAuthEmail() {
	t := ms.T()
	createUserOrganizationFixtures(ms, t)

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
			args:    args{authEmail: "single@domain.com"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "two user org results",
			args:    args{authEmail: "two@domain.com"},
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
