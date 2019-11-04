package gqlgen

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type mutationResolver struct{ *Resolver }

// CreateOrganization adds a now organization, if the current user has appropriate permissions.
func (r *mutationResolver) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*models.Organization, error) {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user":  cUser.Uuid,
		"query": *graphql.GetRequestContext(ctx),
	}
	if !cUser.CanCreateOrganization() {
		extras["user.admin_role"] = cUser.AdminRole
		domain.Error(c, "insufficient permissions", extras)
		return nil, errors.New(domain.T.Translate(c, "CreateOrganization.NotAllowed"))
	}

	org := models.Organization{
		Name:       input.Name,
		Url:        models.ConvertStringPtrToNullsString(input.URL),
		AuthType:   input.AuthType,
		AuthConfig: input.AuthConfig,
		Uuid:       domain.GetUuid(),
	}

	if err := org.Save(); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "CreateOrganization"))
	}

	return &org, nil
}

// UpdateOrganization updates an organization, if the current user has appropriate permissions.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, input UpdateOrganizationInput) (*models.Organization, error) {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user":  cUser.Uuid,
		"query": graphql.GetRequestContext(ctx).RawQuery,
	}

	var org models.Organization
	err := org.FindByUUID(input.ID)
	if err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "UpdateOrganization.NotFound"))
	}

	if !cUser.CanEditOrganization(org.ID) {
		domain.Error(c, "insufficient permissions", extras)
		return nil, errors.New(domain.T.Translate(c, "UpdateOrganization.NotAllowed"))
	}

	if input.URL != nil {
		org.Url = nulls.NewString(*input.URL)
	}

	org.Name = input.Name
	org.AuthType = input.AuthType
	org.AuthConfig = input.AuthConfig
	if err = org.Save(); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "UpdateOrganization"))
	}

	return &org, nil
}

// CreateOrganizationDomain is the resolver for the `createOrganizationDomain` mutation
func (r *mutationResolver) CreateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user":  cUser.Uuid,
		"query": graphql.GetRequestContext(ctx).RawQuery,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "CreateOrganizationDomain.NotFound"))
	}

	if !cUser.CanEditOrganization(org.ID) {
		domain.Error(c, "insufficient permissions", extras)
		return nil, errors.New(domain.T.Translate(c, "CreateOrganizationDomain.NotAllowed"))
	}

	if err := org.AddDomain(input.Domain); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "CreateOrganizationDomain"))
	}

	return org.OrganizationDomains, nil
}

// RemoveOrganizationDomain is the resolver for the `removeOrganizationDomain` mutation
func (r *mutationResolver) RemoveOrganizationDomain(ctx context.Context, input RemoveOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user":  cUser.Uuid,
		"query": graphql.GetRequestContext(ctx).RawQuery,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "RemoveOrganizationDomain.NotFound"))
	}

	if !cUser.CanEditOrganization(org.ID) {
		domain.Error(c, "insufficient permissions", extras)
		return nil, errors.New(domain.T.Translate(c, "RemoveOrganizationDomain.NotAllowed"))
	}

	if err := org.RemoveDomain(input.Domain); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "RemoveOrganizationDomain"))
	}

	return org.OrganizationDomains, nil
}

// SetThreadLastViewedAt sets the last viewed time for the current user on the given thread
func (r *mutationResolver) SetThreadLastViewedAt(ctx context.Context, input SetThreadLastViewedAtInput) (*models.Thread, error) {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user":  cUser.Uuid,
		"query": graphql.GetRequestContext(ctx).RawQuery,
	}

	var thread models.Thread
	if err := thread.FindByUUID(input.ThreadID); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "SetThreadLastViewedAt.NotFound"))
	}

	if err := thread.UpdateLastViewedAt(cUser.ID, input.Time); err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "SetThreadLastViewedAt"))
	}

	return &thread, nil
}
