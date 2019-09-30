package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

func (r *Resolver) File() FileResolver {
	return &fileResolver{r}
}

type fileResolver struct{ *Resolver }

func (r *fileResolver) URL(ctx context.Context, obj *models.File) (string, error) {
	return "fileURL", nil
}

func (r *fileResolver) ID(ctx context.Context, obj *models.File) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}
