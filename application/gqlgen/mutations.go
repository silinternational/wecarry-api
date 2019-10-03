package gqlgen

import (
	"context"
	"fmt"

	"github.com/gobuffalo/nulls"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateMessage(ctx context.Context, input CreateMessageInput) (*models.Message, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	message, err := ConvertGqlCreateMessageInputToDBMessage(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Message{}, err
	}

	if err := models.DB.Create(&message); err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Message{}, err
	}

	return &message, err
}

func (r *mutationResolver) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanCreateOrganization() {
		return &models.Organization{}, fmt.Errorf("user not allowed to create organizations")
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
		return &models.Organization{}, fmt.Errorf("user not allowed to edit organizations")
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

func (r *mutationResolver) CreateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]*models.OrganizationDomain, error) {
	var org models.Organization
	err := org.FindByUUID(input.OrganizationID)
	if err != nil {
		return []*models.OrganizationDomain{}, err
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanEditOrganization(org.ID) {
		return []*models.OrganizationDomain{}, fmt.Errorf("user not allowed to edit organizations")
	}

	err = org.AddDomain(input.Domain)
	if err != nil {
		return []*models.OrganizationDomain{}, err
	}

	orgDomains := make([]*models.OrganizationDomain, len(org.OrganizationDomains))
	for i, od := range org.OrganizationDomains {
		orgDomains[i] = &od
	}

	return orgDomains, nil
}

func (r *mutationResolver) RemoveOrganizationDomain(ctx context.Context, input RemoveOrganizationDomainInput) ([]*models.OrganizationDomain, error) {
	var org models.Organization
	err := org.FindByUUID(input.OrganizationID)
	if err != nil {
		return []*models.OrganizationDomain{}, err
	}

	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if !cUser.CanEditOrganization(org.ID) {
		return []*models.OrganizationDomain{}, fmt.Errorf("user not allowed to edit organizations")
	}

	err = org.RemoveOrganizationDomain(input.Domain)
	if err != nil {
		return []*models.OrganizationDomain{}, err
	}

	orgDomains := make([]*models.OrganizationDomain, len(org.OrganizationDomains))
	for i, od := range org.OrganizationDomains {
		orgDomains[i] = &od
	}

	return orgDomains, nil
}
