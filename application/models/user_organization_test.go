package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
)

func createUserOrganizationFixtures(ms *ModelSuite, t *testing.T) {
	// reset db tables
	ResetTables(ms.T(), ms.DB)

	singleUuid := domain.GetUuid()
	twoUuid := domain.GetUuid()

	users := []User{
		{
			ID:        1,
			Email:     "single@domain.com",
			FirstName: "Single",
			LastName:  "Result",
			Nickname:  "Single",
			AdminRole: nulls.String{},
			Uuid:      singleUuid,
		},
		{
			ID:        2,
			Email:     "two@domain.com",
			FirstName: "Two",
			LastName:  "Results",
			Nickname:  "Two",
			AdminRole: nulls.String{},
			Uuid:      twoUuid,
		},
	}
	for _, u := range users {
		err := ms.DB.Create(&u)
		if err != nil {
			t.Errorf("unable to create fixture user: %s", err)
			t.FailNow()
		}
	}

	org1uuid := domain.GetUuid()
	org2uuid := domain.GetUuid()

	orgs := []Organization{
		{
			ID:         1,
			Name:       "Org1",
			Url:        nulls.String{},
			AuthType:   "",
			AuthConfig: "{}",
			Uuid:       org1uuid,
		},
		{
			ID:         2,
			Name:       "Org2",
			Url:        nulls.String{},
			AuthType:   "",
			AuthConfig: "{}",
			Uuid:       org2uuid,
		},
	}
	for _, o := range orgs {
		err := ms.DB.Create(&o)
		if err != nil {
			t.Errorf("unable to create fixture organization: %s", err)
			t.FailNow()
		}
	}

	userOrgs := []UserOrganization{
		{
			ID:             1,
			OrganizationID: 1,
			UserID:         1,
			AuthID:         "one",
			AuthEmail:      "single@domain.com",
		},
		{
			ID:             2,
			OrganizationID: 1,
			UserID:         2,
			AuthID:         "two",
			AuthEmail:      "two@domain.com",
		},
		{
			ID:             3,
			OrganizationID: 2,
			UserID:         2,
			AuthID:         "two",
			AuthEmail:      "two@domain.com",
		},
	}
	for _, uo := range userOrgs {
		err := ms.DB.Create(&uo)
		if err != nil {
			t.Errorf("unable to create fixture user_organization: %s", err)
			t.FailNow()
		}
	}
}

func (ms *ModelSuite) TestFindByAuthEmail() {
	t := ms.T()
	ResetTables(t, ms.DB)
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
