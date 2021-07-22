package models

import (
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate/v3"
	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestLocation_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		location Location
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "good",
			location: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				State:       "FL",
				City:        "Miami",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
		},
		{
			name: "empty description",
			location: Location{
				Description: "",
				Country:     "US",
			},
			wantErr:  true,
			errField: "description",
		},
		{
			name: "made-up country",
			location: Location{
				Description: "fantasyland",
				Country:     "bogus",
			},
			wantErr:  true,
			errField: "country",
		},
		{
			name: "out in space",
			location: Location{
				Description: "somewhere over the rainbow",
				Country:     "OZ",
				Latitude:    99.9,
				Longitude:   0,
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "technically on Earth, but not standardized",
			location: Location{
				Description: "who knows",
				Country:     "XX",
				Latitude:    0,
				Longitude:   1000,
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "empty geo coordinates",
			location: Location{
				Description: "null island",
				Country:     "NA",
				Latitude:    0,
				Longitude:   0,
			},
			wantErr:  true,
			errField: "geo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.location.Validate(DB)
			if test.wantErr {
				ms.Greater(vErr.Count(), 0, "Expected an error, but did not get one")
				ms.True(len(vErr.Get(test.errField)) != 0,
					"Expected an error on field %v, but got none (errors: %v)",
					test.errField, vErr.Errors)
				return
			}

			ms.False(vErr.HasAny(), "Unexpected error: %v", vErr)
		})
	}
}

func (ms *ModelSuite) TestLocation_Create() {
	t := ms.T()

	tests := []struct {
		name     string
		location Location
		wantErr  bool
	}{
		{
			name: "good",
			location: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				State:       "FL",
				City:        "Miami",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
		},
		{
			name: "empty description",
			location: Location{
				Description: "",
				Country:     "US",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.location.Create(ms.DB); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func (ms *ModelSuite) TestLocation_DistanceKm() {
	t := ms.T()

	tests := []struct {
		name      string
		location1 Location
		location2 Location
		want      float64
	}{
		{
			name: "same hemispheres",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
			location2: Location{
				Description: "Toronto, Canada",
				Country:     "CA",
				Latitude:    43.6532,
				Longitude:   -79.3832,
			},
			want: 1990.8,
		},
		{
			name: "same east-west hemisphere",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
			location2: Location{
				Description: "Quito, Ecuador",
				Country:     "EC",
				Latitude:    -0.1807,
				Longitude:   -78.4678,
			},
			want: 2890.6,
		},
		{
			name: "same north-south hemisphere",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Country:     "KR",
				Latitude:    37.5665,
				Longitude:   126.9780,
			},
			want: 12423.0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := test.location1.DistanceKm(test.location2)
			ms.InDelta(test.want, d, 0.1)
		})
	}
}

func (ms *ModelSuite) TestLocation_IsNear() {
	t := ms.T()

	tests := []struct {
		name      string
		location1 Location
		location2 Location
		want      bool
	}{
		{
			name: "same country",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
			location2: Location{
				Description: "Seattle, WA, USA",
				Country:     "US",
				Latitude:    47.6062,
				Longitude:   -122.3321,
			},
			want: false,
		},
		{
			name: "different countries, but small distance",
			location1: Location{
				Description: "San Diego, CA, USA",
				Country:     "US",
				Latitude:    32.7157,
				Longitude:   -117.1611,
			},
			location2: Location{
				Description: "Tijuana, Mexico",
				Country:     "MX",
				Latitude:    32.5149,
				Longitude:   -117.0382,
			},
			want: true,
		},
		{
			name: "different countries, large distance",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Country:     "KR",
				Latitude:    37.5665,
				Longitude:   126.9780,
			},
			want: false,
		},
		{
			name: "no country specified, far apart",
			location1: Location{
				Description: "Miami, FL, USA",
				Latitude:    25.7617,
				Longitude:   -80.1918,
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Latitude:    37.5665,
				Longitude:   126.9780,
			},
			want: false,
		},
		{
			name: "no country specified, near",
			location1: Location{
				Description: "San Diego, CA, USA",
				Latitude:    32.7157,
				Longitude:   -117.1611,
			},
			location2: Location{
				Description: "Tijuana, Mexico",
				Latitude:    32.5149,
				Longitude:   -117.0382,
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			near := test.location1.IsNear(test.location2)
			ms.Equal(test.want, near)
		})
	}
}

func (ms *ModelSuite) TestLocations_FindByIDs() {
	t := ms.T()

	locations := createLocationFixtures(ms.DB, 3)

	tests := []struct {
		name string
		ids  []int
		want []string
	}{
		{
			name: "good",
			ids:  []int{locations[0].ID, locations[1].ID, locations[0].ID},
			want: []string{locations[0].Description, locations[1].Description},
		},
		{
			name: "missing",
			ids:  []int{99999},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l Locations
			err := l.FindByIDs(ms.DB, tt.ids)
			ms.NoError(err)

			got := make([]string, len(l))
			for i, ll := range l {
				got[i] = ll.Description
			}
			ms.Equal(tt.want, got, "incorrect location descriptions")
		})
	}
}

func (ms *ModelSuite) TestLocations_DeleteUnused() {
	const (
		nUnusedLocations = 2
		nRequests        = 2
		nMeetings        = 2
	)

	_ = createLocationFixtures(ms.DB, nUnusedLocations)

	users := createMeetingFixtures(ms.DB, nMeetings).Users
	nUsers := len(users)

	watches := createWatchFixtures(ms.DB, users)
	nWatches := len(watches)

	watchLocations := createLocationFixtures(ms.DB, nWatches*2)
	for i := range watches {
		watches[i].DestinationID = nulls.NewInt(watchLocations[i*2].ID)
		watches[i].OriginID = nulls.NewInt(watchLocations[i*2+1].ID)
	}
	ms.NoError(ms.DB.Update(&watches))

	_ = createRequestFixtures(ms.DB, nRequests, false, users[0].ID)

	locations := Locations{}

	domain.Env.MaxLocationDelete = 1
	ms.Error(locations.DeleteUnused())

	domain.Env.MaxLocationDelete = 2
	ms.NoError(locations.DeleteUnused())
	n, _ := DB.Count(&locations)

	// 1 user doesn't get a location in createMeetingFixtures()
	ms.Equal(nRequests*2+nMeetings+nWatches*2+nUsers-1, n, "wrong number of locations remain")
}
