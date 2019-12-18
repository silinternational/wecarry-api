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
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	// get list of orgs that cUser is allowed to see
	orgs := models.Organizations{}
	err := orgs.AllWhereUserIsOrgAdmin(cUser)
	if err != nil {
		return orgs, reportError(ctx, err, "ListOrganizations.Error", extras)
	}

	return orgs, nil
}

func (r *queryResolver) Organization(ctx context.Context, id *string) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user":    cUser.UUID,
		"orgUUID": *id,
	}

	org := &models.Organization{}
	err := org.FindByUUID(*id)
	if domain.IsOtherThanNoRows(err) {
		return org, reportError(ctx, err, "ViewOrganization.Error", extras)
	}

	if org.ID != 0 && cUser.CanViewOrganization(org.ID) {
		return org, nil
	}

	return &models.Organization{}, reportError(ctx, errors.New("user not allowed to view organization"),
		"ViewOrganization.NotFound", extras)
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

	domains, err := obj.GetDomains()
	if err != nil {
		return nil, reportError(ctx, err, "GetOrganizationDomains")
	}

	return domains, nil
}
