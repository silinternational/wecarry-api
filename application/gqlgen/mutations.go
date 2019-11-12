package gqlgen

import (
	"context"
	"errors"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type mutationResolver struct{ *Resolver }

// CreateOrganization adds a new organization, if the current user has appropriate permissions.
func (r *mutationResolver) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}
	if !cUser.CanCreateOrganization() {
		extras["user.admin_role"] = cUser.AdminRole
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "CreateOrganization.NotAllowed", extras)
	}

	org := models.Organization{
		Name:       input.Name,
		Url:        models.ConvertStringPtrToNullsString(input.URL),
		AuthType:   input.AuthType,
		AuthConfig: input.AuthConfig,
		Uuid:       domain.GetUuid(),
	}

	if err := org.Save(); err != nil {
		return nil, reportError(ctx, err, "CreateOrganization")
	}

	return &org, nil
}

// UpdateOrganization updates an organization, if the current user has appropriate permissions.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, input UpdateOrganizationInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}

	var org models.Organization
	if err := org.FindByUUID(input.ID); err != nil {
		return nil, reportError(ctx, err, "UpdateOrganization.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "UpdateOrganization.NotAllowed", extras)
	}

	if input.URL != nil {
		org.Url = nulls.NewString(*input.URL)
	}

	org.Name = input.Name
	org.AuthType = input.AuthType
	org.AuthConfig = input.AuthConfig
	if err := org.Save(); err != nil {
		return nil, reportError(ctx, err, "UpdateOrganization", extras)
	}

	return &org, nil
}

// CreateOrganizationDomain is the resolver for the `createOrganizationDomain` mutation
func (r *mutationResolver) CreateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, reportError(ctx, err, "CreateOrganizationDomain.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "CreateOrganizationDomain.NotAllowed", extras)
	}

	if err := org.AddDomain(input.Domain); err != nil {
		return nil, reportError(ctx, err, "CreateOrganizationDomain", extras)
	}

	return org.GetDomains()
}

// RemoveOrganizationDomain is the resolver for the `removeOrganizationDomain` mutation
func (r *mutationResolver) RemoveOrganizationDomain(ctx context.Context, input RemoveOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, reportError(ctx, err, "RemoveOrganizationDomain.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "RemoveOrganizationDomain.NotAllowed", extras)
	}

	if err := org.RemoveDomain(input.Domain); err != nil {
		return nil, reportError(ctx, err, "RemoveOrganizationDomain", extras)
	}

	return org.GetDomains()
}

// SetThreadLastViewedAt sets the last viewed time for the current user on the given thread
func (r *mutationResolver) SetThreadLastViewedAt(ctx context.Context, input SetThreadLastViewedAtInput) (*models.Thread, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user": cUser.Uuid,
	}

	var thread models.Thread
	if err := thread.FindByUUID(input.ThreadID); err != nil {
		return nil, reportError(ctx, err, "SetThreadLastViewedAt.NotFound", extras)
	}

	if err := thread.UpdateLastViewedAt(cUser.ID, input.Time); err != nil {
		return nil, reportError(ctx, err, "SetThreadLastViewedAt", extras)
	}

	return &thread, nil
}
