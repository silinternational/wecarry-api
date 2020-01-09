package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"

	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestWatch_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		watch    Watch
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			watch: Watch{
				UUID:    domain.GetUUID(),
				OwnerID: 1,
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			watch: Watch{
				OwnerID: 1,
			},
			wantErr:  true,
			errField: "uuid",
		},
		{
			name: "missing owner_id",
			watch: Watch{
				UUID: domain.GetUUID(),
			},
			wantErr:  true,
			errField: "owner_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.watch.Validate(DB)
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

func createWatchFixtures(tx *pop.Connection, users Users) Watches {
	watches := make(Watches, len(users)*2)
	locations := createLocationFixtures(tx, len(watches))
	for i := range watches {
		watches[i].UUID = domain.GetUUID()
		watches[i].OwnerID = users[i/2].ID
		watches[i].LocationID = nulls.NewInt(locations[i].ID)
		mustCreate(tx, &watches[i])
	}
	return watches
}

func (ms *ModelSuite) TestWatch_FindByUUID() {
	t := ms.T()

	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 2).Users)

	tests := []struct {
		name    string
		uuid    string
		want    Watch
		wantErr string
	}{
		{name: "user 0", uuid: watches[0].UUID.String(), want: watches[0]},
		{name: "user 1", uuid: watches[2].UUID.String(), want: watches[2]},
		{name: "blank uuid", uuid: "", wantErr: "watch uuid must not be blank"},
		{name: "wrong uuid", uuid: domain.GetUUID().String(), wantErr: "no rows in result set"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var watch Watch
			err := watch.FindByUUID(test.uuid)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "wrong error type")
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(test.want.UUID, watch.UUID, "incorrect uuid")
		})
	}
}

func (ms *ModelSuite) TestWatches_FindByUser() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	watches := createWatchFixtures(ms.DB, users)
	noWatches := createUserFixtures(ms.DB, 1).Users[0]

	tests := []struct {
		name string
		user User
		want Watches
	}{
		{name: "user 0", user: users[0], want: Watches{watches[1], watches[0]}},
		{name: "user 1", user: users[1], want: Watches{watches[3], watches[2]}},
		{name: "no watches", user: noWatches, want: Watches{}},
		{name: "wrong user", user: User{}, want: Watches{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Watches{}
			err := got.FindByUser(test.user)
			ms.NoError(err, "unexpected error")

			gotIDs := make([]string, len(got))
			for i := range got {
				gotIDs[i] = got[i].UUID.String()
			}

			wantIDs := make([]string, len(test.want))
			for i := range test.want {
				wantIDs[i] = test.want[i].UUID.String()
			}

			ms.Equal(wantIDs, gotIDs, "wrong list of watches")
		})
	}
}

func (ms *ModelSuite) TestWatch_GetOwner() {
	users := createUserFixtures(ms.DB, 2).Users
	watches := createWatchFixtures(ms.DB, users)

	owner, err := watches[0].GetOwner()
	ms.NoError(err, "unexpected error")
	ms.Equal(users[0].UUID, owner.UUID, "incorrect owner")
}

func (ms *ModelSuite) TestWatch_GetSetLocation() {
	newLoc := createLocationFixtures(ms.DB, 1)[0]
	watches := createWatchFixtures(ms.DB, createUserFixtures(ms.DB, 1).Users)

	err := watches[0].SetLocation(newLoc)
	ms.NoError(err, "unexpected error from SetLocation()")

	got, err := watches[0].GetLocation()
	ms.NoError(err, "unexpected error from GetLocation()")
	ms.Equal(newLoc.Country, got.Country, "country doesn't match")
	ms.Equal(newLoc.Description, got.Description, "description doesn't match")
	ms.InDelta(newLoc.Latitude.Float64, got.Latitude.Float64, 0.0001, "latitude doesn't match")
	ms.InDelta(newLoc.Longitude.Float64, got.Longitude.Float64, 0.0001, "longitude doesn't match")
}
