package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"

	"testing"
)

func (ms *ModelSuite) TestOrganization_FindOrgByUUID() {
	t := ms.T()
	org, _ := createOrgFixtures(ms, t)

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
			args: args{org.UUID.String()},
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
				} else if org.UUID != test.want.UUID {
					t.Errorf("found %v, expected %v", org, test.want)
				}
			}
		})
	}

	if err := ms.DB.Destroy(&org); err != nil {
		t.Errorf("error deleting test data: %v", err)
	}
}

func (ms *ModelSuite) TestOrganization_Create() {
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
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
				Url:        nulls.NewString("https://www.example.com"),
			},
			wantErr: false,
		},
		{
			name: "minimum",
			org: Organization{
				Name:       "Bits 'R' Us",
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
			},
			wantErr: false,
		},
		{
			name: "missing auth config",
			org: Organization{
				Name:     "Bits 'R' Us",
				AuthType: AuthTypeSaml,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.org.Save()
			if test.wantErr == true {
				if err == nil {
					t.Errorf("Expected an error, but did not get one")
				}
			} else if err != nil {
				t.Errorf("Unexpected error %v", err)
			} else {
				var org Organization
				err := org.FindByUUID(test.org.UUID.String())
				if err != nil {
					t.Errorf("Couldn't find new org %v: %v", test.org.Name, err)
				}
				if org.UUID != test.org.UUID {
					t.Errorf("newly created org doesn't match, found %v, expected %v", org, test.org)
				}
			}

			// clean up
			if err := ms.DB.Destroy(&test.org); err != nil {
				t.Errorf("error deleting test data: %v", err)
			}
		})
	}
}

func (ms *ModelSuite) TestOrganization_Validate() {
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
				UUID:       domain.GetUUID(),
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			org: Organization{
				UUID:       domain.GetUUID(),
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "missing uuid",
			org: Organization{
				Name:       "Babelfish Warehouse",
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing auth type",
			org: Organization{
				Name:       "Babelfish Warehouse",
				UUID:       domain.GetUUID(),
				AuthConfig: "{}",
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

func (ms *ModelSuite) TestOrganization_FindByDomain() {
	t := ms.T()
	org, orgDomain := createOrgFixtures(ms, t)

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
				} else if org.UUID != test.want.UUID {
					t.Errorf("found %v, expected %v", org, test.want)
				}
			}
		})
	}

	// delete org fixture and org domain by cascading delete
	if err := ms.DB.Destroy(&org); err != nil {
		t.Errorf("error deleting test data: %v", err)
	}
}

func (ms *ModelSuite) TestOrganization_AddRemoveDomain() {
	t := ms.T()

	orgFixtures := createOrganizationFixtures(ms.DB, 2)

	err := orgFixtures[0].AddDomain("first.com")
	ms.NoError(err, "unable to add first domain to Org1: %s", err)
	domains, _ := orgFixtures[0].GetDomains()
	if len(domains) != 1 {
		t.Errorf("did not get error, but failed to add first domain to Org1")
	}

	err = orgFixtures[0].AddDomain("second.com")
	ms.NoError(err, "unable to add second domain to Org1: %s", err)
	domains, _ = orgFixtures[0].GetDomains()
	if len(domains) != 2 {
		t.Errorf("did not get error, but failed to add second domain to Org1")
	}

	err = orgFixtures[1].AddDomain("second.com")
	ms.Error(err, "was to add existing domain (second.com) to Org2 but should have gotten error")
	domains, _ = orgFixtures[0].GetDomains()
	if len(domains) != 2 {
		t.Errorf("after reloading org domains we did not get what we expected (%v), got: %v", 2, len(orgFixtures[0].OrganizationDomains))
	}

	err = orgFixtures[0].RemoveDomain("first.com")
	ms.NoError(err, "unable to remove domain: %s", err)
	domains, _ = orgFixtures[0].GetDomains()
	if len(domains) != 1 {
		t.Errorf("org domains count after removing domain is not correct, expected %v, got: %v", 1, len(orgFixtures[0].OrganizationDomains))
	}

}

func (ms *ModelSuite) TestOrganization_Save() {
	t := ms.T()

	orgFixtures := []Organization{
		{ID: 1, Name: "Org1"},
		{ID: 2, Name: "Org2"},
	}

	for i := range orgFixtures {
		orgFixtures[i].CreatedAt = time.Time{}
		orgFixtures[i].UpdatedAt = time.Time{}
		orgFixtures[i].Url = nulls.String{}
		orgFixtures[i].AuthType = "na"
		orgFixtures[i].AuthConfig = "{}"
		orgFixtures[i].UUID = domain.GetUUID()

		err := ms.DB.Create(&orgFixtures[i])
		if err != nil {
			t.Errorf("Unable to create org fixture: %s", err)
			t.FailNow()
		}
	}

	// test save of existing organization
	orgFixtures[0].Name = "changed"
	err := orgFixtures[0].Save()
	if err != nil {
		t.Error(err)
	}
	// load org from ms.DB.to ensure change saved
	var found Organization
	err = ms.DB.Where("id = ?", orgFixtures[0].ID).First(&found)
	if err != nil {
		t.Error(err)
	}

	if found.Name != orgFixtures[0].Name {
		t.Errorf("Org name not changed after save, wanted: %s, have: %s", orgFixtures[0].Name, found.Name)
	}

	// create new org
	newOrg := Organization{
		Name:       "new org",
		Url:        nulls.String{},
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
		UUID:       domain.GetUUID(),
	}

	err = newOrg.Save()
	if err != nil {
		t.Error(err)
	}

	if newOrg.ID == 0 {
		t.Error("new organization not updated after save")
	}

}

func (ms *ModelSuite) TestOrganization_All() {
	t := ms.T()

	orgFixtures := createOrganizationFixtures(ms.DB, 2)

	var allOrgs Organizations
	err := allOrgs.All()
	if err != nil {
		t.Error(err)
	}

	if len(allOrgs) != len(orgFixtures) {
		t.Errorf("Did not get expected number of orgs, got %v, wanted %v", len(allOrgs), len(orgFixtures))
	}

}

func (ms *ModelSuite) TestOrganization_GetDomains() {
	f := CreateFixturesForOrganizationGetDomains(ms)

	orgDomains, err := f.Organizations[0].GetDomains()
	ms.NoError(err)

	domains := make([]string, len(orgDomains))
	for i := range orgDomains {
		domains[i] = orgDomains[i].Domain
	}

	expected := []string{
		f.OrganizationDomains[1].Domain,
		f.OrganizationDomains[2].Domain,
		f.OrganizationDomains[0].Domain,
	}

	ms.Equal(expected, domains, "incorrect list of domains")
}

func (ms *ModelSuite) TestOrganization_GetUsers() {
	t := ms.T()

	f := createFixturesForOrganizationGetUsers(ms)

	tests := []struct {
		name        string
		org         Organization
		wantUserIDs []int
		wantErr     string
	}{
		{name: "org 0", org: f.Organizations[0], wantUserIDs: []int{f.Users[0].ID, f.Users[2].ID, f.Users[1].ID}},
		{name: "non-existent org", org: Organization{}, wantErr: "invalid Organization ID"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users, err := test.org.GetUsers()

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}

			ms.NoError(err)
			userIDs := make([]int, len(users))
			for i := range users {
				userIDs[i] = users[i].ID
			}
			ms.Equal(test.wantUserIDs, userIDs)
		})
	}
}

func (ms *ModelSuite) TestOrganization_AllWhereUserIsOrgAdmin() {
	t := ms.T()

	f := createFixturesForOrganization_AllWhereUserIsOrgAdmin(ms)
	ofs := f.Organizations
	ufs := f.Users

	allOrgIDs := []int{ofs[0].ID, ofs[1].ID, ofs[2].ID, ofs[3].ID, ofs[4].ID}

	tests := []struct {
		name       string
		user       User
		wantOrgIDs []int
		wantErr    string
	}{
		{name: "SuperAdmin", user: ufs[0], wantOrgIDs: allOrgIDs},
		{name: "SalesAdmin", user: ufs[1], wantOrgIDs: allOrgIDs},
		{name: "NoRole", user: ufs[2], wantOrgIDs: []int{}},
		{name: "SingleAdmin", user: ufs[3], wantOrgIDs: []int{ofs[3].ID}},
		{name: "DoubleAdmin", user: ufs[4], wantOrgIDs: []int{ofs[4].ID}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got Organizations
			err := got.AllWhereUserIsOrgAdmin(test.user)

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr)
				return
			}

			ms.NoError(err)
			gotIDs := make([]int, len(got))
			for i := range got {
				gotIDs[i] = got[i].ID
			}
			ms.Equal(test.wantOrgIDs, gotIDs, "incorrect list of org IDs")
		})
	}

}
