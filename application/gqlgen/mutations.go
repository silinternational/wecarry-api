package gqlgen

import (
	"context"
	"errors"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanCreateOrganization() {
		return &models.Organization{}, errors.New("user not allowed to create organizations")
	}

	org := models.Organization{
		Name:       input.Name,
		Url:        models.ConvertStringPtrToNullsString(input.URL),
		AuthType:   input.AuthType,
		AuthConfig: input.AuthConfig,
		Uuid:       domain.GetUuid(),
	}

	err := org.Save()
	return &org, err
}

func (r *mutationResolver) UpdateOrganization(ctx context.Context, input UpdateOrganizationInput) (*models.Organization, error) {
	var org models.Organization
	err := org.FindByUUID(input.ID)
	if err != nil {
		return &models.Organization{}, err
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanEditOrganization(org.ID) {
		return &models.Organization{}, errors.New("user not allowed to edit organizations")
	}

	if input.URL != nil {
		org.Url = nulls.NewString(*input.URL)
	}

	org.Name = input.Name
	org.AuthType = input.AuthType
	org.AuthConfig = input.AuthConfig
	err = org.Save()

	return &org, err
}

func (r *mutationResolver) CreateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	var org models.Organization
	err := org.FindByUUID(input.OrganizationID)
	if err != nil {
		return nil, err
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanEditOrganization(org.ID) {
		return nil, errors.New("user not allowed to edit organizations")
	}

	err = org.AddDomain(input.Domain)
	if err != nil {
		return nil, err
	}

	return org.OrganizationDomains, nil
}

func (r *mutationResolver) RemoveOrganizationDomain(ctx context.Context, input RemoveOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	var org models.Organization
	err := org.FindByUUID(input.OrganizationID)
	if err != nil {
		return nil, err
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanEditOrganization(org.ID) {
		return nil, errors.New("user not allowed to edit organizations")
	}

	err = org.RemoveDomain(input.Domain)
	if err != nil {
		return nil, err
	}

	return org.OrganizationDomains, nil
}

// SetThreadLastViewedAt sets the last viewed time for the current user on the given thread
func (r *mutationResolver) SetThreadLastViewedAt(ctx context.Context, input SetThreadLastViewedAtInput) (*models.Thread, error) {
	var thread models.Thread
	if err := thread.FindByUUID(input.ThreadID); err != nil {
		return &thread, err
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if err := thread.UpdateLastViewedAt(cUser.ID, input.Time); err != nil {
		return &thread, err
	}

	return &thread, nil
}
