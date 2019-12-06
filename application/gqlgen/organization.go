package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

// Organization returns the organization resolver. It is required by GraphQL
func (r *Resolver) Organization() OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

// ID resolves the `ID` property of the organization query. It provides the UUID instead of the autoincrement ID.
func (r *organizationResolver) ID(ctx context.Context, obj *models.Organization) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
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
		return nil, reportError(ctx, err, "GetOrganizationDomains")
	}

	return domains, nil
}
