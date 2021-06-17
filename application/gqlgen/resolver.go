//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Resolver is required by gqlgen
type Resolver struct{}

// Mutation is required by gqlgen
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query is required by gqlgen
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *Resolver) MeetingInvite() MeetingInviteResolver {
	return &meetingInviteResolver{r}
}

type meetingInviteResolver struct{ *Resolver }

func (m *meetingInviteResolver) AvatarURL(ctx context.Context, obj *models.MeetingInvite) (string, error) {
	if obj == nil {
		return "", nil
	}

	return obj.AvatarURL(), nil
}

func (m *meetingInviteResolver) Meeting(ctx context.Context, obj *models.MeetingInvite) (*models.Meeting, error) {
	if obj == nil {
		return nil, nil
	}

	mtg, err := obj.Meeting(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "MeetingInvite.Meeting")
	}
	return &mtg, nil
}

func (m *meetingInviteResolver) Inviter(ctx context.Context, obj *models.MeetingInvite) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	inviter, err := obj.Inviter(ctx)
	if err != nil {
		return nil, domain.ReportError(ctx, err, "MeetingInvite.Inviter")
	}

	return getPublicProfile(ctx, &inviter), nil
}

func (r *Resolver) MeetingParticipant() MeetingParticipantResolver {
	return &meetingParticipantResolver{r}
}

type meetingParticipantResolver struct{ *Resolver }

func (m *meetingParticipantResolver) Meeting(ctx context.Context, obj *models.MeetingParticipant) (*models.Meeting,
	error) {

	if obj == nil {
		return nil, nil
	}

	mtg, err := obj.Meeting(ctx)
	if err != nil {
		domain.NewExtra(ctx, "meetingParticipant", *obj)
		return nil, domain.ReportError(ctx, err, "MeetingParticipant.Meeting")
	}
	return &mtg, err
}

func (m *meetingParticipantResolver) User(ctx context.Context, obj *models.MeetingParticipant) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}

	user, err := obj.User(ctx)
	if err != nil {
		domain.NewExtra(ctx, "meetingParticipant", *obj)
		return nil, domain.ReportError(ctx, err, "MeetingParticipant.User")
	}

	return &user, nil
}

func (m *meetingParticipantResolver) Invite(ctx context.Context, obj *models.MeetingParticipant) (*models.MeetingInvite,
	error) {

	if obj == nil {
		return nil, nil
	}

	inv, err := obj.Invite(ctx)
	if err != nil {
		domain.NewExtra(ctx, "meetingParticipant", *obj)
		return nil, domain.ReportError(ctx, err, "MeetingParticipant.Invite")
	}

	return inv, nil
}
