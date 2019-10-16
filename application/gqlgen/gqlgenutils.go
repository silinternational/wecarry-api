package gqlgen

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/models"
)

func GetStringFromNullsString(inString nulls.String) *string {
	output := ""
	if inString.Valid {
		output = inString.String
	}

	return &output
}

func GetFloat64FromNullsFloat64(in nulls.Float64) *float64 {
	var output float64
	if in.Valid {
		output = in.Float64
	}
	return &output
}

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
	}

	setOptionalStringField(input.Country, &(l.Country))
	setOptionalFloatField(input.Latitude, &(l.Latitude))
	setOptionalFloatField(input.Longitude, &(l.Longitude))

	return l
}
