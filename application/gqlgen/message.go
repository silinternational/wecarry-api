package gqlgen

import (
	"context"

	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// Message returns the message resolver. It is required by GraphQL
func (r *Resolver) Message() MessageResolver {
	return &messageResolver{r}
}

type messageResolver struct{ *Resolver }

// ID resolves the `ID` property of the message query
func (r *messageResolver) ID(ctx context.Context, obj *models.Message) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UUID.String(), nil
}

// Sender resolves the `sender` property of the message query
func (r *messageResolver) Sender(ctx context.Context, obj *models.Message) (*PublicProfile, error) {
	if obj == nil {
		return nil, nil
	}
	user, err := obj.GetSender()
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetMessageSender")
	}

	return getPublicProfile(ctx, user), nil
}

// Thread resolves the `thread` property of the message query
func (r *messageResolver) Thread(ctx context.Context, obj *models.Message) (*models.Thread, error) {
	if obj == nil {
		return nil, nil
	}

	thread, err := obj.GetThread()
	if err != nil {
		return nil, domain.ReportError(ctx, err, "GetMessageThread")
	}

	return thread, nil
}

// Message resolves the `message` model
func (r *queryResolver) Message(ctx context.Context, id *string) (*models.Message, error) {
	if id == nil {
		return nil, nil
	}
	currentUser := models.CurrentUser(ctx)
	var message models.Message

	if err := message.FindByUserAndUUID(currentUser, *id); err != nil {
		extras := map[string]interface{}{
			"user": currentUser.UUID.String(),
		}
		return nil, domain.ReportError(ctx, err, "GetMessage", extras)
	}

	return &message, nil
}

// CreateMessage is a mutation resolver for creating a new message
func (r *mutationResolver) CreateMessage(ctx context.Context, input CreateMessageInput) (*models.Message, error) {
	var message models.Message
	if err := message.Create(ctx, input.PostID, input.ThreadID, input.Content); err != nil {
		return nil, domain.ReportError(ctx, err, "CreateMessage")
	}

	return &message, nil
}
