package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"

	"github.com/silinternational/wecarry-api/auth/azureadv2"
	"github.com/silinternational/wecarry-api/auth/google"
	"github.com/silinternational/wecarry-api/domain"
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
	var file File
	ms.Nil(file.Store("logo.gif", []byte("GIF89a")), "unexpected error storing file")

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
				FileID:     nulls.NewInt(file.ID),
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

	err := orgFixtures[0].AddDomain("first.com", "", "")
	ms.NoError(err, "unable to add first domain to Org1: %s", err)
	domains, _ := orgFixtures[0].Domains()
	if len(domains) != 1 {
		t.Errorf("did not get error, but failed to add first domain to Org1")
	}

	err = orgFixtures[0].AddDomain("second.com", "", "")
	ms.NoError(err, "unable to add second domain to Org1: %s", err)
	domains, _ = orgFixtures[0].Domains()
	if len(domains) != 2 {
		t.Errorf("did not get error, but failed to add second domain to Org1")
	}

	err = orgFixtures[1].AddDomain("second.com", "", "")
	ms.Error(err, "was able to add existing domain (second.com) to Org2 but should have gotten error")
	domains, _ = orgFixtures[0].Domains()
	ms.Equal(2, len(domains), "after reloading org domains we did not get what we expected")

	err = orgFixtures[0].RemoveDomain("first.com")
	ms.NoError(err, "unable to remove domain: %s", err)
	domains, _ = orgFixtures[0].Domains()
	ms.Equal(1, len(domains), "org domains count after removing domain is not correct")
}

func (ms *ModelSuite) TestOrganization_Save() {
	t := ms.T()

	orgFixtures := createOrganizationFixtures(ms.DB, 2)
	var file File
	ms.Nil(file.Store("logo.gif", []byte("GIF89a")), "unexpected error storing file")

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
		FileID:     nulls.NewInt(file.ID),
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

	orgDomains, err := f.Organizations[0].Domains()
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

	allOrgIDs := []string{ofs[0].Name, ofs[1].Name, ofs[2].Name, ofs[3].Name, ofs[4].Name}

	tests := []struct {
		name       string
		user       User
		wantOrgIDs []string
		wantErr    string
	}{
		{name: "SuperAdmin", user: ufs[0], wantOrgIDs: allOrgIDs},
		{name: "SalesAdmin", user: ufs[1], wantOrgIDs: allOrgIDs},
		{name: "NoRole", user: ufs[2], wantOrgIDs: []string{}},
		{name: "Admin User 4", user: ufs[3], wantOrgIDs: []string{ofs[3].Name}},
		{name: "Admin User 5", user: ufs[4], wantOrgIDs: []string{ofs[4].Name, ofs[3].Name}},
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
			gotNames := make([]string, len(got))
			for i := range got {
				gotNames[i] = got[i].Name
			}
			ms.Equal(test.wantOrgIDs, gotNames, "incorrect list of org Names")
		})
	}

}

func (ms *ModelSuite) TestOrganization_LogoURL() {
	t := ms.T()

	orgFixtures := createOrganizationFixtures(ms.DB, 1)
	var file File
	ms.NoError(ms.DB.Find(&file, orgFixtures[0].FileID.Int))
	logoURL := file.URL

	tests := []struct {
		name    string
		org     Organization
		want    string
		wantNil bool
		wantErr string
	}{
		{name: "good", org: orgFixtures[0], want: logoURL},
		{name: "bad", org: Organization{FileID: nulls.NewInt(1)}, wantErr: "no rows in result set"},
		{name: "no file", org: Organization{}, wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logoURL, err := tt.org.LogoURL()
			if tt.wantErr != "" {
				ms.Errorf(err, "expected error %s but got no error", tt.wantErr, err)
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err)
			if tt.wantNil {
				ms.Nil(logoURL)
				return
			}
			ms.Equal(tt.want, *logoURL)
		})
	}
}

func (ms *ModelSuite) TestOrganization_CreateTrust() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 4)
	trust := OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID}
	ms.NoError(trust.CreateSymmetric())

	tests := []struct {
		name      string
		primary   Organization
		secondary Organization
		want      int
		wantErr   string
	}{
		{name: "pre-existing", primary: orgs[0], secondary: orgs[1], want: 1},
		{name: "reverse exists", primary: orgs[1], secondary: orgs[0], want: 1},
		{name: "new for org 0", primary: orgs[0], secondary: orgs[2], want: 2},
		{name: "new for org 1", primary: orgs[1], secondary: orgs[2], want: 2},
		{name: "new for org 3", primary: orgs[3], secondary: orgs[0], want: 1},
		{name: "bad org", primary: Organization{}, secondary: orgs[0], wantErr: "must be valid"},
		{name: "bad org2", primary: orgs[0], secondary: Organization{}, wantErr: "no rows in result set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.primary.CreateTrust(tt.secondary.UUID.String())
			if tt.wantErr != "" {
				ms.Errorf(err, "expected error %s but got no error", tt.wantErr, err)
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err)
			trustedOrgs, err := tt.primary.TrustedOrganizations()
			ms.NoError(err)
			ms.Equal(tt.want, len(trustedOrgs))
		})
	}
}

func (ms *ModelSuite) TestOrganization_RemoveTrust() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 4)
	trusts := OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[2].ID},
	}
	ms.NoError(trusts[0].CreateSymmetric())
	ms.NoError(trusts[1].CreateSymmetric())

	tests := []struct {
		name      string
		primary   Organization
		secondary Organization
		want      int
		wantErr   string
	}{
		{name: "exists", primary: orgs[0], secondary: orgs[1], want: 0},
		{name: "reverse exists", primary: orgs[2], secondary: orgs[1], want: 0},
		{name: "not existing", primary: orgs[3], secondary: orgs[0], want: 0},
		{name: "bad org", primary: Organization{}, secondary: orgs[0], wantErr: "must be valid"},
		{name: "bad org2", primary: orgs[0], secondary: Organization{}, wantErr: "no rows in result set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.primary.RemoveTrust(tt.secondary.UUID.String())
			if tt.wantErr != "" {
				ms.Errorf(err, "expected error %s but got no error", tt.wantErr, err)
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err)
			trustedOrgs, err := tt.primary.TrustedOrganizations()
			ms.NoError(err)
			ms.Equal(tt.want, len(trustedOrgs))
		})
	}
}

func (ms *ModelSuite) TestOrganization_TrustedOrganizations() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 4)
	trusts := OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[2].ID, SecondaryID: orgs[0].ID},
	}
	ms.NoError(trusts[0].CreateSymmetric())
	ms.NoError(trusts[1].CreateSymmetric())

	tests := []struct {
		name    string
		primary Organization
		want    []int
	}{
		{name: "exists", primary: orgs[0], want: []int{orgs[1].ID, orgs[2].ID}},
		{name: "exists", primary: orgs[1], want: []int{orgs[0].ID}},
		{name: "exists", primary: orgs[2], want: []int{orgs[0].ID}},
		{name: "exists", primary: orgs[3], want: []int{}},
		{name: "bad org", primary: Organization{}, want: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.primary.TrustedOrganizations()
			ms.NoError(err, "unexpected error from UUT")
			ms.Equal(len(tt.want), len(got), "received incorrect number of orgs")
			ids := make([]int, len(got))
			for i := range got {
				ids[i] = got[i].ID
				ms.Contains(tt.want, got[i].ID, "received unexpected org in trusted list")
			}
			for i := range tt.want {
				ms.Contains(ids, tt.want[i], "didn't get expected org in trusted list")
			}
		})
	}
}

func (ms *ModelSuite) TestOrganization_AttachLogo() {
	orgs := createOrganizationFixtures(ms.DB, 3)
	files := createFileFixtures(3)
	orgs[1].FileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&orgs[1], "file_id"))

	tests := []struct {
		name    string
		org     Organization
		oldLogo *File
		newLogo string
		want    File
		wantErr string
	}{
		{
			name:    "no previous file",
			org:     orgs[0],
			newLogo: files[1].UUID.String(),
			want:    files[1],
		},
		{
			name:    "previous file",
			org:     orgs[1],
			oldLogo: &files[0],
			newLogo: files[2].UUID.String(),
			want:    files[2],
		},
		{
			name:    "bad ID",
			org:     orgs[2],
			newLogo: uuid.UUID{}.String(),
			wantErr: "no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.org.AttachLogo(tt.newLogo)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(tt.want.UUID.String(), got.UUID.String(), "wrong file returned")
			ms.Equal(true, got.Linked, "new logo file is not marked as linked")
			if tt.oldLogo != nil {
				ms.Equal(false, tt.oldLogo.Linked, "old logo file is not marked as unlinked")
			}
		})
	}
}

func (ms *ModelSuite) TestOrganization_GetAuthProvider() {

	adTenantID := "TestADTenant"
	adSecret := "TestADKey"
	adApplicationID := "testADSecret"

	adAuthConfig :=
		`{
    "TenantID": "` + adTenantID + `",
    "ClientSecret": "` + adSecret + `",
    "ApplicationID": "` + adApplicationID + `"
}`

	domain.Env.GoogleKey = "TestGoogleKey"
	domain.Env.GoogleSecret = "testGoogleSecret"

	uid := domain.GetUUID()
	org := Organization{
		Name:       "testorg1",
		AuthType:   AuthTypeAzureAD,
		AuthConfig: adAuthConfig,
		UUID:       uid,
	}
	err := org.Save()
	ms.NoError(err, "unable to create organization fixture")

	orgDomain1 := OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "domain1.com",
		AuthType:       AuthTypeDefault,
		AuthConfig:     "",
	}
	err = orgDomain1.Save()
	ms.NoError(err, "unable to create orgDomain1 fixture")

	orgDomain2 := OrganizationDomain{
		OrganizationID: org.ID,
		Domain:         "domain2.com",
		AuthType:       AuthTypeGoogle,
		AuthConfig:     "",
	}
	err = orgDomain2.Save()
	ms.NoError(err, "unable to create orgDomain2 fixture")

	var o Organization
	err = o.FindByUUID(uid.String())
	ms.NoError(err, "unable to find organization fixture")

	// should get type azuread:
	provider, err := o.GetAuthProvider("test@domain1.com")
	ms.NoError(err, "unable to get authprovider for test@domain1.com")
	ms.IsType(&azureadv2.Provider{}, provider, "auth provider not expected azureAD type")

	// should get type google:
	provider, err = o.GetAuthProvider("test@domain2.com")
	ms.NoError(err, "unable to get authprovider for test@domain2.com")
	ms.IsType(&google.Provider{}, provider, "auth provider not expected google type")
}
