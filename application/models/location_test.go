package models

import (
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
				Latitude:    nulls.NewFloat64(0),
			},
			wantErr:  true,
			errField: "geo",
		},
		{
			name: "only long",
			location: Location{
				Description: "only long",
				Country:     "NA",
				Longitude:   nulls.NewFloat64(0),
			},
			wantErr:  true,
			errField: "geo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.location.Validate(DB)
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
			l := &Location{
				ID:          tt.location.ID,
				Description: tt.location.Description,
				Country:     tt.location.Country,
				Latitude:    tt.location.Latitude,
				Longitude:   tt.location.Longitude,
			}
			if err := l.Create(); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
