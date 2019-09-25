package gqlgen

import (
	"github.com/silinternational/wecarry-api/domain"
)

// GetSelectFieldsFromRequestFields gets the intersection of all fields for the db model
//  and the top-level requested fields
func GetSelectFieldsFromRequestFields(fields map[string]string, requestFields []string) []string {
	if len(requestFields) == 0 {
		return []string{}
	}

	selectFields := []string{}
	for gqlField, dbField := range fields {
		if domain.IsStringInSlice(gqlField, requestFields) {
			selectFields = append(selectFields, dbField)
		}
	}

	return selectFields
}
