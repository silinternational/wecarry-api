package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"

	"testing"
)

func (ms *ModelSuite) TestFindOrgByUUID() {
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

	if err := ms.DB.Destroy(&org); err != nil {
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
				Uuid:       domain.GetUuid(),
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
			},
			wantErr: false,
		},
		{
			name: "missing auth config",
			org: Organization{
				Name:     "Bits 'R' Us",
				Uuid:     domain.GetUuid(),
				AuthType: AuthTypeSaml,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ms.DB.Create(&test.org)
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
			if err := ms.DB.Destroy(&test.org); err != nil {
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
				AuthType:   AuthTypeSaml,
				AuthConfig: "{}",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			org: Organization{
				Uuid:       domain.GetUuid(),
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
				Uuid:       domain.GetUuid(),
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

func (ms *ModelSuite) TestOrganizationFindByDomain() {
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
				} else if org.Uuid != test.want.Uuid {
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

func createOrgFixtures(ms *ModelSuite, t *testing.T) (Organization, OrganizationDomain) {
	// Load Organization test fixtures
	org := Organization{
		Name:       "ACME",
		Uuid:       domain.GetUuid(),
		AuthType:   AuthTypeSaml,
		AuthConfig: "{}",
	}
	if err := ms.DB.Create(&org); err != nil {
		t.Errorf("could not create org fixtures ... %v", err)
		t.FailNow()
	}

	// Load Organization Domains test fixtures
	orgDomain := OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "example.org",
	}
	if err := ms.DB.Create(&orgDomain); err != nil {
		t.Errorf("could not create org domain fixtures ... %v", err)
		t.FailNow()
	}

	return org, orgDomain
}

func (ms *ModelSuite) TestOrganization_AddRemoveDomain() {
	t := ms.T()

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
	for i := range orgFixtures {
		err := ms.DB.Create(&orgFixtures[i])
		if err != nil {
			t.Errorf("Unable to create org fixture: %s", err)
			t.FailNow()
		}
	}

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
	for i := range orgFixtures {
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
		Uuid:       domain.GetUuid(),
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
	for i := range orgFixtures {
		err := ms.DB.Create(&orgFixtures[i])
		if err != nil {
			t.Errorf("Unable to create org fixture: %s", err)
			t.FailNow()
		}
	}

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
