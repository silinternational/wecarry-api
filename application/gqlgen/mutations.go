package gqlgen

import (
	"context"
	"errors"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/nulls"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type mutationResolver struct{ *Resolver }

// CreateOrganization adds a new organization, if the current user has appropriate permissions.
func (r *mutationResolver) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}
	if !cUser.CanCreateOrganization() {
		extras["user.admin_role"] = cUser.AdminRole
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "CreateOrganization.Unauthorized", extras)
	}

	org := models.Organization{
		Name:       input.Name,
		Url:        models.ConvertStringPtrToNullsString(input.URL),
		AuthType:   input.AuthType,
		AuthConfig: input.AuthConfig,
	}

	if input.LogoFileID != nil {
		if _, err := org.AttachLogo(*input.LogoFileID); err != nil {
			return nil, reportError(ctx, err, "CreateOrganization.LogoFileNotFound")
		}
	}

	if err := org.Save(); err != nil {
		return nil, reportError(ctx, err, "CreateOrganization")
	}

	return &org, nil
}

// UpdateOrganization updates an organization, if the current user has appropriate permissions.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, input UpdateOrganizationInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var org models.Organization
	if err := org.FindByUUID(input.ID); err != nil {
		return nil, reportError(ctx, err, "UpdateOrganization.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "UpdateOrganization.Unauthorized", extras)
	}

	if input.URL != nil {
		org.Url = nulls.NewString(*input.URL)
	}

	if input.LogoFileID != nil {
		if _, err := org.AttachLogo(*input.LogoFileID); err != nil {
			return nil, reportError(ctx, err, "UpdateOrganization.LogoFileNotFound")
		}
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
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, reportError(ctx, err, "CreateOrganizationDomain.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "CreateOrganizationDomain.Unauthorized", extras)
	}

	if err := org.AddDomain(input.Domain, domain.ConvertStrPtrToString(input.AuthType), domain.ConvertStrPtrToString(input.AuthConfig)); err != nil {
		return nil, reportError(ctx, err, "CreateOrganizationDomain", extras)
	}

	domains, err2 := org.GetDomains()
	if err2 != nil {
		// don't return an error since the AddDomain operation succeeded
		_ = reportError(ctx, err2, "", extras)
	}

	return domains, nil
}

// UpdateOrganizationDomain is the resolver for the `updateOrganizationDomain` mutation
func (r *mutationResolver) UpdateOrganizationDomain(ctx context.Context, input CreateOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, reportError(ctx, err, "UpdateOrganizationDomain.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "UpdateOrganizationDomain.Unauthorized", extras)
	}

	var orgDomain models.OrganizationDomain
	if err := orgDomain.FindByDomain(input.Domain); err != nil {
		return nil, reportError(ctx, err, "UpdateOrganizationDomain.NotFound", extras)
	}

	orgDomain.AuthType = domain.ConvertStrPtrToString(input.AuthType)
	orgDomain.AuthConfig = domain.ConvertStrPtrToString(input.AuthConfig)
	if err := orgDomain.Save(); err != nil {
		return nil, reportError(ctx, err, "UpdateOrganizationDomain.SaveError", extras)
	}

	domains, err2 := org.GetDomains()
	if err2 != nil {
		// don't return an error since the operation succeeded
		_ = reportError(ctx, err2, "", extras)
	}

	return domains, nil
}

// RemoveOrganizationDomain is the resolver for the `removeOrganizationDomain` mutation
func (r *mutationResolver) RemoveOrganizationDomain(ctx context.Context, input RemoveOrganizationDomainInput) ([]models.OrganizationDomain, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var org models.Organization
	if err := org.FindByUUID(input.OrganizationID); err != nil {
		return nil, reportError(ctx, err, "RemoveOrganizationDomain.NotFound", extras)
	}

	if !cUser.CanEditOrganization(org.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "RemoveOrganizationDomain.Unauthorized", extras)
	}

	if err := org.RemoveDomain(input.Domain); err != nil {
		return nil, reportError(ctx, err, "RemoveOrganizationDomain", extras)
	}

	domains, err2 := org.GetDomains()
	if err2 != nil {
		// don't return an error since the RemoveDomain operation succeeded
		_ = reportError(ctx, err2, "", extras)
	}

	return domains, nil
}

// SetThreadLastViewedAt sets the last viewed time for the current user on the given thread
func (r *mutationResolver) SetThreadLastViewedAt(ctx context.Context, input SetThreadLastViewedAtInput) (*models.Thread, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
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

// CreateOrganizationTrust establishes a OrganizationTrust between two organizations
func (r *mutationResolver) CreateOrganizationTrust(ctx context.Context, input CreateOrganizationTrustInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var organization models.Organization
	if err := organization.FindByUUID(input.PrimaryID); err != nil {
		return nil, reportError(ctx, err, "CreateOrganizationTrust.FindPrimaryOrganization", extras)
	}

	if !cUser.CanCreateOrganizationTrust() {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "CreateOrganizationTrust.Unauthorized", extras)
	}

	if err := organization.CreateTrust(input.SecondaryID); err != nil {
		return nil, reportError(ctx, err, "CreateOrganizationTrust", extras)
	}

	return &organization, nil
}

// RemoveOrganizationTrust removes a OrganizationTrust between two organizations
func (r *mutationResolver) RemoveOrganizationTrust(ctx context.Context, input RemoveOrganizationTrustInput) (*models.Organization, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var organization models.Organization
	if err := organization.FindByUUID(input.PrimaryID); err != nil {
		return nil, reportError(ctx, err, "RemoveOrganizationTrust.FindPrimaryOrganization", extras)
	}

	if !cUser.CanRemoveOrganizationTrust(organization.ID) {
		err := errors.New("insufficient permissions")
		return nil, reportError(ctx, err, "RemoveOrganizationTrust.Unauthorized", extras)
	}

	if err := organization.RemoveTrust(input.SecondaryID); err != nil {
		return nil, reportError(ctx, err, "RemoveOrganizationTrust", extras)
	}

	return &organization, nil
}

// CreateMeetingInvites implements the `createMeetingInvites` mutation
func (r *mutationResolver) CreateMeetingInvites(ctx context.Context, input CreateMeetingInvitesInput) (
	[]models.MeetingInvite, error) {

	cUser := models.GetCurrentUserFromGqlContext(ctx)
	extras := map[string]interface{}{
		"user": cUser.UUID,
	}

	var m models.Meeting
	if err := m.FindByUUID(input.MeetingID); err != nil {
		return nil, reportError(ctx, err, "CreateMeetingInvite.FindMeeting", extras)
	}

	inv := models.MeetingInvite{
		MeetingID: m.ID,
		InviterID: cUser.ID,
	}

	badEmails := make([]string, 0)
	for _, email := range input.Emails {
		inv.Email = email
		if err := inv.Create(); err != nil {
			badEmails = append(badEmails, email)
			domain.ErrLogger.Printf("error creating meeting invite for email '%s', %s", email, err)
		}
	}
	if len(badEmails) > 0 {
		emailList := strings.Join(badEmails, ", ")
		graphql.AddError(ctx, gqlerror.Errorf("problem creating invite for %v", emailList))
	}

	invites, err := m.Invites()
	if err != nil {
		return nil, reportError(ctx, err, "CreateMeetingInvite.ListInvites", extras)
	}
	return invites, nil
}

func (r *mutationResolver) CreateMeetingParticipant(ctx context.Context, input CreateMeetingParticipantInput) (*models.MeetingParticipant, error) {
	var m models.MeetingParticipant
	return &m, nil
}

func (r *mutationResolver) RemoveMeetingInvite(ctx context.Context, input RemoveMeetingInviteInput) ([]models.MeetingInvite, error) {
	return []models.MeetingInvite{}, nil
}

func (r *mutationResolver) RemoveMeetingParticipant(ctx context.Context, input RemoveMeetingParticipantInput) ([]models.MeetingParticipant, error) {
	return []models.MeetingParticipant{}, nil
}
