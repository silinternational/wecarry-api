package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
)

func createUserOrganizationFixtures(t *testing.T) {
	// reset db tables
	resetTables(t)

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
		err := DB.Create(&u)
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
		err := DB.Create(&o)
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
		err := DB.Create(&uo)
		if err != nil {
			t.Errorf("unable to create fixture user_organization: %s", err)
			t.FailNow()
		}
	}
}

func (ms *ModelSuite) TestFindByAuthEmail() {
	t := ms.T()
	resetTables(t)
	createUserOrganizationFixtures(t)

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
			got, err := UserOrganizationFindByAuthEmail(tt.args.authEmail, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserOrganizationFindByAuthEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("UserOrganizationFindByAuthEmail() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestFindUserOrganization() {
	t := ms.T()
	resetTables(t)
	createUserOrganizationFixtures(t)

	type args struct {
		user User
		org  Organization
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "user 1, org 1",
			args: args{
				user: User{Email: "single@domain.com"},
				org:  Organization{Name: "Org1"},
			},
			wantErr: false,
		},
		{
			name: "user 2, org 1",
			args: args{
				user: User{Email: "two@domain.com"},
				org:  Organization{Name: "Org1"},
			},
			wantErr: false,
		},
		{
			name: "user 2, org 2",
			args: args{
				user: User{Email: "two@domain.com"},
				org:  Organization{Name: "Org2"},
			},
			wantErr: false,
		},
		{
			name: "user 1, org 2",
			args: args{
				user: User{Email: "single@domain.com"},
				org:  Organization{Name: "Org2"},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var user User
			if err := DB.Where("email = ?", test.args.user.Email).First(&user); err != nil {
				t.Errorf("couldn't find test user '%v'", test.args.user.Email)
			}

			var org Organization
			if err := DB.Where("name = ?", test.args.org.Name).First(&org); err != nil {
				t.Errorf("couldn't find test org '%v'", test.args.org.Name)
			}

			uo, err := FindUserOrganization(user, org)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("FindOrgByUUID() returned an error: %v", err)
				} else if (uo.UserID != user.ID) || (uo.OrganizationID != org.ID) {
					t.Errorf("received wrong UserOrganization (UserID=%v, OrganizationID=%v), expected (user.ID=%v, org.ID=%v)",
						uo.UserID, uo.OrganizationID, user.ID, org.ID)
				}
			}
		})
	}
}
