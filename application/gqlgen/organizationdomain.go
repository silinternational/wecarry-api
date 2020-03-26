package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// OrganizationDomain is required by gqlgen
func (r *Resolver) OrganizationDomain() OrganizationDomainResolver {
	return &organizationDomainResolver{r}
}

type organizationDomainResolver struct{ *Resolver }

// Organization associated with the domain
func (r *organizationDomainResolver) Organization(ctx context.Context, obj *models.OrganizationDomain) (
	*models.Organization, error) {

	if obj == nil {
		return &models.Organization{}, nil
	}

	organization, err := obj.Organization()
	if err != nil {
		return &models.Organization{}, domain.ReportError(ctx, err, "GetOrganizationDomainOrganization")
	}

	return &organization, nil
}
