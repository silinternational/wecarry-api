package gqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/models"
)

// OrganizationFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func OrganizationFields() map[string]string {
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

// Organization returns the organization resolver. It is required by GraphQL
func (r *Resolver) Organization() OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

// ID resolves the `ID` property of the organization model. It provides the UUID instead of the autoincrement ID.
func (r *organizationResolver) ID(ctx context.Context, obj *models.Organization) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

// URL resolves the `URL` property, converting a nulls.String to a *string.
func (r *organizationResolver) URL(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return models.GetStringFromNullsString(obj.Url), nil
}

// Domains resolves the `domains` property, retrieving the list of organization_domains from the organization model
func (r *organizationResolver) Domains(ctx context.Context, obj *models.Organization) ([]models.OrganizationDomain, error) {
	if obj == nil {
		return nil, nil
	}

	domains, err := obj.GetDomains()
	if err != nil {
		return nil, reportError(ctx, err, "GetMessage")
	}

	return domains, nil
}

func getSelectFieldsForOrganizations(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(OrganizationFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	return selectFields
}
