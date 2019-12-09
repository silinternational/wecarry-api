package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

// Location returns the location resolver
func (r *Resolver) Location() LocationResolver {
	return &locationResolver{r}
}

type locationResolver struct{ *Resolver }

// Latitude resolves the Latitude property of the location query
func (r *locationResolver) Latitude(ctx context.Context, obj *models.Location) (*float64, error) {
	if obj == nil || !obj.Latitude.Valid {
		return nil, nil
	}
	v := obj.Latitude.Float64
	return &v, nil
}

// Longitude resolves the Longitude property of the location query
func (r *locationResolver) Longitude(ctx context.Context, obj *models.Location) (*float64, error) {
	if obj == nil || !obj.Longitude.Valid {
		return nil, nil
	}
	v := obj.Longitude.Float64
	return &v, nil
}
