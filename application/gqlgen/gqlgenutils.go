package gqlgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

func setStringField(input *string, output *string) {
	if input != nil {
		*output = *input
	}
}

func setOptionalStringField(input *string, output *nulls.String) {
	if input != nil {
		*output = nulls.NewString(*input)
		return
	}
	*output = nulls.String{}
}

func setOptionalFloatField(input *float64, output *nulls.Float64) {
	if input != nil {
		*output = nulls.NewFloat64(*input)
		return
	}
	*output = nulls.Float64{}
}

func convertOptionalLocation(input *LocationInput) *models.Location {
	if input != nil {
		l := convertLocation(*input)
		return &l
	}
	return nil
}

func convertLocation(input LocationInput) models.Location {
	l := models.Location{
		Description: input.Description,
		Country:     input.Country,
	}

	setOptionalFloatField(input.Latitude, &l.Latitude)
	setOptionalFloatField(input.Longitude, &l.Longitude)

	return l
}

func convertUserPreferencesToStandardPreferences(input *UpdateUserPreferencesInput) (models.StandardPreferences, error) {
	if input == nil {
		return models.StandardPreferences{}, nil
	}

	stPrefs := models.StandardPreferences{}

	if input.Language != nil {
		lang := strings.ToLower(fmt.Sprintf("%v", *input.Language))
		if !domain.IsLanguageAllowed(lang) {
			return models.StandardPreferences{}, errors.New("user preference language not allowed ... " + lang)
		}
		stPrefs.Language = lang
	}

	if input.TimeZone != nil {
		if !domain.IsTimeZoneAllowed(*input.TimeZone) {
			return models.StandardPreferences{}, errors.New("user preference time zone not allowed ... " + *input.TimeZone)
		}
		stPrefs.TimeZone = *input.TimeZone
	}

	if input.WeightUnit != nil {
		unit := strings.ToLower(fmt.Sprintf("%v", *input.WeightUnit))
		if !domain.IsWeightUnitAllowed(unit) {
			return models.StandardPreferences{}, errors.New("user preference weight unit not allowed ... " + unit)
		}
		stPrefs.WeightUnit = unit
	}

	return stPrefs, nil
}
