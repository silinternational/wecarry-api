package gqlgen

import (
	"context"
	"github.com/gobuffalo/pop"
	"github.com/silinternational/handcarry-api/domain"
)

// CallDBEagerWithRelatedFields calls db.Eager with the
// related database fields that should be populated in the sql select
// The keys of the relatedFields should match what gqlgen uses and
//   the values should be the match what the pop model uses for the same field
func CallDBEagerWithRelatedFields(relatedFields map[string]string, db *pop.Connection, ctx context.Context) *pop.Connection {
	requestFields := GetRequestFields(ctx)

	updateFields := []string{}
	for gqlField, dbField := range relatedFields {

		if domain.IsStringInSlice(gqlField, requestFields) {
			updateFields = append(updateFields, dbField)
		}
	}

	if len(updateFields) > 0 {
		db = db.Eager(updateFields...)
	}

	return db
}
