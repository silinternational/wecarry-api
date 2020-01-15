package models

import (
	"testing"
)

type trustFixtures struct {
	Organizations
	Trusts
}

func (ms *ModelSuite) TestTrust_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		trust    Trust
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			trust: Trust{
				PrimaryID:   1,
				SecondaryID: 2,
			},
			wantErr: false,
		},
		{
			name: "missing primary_id",
			trust: Trust{
				SecondaryID: 2,
			},
			wantErr:  true,
			errField: "primary_id",
		},
		{
			name: "missing secondary_id",
			trust: Trust{
				PrimaryID: 1,
			},
			wantErr:  true,
			errField: "secondary_id",
		},
		{
			name: "primary_id = secondary_id",
			trust: Trust{
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

func createTrustFixtures(ms *ModelSuite) trustFixtures {
	orgs := createOrganizationFixtures(ms.DB, 4)

	trusts := make([]Trust, 2)
	trusts[0].PrimaryID = orgs[1].ID
	trusts[0].SecondaryID = orgs[0].ID
	trusts[1].PrimaryID = orgs[1].ID
	trusts[1].SecondaryID = orgs[2].ID
	for i := range trusts {
		mustCreate(ms.DB, &trusts[i])
	}

	return trustFixtures{
		Organizations: orgs,
		Trusts:        trusts,
	}
}

func (ms *ModelSuite) TestTrust_Create() {
	t := ms.T()

	orgs := createTrustFixtures(ms).Organizations

	tests := []struct {
		name    string
		trust   Trust
		want    int
		wantErr string
	}{
		{name: "0 and 1", trust: Trust{PrimaryID: orgs[0].ID, SecondaryID: orgs[1].ID}, want: 1},
		{name: "1 and 0", trust: Trust{PrimaryID: orgs[1].ID, SecondaryID: orgs[0].ID}, want: 2},
		{name: "0 and 2", trust: Trust{PrimaryID: orgs[0].ID, SecondaryID: orgs[2].ID}, want: 2},
		{name: "2 and 0", trust: Trust{PrimaryID: orgs[2].ID, SecondaryID: orgs[0].ID}, want: 2},
		{name: "invalid1", trust: Trust{PrimaryID: 0, SecondaryID: orgs[1].ID}, wantErr: "must be valid"},
		{name: "invalid2", trust: Trust{PrimaryID: orgs[0].ID, SecondaryID: 0}, wantErr: "must be valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTrust := Trust{
				PrimaryID:   tt.trust.PrimaryID,
				SecondaryID: tt.trust.SecondaryID,
			}
			err := newTrust.Create()
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")

			org := Organization{ID: tt.trust.PrimaryID}
			orgs, err := org.TrustedOrganizations()
			ms.NoError(err)

			ms.Equal(tt.want, len(orgs), "incorrect number of Trust records")
		})
	}
}

func (ms *ModelSuite) TestTrust_FindByOrgIDs() {
	t := ms.T()

	f := createTrustFixtures(ms)

	tests := []struct {
		name    string
		id1     int
		id2     int
		want    Trust
		wantErr string
	}{
		{name: "0 and 1", id1: f.Organizations[0].ID, id2: f.Organizations[1].ID, want: f.Trusts[0]},
		{name: "1 and 0", id1: f.Organizations[1].ID, id2: f.Organizations[0].ID, want: f.Trusts[0]},
		{name: "0 and 2", id1: f.Organizations[0].ID, id2: f.Organizations[2].ID, wantErr: "no rows"},
		{name: "2 and 0", id1: f.Organizations[2].ID, id2: f.Organizations[0].ID, wantErr: "no rows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trust Trust
			err := trust.FindByOrgIDs(tt.id1, tt.id2)
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

	f := createTrustFixtures(ms)

	tests := []struct {
		name    string
		id      int
		want    Trusts
		wantErr string
	}{
		{name: "0", id: f.Organizations[0].ID, want: Trusts{f.Trusts[0]}},
		{name: "1", id: f.Organizations[1].ID, want: Trusts{f.Trusts[0], f.Trusts[1]}},
		{name: "2", id: f.Organizations[2].ID, want: Trusts{f.Trusts[1]}},
		{name: "3", id: f.Organizations[3].ID, want: Trusts{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trusts Trusts
			err := trusts.FindByOrgID(tt.id)
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
