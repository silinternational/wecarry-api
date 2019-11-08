package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/models"
)

// OrganizationDomainFields maps GraphQL fields to their equivalent database fields.
func OrganizationDomainFields() map[string]string {
	return map[string]string{
		"organizationID": "organization_id",
		"domain":         "domain",
	}
}

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
	return obj.GetOrganizationUUID()
}
