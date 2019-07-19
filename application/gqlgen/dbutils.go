package gqlgen

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/pop"
)

// CallDBEagerWithRelatedFields calls db.Eager with the
// related database fields that should be populated in the sql select
// The keys of the relatedFields should match what gqlgen uses and
//   the values should be the match what the pop model uses for the same field
func CallDBEagerWithRelatedFields(relatedFields map[string]string, db *pop.Connection, ctx context.Context) *pop.Connection {

	rctxt := graphql.GetRequestContext(ctx)
	rawQuery := rctxt.RawQuery

	updateFields := []string{}
	for gqlField, dbField := range relatedFields {

		if strings.Contains(rawQuery, gqlField) {
			updateFields = append(updateFields, dbField)
		}
	}

	if len(updateFields) > 0 {
		db = db.Eager(updateFields...)
	}

	return db
}
