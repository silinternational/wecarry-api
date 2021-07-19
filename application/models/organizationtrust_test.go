package models

import (
	"testing"
)

type trustFixtures struct {
	Organizations
	OrganizationTrusts
}

func (ms *ModelSuite) TestTrust_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		trust    OrganizationTrust
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			trust: OrganizationTrust{
				PrimaryID:   1,
				SecondaryID: 2,
			},
			wantErr: false,
		},
		{
			name: "missing primary_id",
			trust: OrganizationTrust{
				SecondaryID: 2,
			},
			wantErr:  true,
			errField: "primary_id",
		},
		{
			name: "missing secondary_id",
			trust: OrganizationTrust{
				PrimaryID: 1,
			},
			wantErr:  true,
			errField: "secondary_id",
		},
		{
			name: "primary_id = secondary_id",
			trust: OrganizationTrust{
				SecondaryID: 1,
				PrimaryID:   1,
			},
			wantErr:  true,
			errField: "secondary_equals_primary",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.trust.Validate(DB)
			if test.wantErr {
				ms.True(vErr.Count() != 0, "Expected an error, but did not get one")
				ms.True(len(vErr.Get(test.errField)) > 0,
					"Expected an error on field %v, but got none (errors: %v)",
					test.errField, vErr.Errors)
				return
			}
			ms.False(vErr.HasAny(), "Unexpected error: %v", vErr)
		})
	}
}

func (ms *ModelSuite) TestTrust_Create() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 4)
	trusts := OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[0].ID},
	}
	createFixture(ms, &trusts)

	tests := []struct {
		name    string
		trust   OrganizationTrust
		want    int
		wantErr string
	}{
		{name: "exists", trust: OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID}, want: 1},
		{name: "new", trust: OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: orgs[2].ID}, want: 2},
		{name: "invalid1", trust: OrganizationTrust{PrimaryID: 0, SecondaryID: orgs[1].ID}, wantErr: "must be valid"},
		{name: "invalid2", trust: OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: 0}, wantErr: "must be valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTrust := OrganizationTrust{
				PrimaryID:   tt.trust.PrimaryID,
				SecondaryID: tt.trust.SecondaryID,
			}
			err := newTrust.CreateSymmetric(ms.DB)
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")

			org := Organization{ID: tt.trust.PrimaryID}
			orgs, err := org.TrustedOrganizations(ms.DB)
			ms.NoError(err)

			ms.Equal(tt.want, len(orgs), "incorrect number of OrganizationTrust records")
		})
	}
}

func (ms *ModelSuite) TestTrust_Remove() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 3)
	trusts := OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[0].ID},
	}
	createFixture(ms, &trusts)

	tests := []struct {
		name    string
		trust   OrganizationTrust
		want    int
		wantErr string
	}{
		{name: "not existing", trust: OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: orgs[2].ID}, want: 1},
		{name: "exists", trust: OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID}, want: 0},
		{name: "invalid1", trust: OrganizationTrust{PrimaryID: 0, SecondaryID: orgs[1].ID}, wantErr: "must be valid"},
		{name: "invalid2", trust: OrganizationTrust{PrimaryID: orgs[0].ID, SecondaryID: 0}, wantErr: "must be valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trust OrganizationTrust
			err := trust.RemoveSymmetric(ms.DB, tt.trust.PrimaryID, tt.trust.SecondaryID)
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")

			org1 := Organization{ID: tt.trust.PrimaryID}
			orgs1, err := org1.TrustedOrganizations(ms.DB)
			ms.NoError(err)

			org2 := Organization{ID: tt.trust.SecondaryID}
			orgs2, err := org2.TrustedOrganizations(ms.DB)
			ms.NoError(err)

			ms.Equal(tt.want, len(orgs1)+len(orgs2), "incorrect number of OrganizationTrust records")
		})
	}
}

func (ms *ModelSuite) TestTrust_FindByOrgIDs() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 4)
	trusts := OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[0].ID},
	}
	createFixture(ms, &trusts)

	tests := []struct {
		name    string
		id1     int
		id2     int
		want    OrganizationTrust
		wantErr string
	}{
		{name: "0 and 1", id1: orgs[0].ID, id2: orgs[1].ID, want: trusts[0]},
		{name: "1 and 0", id1: orgs[1].ID, id2: orgs[0].ID, want: trusts[1]},
		{name: "0 and 2", id1: orgs[0].ID, id2: orgs[2].ID, wantErr: "no rows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trust OrganizationTrust
			err := trust.FindByOrgIDs(ms.DB, tt.id1, tt.id2)
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")

			ms.Equal(tt.want.ID, trust.ID)
			ms.Equal(tt.want.PrimaryID, trust.PrimaryID)
			ms.Equal(tt.want.SecondaryID, trust.SecondaryID)
		})
	}
}

func (ms *ModelSuite) TestTrusts_FindByOrgID() {
	t := ms.T()

	orgs := createOrganizationFixtures(ms.DB, 4)
	trusts := OrganizationTrusts{
		{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID},
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[0].ID},
		{PrimaryID: orgs[1].ID, SecondaryID: orgs[2].ID},
		{PrimaryID: orgs[2].ID, SecondaryID: orgs[1].ID},
	}
	createFixture(ms, &trusts)

	tests := []struct {
		name    string
		id      int
		want    OrganizationTrusts
		wantErr string
	}{
		{name: "0", id: orgs[0].ID, want: OrganizationTrusts{trusts[0]}},
		{name: "1", id: orgs[1].ID, want: OrganizationTrusts{trusts[1], trusts[2]}},
		{name: "2", id: orgs[2].ID, want: OrganizationTrusts{trusts[3]}},
		{name: "3", id: orgs[3].ID, want: OrganizationTrusts{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trusts OrganizationTrusts
			err := trusts.FindByOrgID(ms.DB, tt.id)
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")

			ms.Equal(len(tt.want), len(trusts))
			for i := range tt.want {
				ms.Equal(tt.want[i].ID, trusts[i].ID)
				ms.Equal(tt.want[i].PrimaryID, trusts[i].PrimaryID)
				ms.Equal(tt.want[i].SecondaryID, trusts[i].SecondaryID)
			}
		})
	}
}
