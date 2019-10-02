package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

// File returns the file resolver
func (r *Resolver) File() FileResolver {
	return &fileResolver{r}
}

type fileResolver struct{ *Resolver }

func (r *fileResolver) URL(ctx context.Context, obj *models.File) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.URL.String, nil
}

// ID resolves the ID property of the file model
func (r *fileResolver) ID(ctx context.Context, obj *models.File) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}
