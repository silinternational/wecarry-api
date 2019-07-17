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
