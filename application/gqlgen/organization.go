package gqlgen

import (
	"context"
	"errors"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Organization returns the organization resolver. It is required by GraphQL
func (r *Resolver) Organization() OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

func (r *queryResolver) Organizations(ctx context.Context) ([]models.Organization, error) {
	cUser := models.CurrentUser(ctx)

	// get list of orgs that cUser is allowed to see
	orgs := models.Organizations{}
	if err := orgs.AllWhereUserIsOrgAdmin(ctx, cUser); err != nil {
		return orgs, domain.ReportError(ctx, err, "ListOrganizations.Error")
	}

	return orgs, nil
}

func (r *queryResolver) Organization(ctx context.Context, id *string) (*models.Organization, error) {
	cUser := models.CurrentUser(ctx)
	domain.NewExtra(ctx, "orgUUID", *id)

	org := &models.Organization{}
	if err := org.FindByUUID(models.Tx(ctx), *id); err != nil {
		return org, domain.ReportError(ctx, err, "ViewOrganization.Error")
	}

	if org.ID != 0 && cUser.CanViewOrganization(ctx, org.ID) {
		return org, nil
	}

	return &models.Organization{}, domain.ReportError(ctx, errors.New("user not allowed to view organization"),
		"ViewOrganization.NotFound")
}

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

	domains, err := obj.Domains(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetOrganizationDomains")
	}

	return domains, nil
}

// LogoURL retrieves a URL for the organization logo.
func (r *organizationResolver) LogoURL(ctx context.Context, obj *models.Organization) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	logoURL, err := obj.LogoURL(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetOrganizationLogoURL")
	}

	return logoURL, nil
}

// TrustedOrganizations lists all of the Organizations connected to this Organization by a OrganizationTrust
func (r *organizationResolver) TrustedOrganizations(ctx context.Context, obj *models.Organization) ([]models.Organization, error) {
	if obj == nil {
		return nil, nil
	}

	organizations, err := obj.TrustedOrganizations(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetOrganizationTrustedOrganizations")
	}

	return organizations, nil
}
