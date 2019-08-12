package gqlgen

import (
	"context"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

func OrganizationSimpleFields() map[string]string {
	return map[string]string{
		"id":         "uuid",
		"name":       "name",
		"url":        "url",
		"authType":   "auth_type",
		"authConfig": "auth_config",
		"createdAt":  "created_at",
		"updatedAt":  "updated_at",
	}
}

func (r *Resolver) Organization() OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

func (r *organizationResolver) URL(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Url), nil
}

func (r *organizationResolver) CreatedAt(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *organizationResolver) UpdatedAt(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}
