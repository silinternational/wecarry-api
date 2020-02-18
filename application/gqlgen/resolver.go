//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"context"

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

func (r *queryResolver) SystemConfig(ctx context.Context) (*SystemConfig, error) {
	return nil, nil
}

func (r *Resolver) MeetingInvitation() MeetingInvitationResolver {
	return &meetingInvitationResolver{r}
}

type meetingInvitationResolver struct{ *Resolver }

func (m *meetingInvitationResolver) AvatarURL(ctx context.Context, obj *models.MeetingInvitation) (string, error) {
	if obj == nil {
		return "", nil
	}

	return obj.AvatarURL(), nil
}

func (m *meetingInvitationResolver) Meeting(ctx context.Context, obj *models.MeetingInvitation) (*models.Meeting, error) {
	if obj == nil {
		return nil, nil
	}

	mtg, err := obj.Meeting()
	return &mtg, err
}

func (m *meetingInvitationResolver) Inviter(ctx context.Context, obj *models.MeetingInvitation) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}

	inviter, err := obj.Inviter()
	if err != nil {
		return nil, reportError(ctx, err, "MeetingInvitation.GetInviter")
	}

	return getPublicProfile(ctx, &inviter), nil
}
