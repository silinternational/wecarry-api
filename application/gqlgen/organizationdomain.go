package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

// OrganizationDomain is required by gqlgen
func (r *Resolver) OrganizationDomain() OrganizationDomainResolver {
	return &organizationDomainResolver{r}
}

type organizationDomainResolver struct{ *Resolver }

// OrganizationID converts the organization's autoincrement ID to its UUID
func (r *organizationDomainResolver) OrganizationID(ctx context.Context, obj *models.OrganizationDomain) (string, error) {
	if obj == nil {
		return "", nil
	}

	id, err := obj.GetOrganizationUUID()
	if err != nil {
		return "", reportError(ctx, err, "GetOrganizationDomainOrganizationID")
	}

	return id, nil
}
