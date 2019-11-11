package gqlgen

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"

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

// getFunctionName provides the filename, line number, and function name of the 2nd caller.
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
	domain.Error(c, err.Error(), allExtras)

	if domain.T == nil {
		return errors.New(errID)
	}
	return errors.New(domain.T.Translate(c, errID))
}
