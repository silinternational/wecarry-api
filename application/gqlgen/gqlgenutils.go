package gqlgen

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/models"
)

func setOptionalStringField(input *string, output *string) {
	if input != nil {
		*output = *input
	}
}

func setOptionalFloatField(input *float64, output *nulls.Float64) {
	if input != nil {
		*output = nulls.NewFloat64(*input)
	}
}

func convertGqlLocationInputToDBLocation(input LocationInput) models.Location {
	l := models.Location{
		Description: input.Description,
		Country:     input.Country,
	}

	setOptionalFloatField(input.Latitude, &l.Latitude)
	setOptionalFloatField(input.Longitude, &l.Longitude)

	return l
}
