package gqlgen

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/nulls"
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

// getFunctionName provides the filename, line number, and function name of the caller, skipping the top `skip`
// functions on the stack.
func getFunctionName(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "?"
	}

	fn := runtime.FuncForPC(pc)
	return fmt.Sprintf("%s:%d %s", file, line, fn.Name())
}

// reportError logs an error with details, and returns a user-friendly, translated error identified by translation key
// string `errID`.
func reportError(ctx context.Context, err error, errID string, extras ...map[string]interface{}) error {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	allExtras := map[string]interface{}{
		"query":    graphql.GetRequestContext(ctx).RawQuery,
		"function": getFunctionName(2),
	}
	for _, e := range extras {
		for key, val := range e {
			allExtras[key] = val
		}
	}

	errStr := errID
	if err != nil {
		errStr = err.Error()
	}
	domain.Error(c, errStr, allExtras)

	if domain.T == nil {
		return errors.New(errID)
	}
	return errors.New(domain.T.Translate(c, errID))
}
