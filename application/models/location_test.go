package models

import (
	"math"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate"
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
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
		},
		{
			name: "no geo",
			location: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
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
				Latitude:    nulls.NewFloat64(99.9),
				Longitude:   nulls.NewFloat64(0),
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "technically on Earth, but not standardized",
			location: Location{
				Description: "who knows",
				Country:     "XX",
				Latitude:    nulls.NewFloat64(0),
				Longitude:   nulls.NewFloat64(1000),
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "empty geo coordinates",
			location: Location{
				Description: "null island",
				Country:     "NA",
				Latitude:    nulls.NewFloat64(0),
				Longitude:   nulls.NewFloat64(0),
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "only lat",
			location: Location{
				Description: "only lat",
				Country:     "NA",
				Latitude:    nulls.NewFloat64(1.0),
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "only long",
			location: Location{
				Description: "only long",
				Country:     "NA",
				Longitude:   nulls.NewFloat64(1.0),
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
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
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
			if err := tt.location.Create(); (err != nil) != tt.wantErr {
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
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Toronto, Canada",
				Country:     "CA",
				Latitude:    nulls.NewFloat64(43.6532),
				Longitude:   nulls.NewFloat64(-79.3832),
			},
			want: 1990.8,
		},
		{
			name: "same east-west hemisphere",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Quito, Ecuador",
				Country:     "EC",
				Latitude:    nulls.NewFloat64(-0.1807),
				Longitude:   nulls.NewFloat64(-78.4678),
			},
			want: 2890.6,
		},
		{
			name: "same north-south hemisphere",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Country:     "KR",
				Latitude:    nulls.NewFloat64(37.5665),
				Longitude:   nulls.NewFloat64(126.9780),
			},
			want: 12423.0,
		},
		{
			name: "not valid",
			location1: Location{
				Latitude:  nulls.Float64{},
				Longitude: nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Country:     "KR",
				Latitude:    nulls.NewFloat64(37.5665),
				Longitude:   nulls.NewFloat64(126.9780),
			},
			want: math.NaN(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := test.location1.DistanceKm(test.location2)
			if math.IsNaN(test.want) {
				ms.True(math.IsNaN(d))
			} else {
				ms.InDelta(test.want, d, 0.1)
			}
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
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Seattle, WA, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(47.6062),
				Longitude:   nulls.NewFloat64(-122.3321),
			},
			want: false,
		},
		{
			name: "different countries, but small distance",
			location1: Location{
				Description: "San Diego, CA, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(32.7157),
				Longitude:   nulls.NewFloat64(-117.1611),
			},
			location2: Location{
				Description: "Tijuana, Mexico",
				Country:     "MX",
				Latitude:    nulls.NewFloat64(32.5149),
				Longitude:   nulls.NewFloat64(-117.0382),
			},
			want: true,
		},
		{
			name: "different countries, large distance",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Country:     "KR",
				Latitude:    nulls.NewFloat64(37.5665),
				Longitude:   nulls.NewFloat64(126.9780),
			},
			want: false,
		},
		{
			name: "no country specified, far apart",
			location1: Location{
				Description: "Miami, FL, USA",
				Latitude:    nulls.NewFloat64(25.7617),
				Longitude:   nulls.NewFloat64(-80.1918),
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Latitude:    nulls.NewFloat64(37.5665),
				Longitude:   nulls.NewFloat64(126.9780),
			},
			want: false,
		},
		{
			name: "no country specified, near",
			location1: Location{
				Description: "San Diego, CA, USA",
				Latitude:    nulls.NewFloat64(32.7157),
				Longitude:   nulls.NewFloat64(-117.1611),
			},
			location2: Location{
				Description: "Tijuana, Mexico",
				Latitude:    nulls.NewFloat64(32.5149),
				Longitude:   nulls.NewFloat64(-117.0382),
			},
			want: true,
		},
		{
			name: "no coordinates, far",
			location1: Location{
				Description: "Miami, FL, USA",
				Country:     "US",
			},
			location2: Location{
				Description: "Seoul, Republic of Korea",
				Country:     "KR",
			},
			want: false,
		},
		{
			name: "no coordinates, near",
			location1: Location{
				Description: "San Diego, CA, USA",
				Country:     "US",
			},
			location2: Location{
				Description: "Chula Vista, CA, USA",
				Country:     "US",
			},
			want: false,
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
			err := l.FindByIDs(tt.ids)
			ms.NoError(err)

			got := make([]string, len(l))
			for i, ll := range l {
				got[i] = ll.Description
			}
			ms.Equal(tt.want, got, "incorrect location descriptions")
		})
	}
}
