package gqlgen

import (
	"github.com/gobuffalo/nulls"
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
