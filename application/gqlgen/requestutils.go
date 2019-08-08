package gqlgen

import (
	"context"
	"github.com/silinternational/handcarry-api/domain"

	//"fmt"
	"github.com/99designs/gqlgen/graphql"
)

func getSubFields(reqCtx *graphql.RequestContext, parentField graphql.CollectedField) []string {
	fieldNames := []string{parentField.Alias}
	subFields := graphql.CollectFields(reqCtx, parentField.SelectionSet, []string{"BlockChoices"})

	if len(subFields) == 0 {
		return fieldNames
	}

	for _, f := range subFields {
		fieldNames = append(fieldNames, getSubFields(reqCtx, f)...)
	}

	return fieldNames
}

func GetRequestFields(ctx context.Context) []string {
	fields := graphql.CollectFieldsCtx(ctx, []string{"Block"})
	reqCtx := graphql.GetRequestContext(ctx)

	fieldNames := []string{}

	for _, f := range fields {
		fieldNames = append(fieldNames, getSubFields(reqCtx, f)...)
	}
	return fieldNames
}

// GetSelectFieldsFromRequestFields gets the intersection of the non-relational fields for the db model
//  and the top-level requested fields
func GetSelectFieldsFromRequestFields(simpleFields map[string]string, requestFields []string) []string {
	if len(requestFields) == 0 {
		return []string{}
	}

	// TODO: GetRequestFields gets *all* request fields smashed into one list. Need something that
	// gives just the request fields from the object of interest.
	selectFields := []string{}
	for gqlField, dbField := range simpleFields {
		if domain.IsStringInSlice(gqlField, requestFields) {
			selectFields = append(selectFields, dbField)
		}
	}

	return selectFields
}
