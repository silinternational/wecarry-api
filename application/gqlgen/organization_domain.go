package gqlgen

import (
	"context"

	"github.com/silinternational/handcarry-api/models"
)

func OrganizationDomainFields() map[string]string {
	return map[string]string{
		"domain":         "domain",
		"organizationId": "organizationId",
	}
}

func (r *Resolver) OrganizationDomain() OrganizationDomainResolver {
	return &organizationDomainResolver{r}
}

type organizationDomainResolver struct{ *Resolver }

func (r *organizationDomainResolver) OrganizationID(ctx context.Context, obj *models.OrganizationDomain) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Organization.Uuid.String(), nil
}
