package gqlgen

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

// MessageFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func MessageFields() map[string]string {
	return map[string]string{
		"id":        "uuid",
		"content":   "content",
		"thread":    "thread_id",
		"sender":    "sent_by_id",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}
}

// Message returns the message resolver. It is required by GraphQL
func (r *Resolver) Message() MessageResolver {
	return &messageResolver{r}
}

type messageResolver struct{ *Resolver }

// ID resolves the `ID` property of the message model
func (r *messageResolver) ID(ctx context.Context, obj *models.Message) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

// Sender resolves the `sender` property of the message model
func (r *messageResolver) Sender(ctx context.Context, obj *models.Message) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	user, err := obj.GetSender(GetSelectFieldsForUsers(ctx))
	if err != nil {
		c := models.GetBuffaloContextFromGqlContext(ctx)
		extras := map[string]interface{}{
			"query": graphql.GetRequestContext(ctx).RawQuery,
		}
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "GetMessageSender"))
	}

	return user, nil
}

// Thread resolves the `thread` property of the message model
func (r *messageResolver) Thread(ctx context.Context, obj *models.Message) (*models.Thread, error) {
	if obj == nil {
		return nil, nil
	}

	thread, err := obj.GetThread(getSelectFieldsForThreads(ctx))
	if err != nil {
		c := models.GetBuffaloContextFromGqlContext(ctx)
		extras := map[string]interface{}{
			"query": graphql.GetRequestContext(ctx).RawQuery,
		}
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "GetMessageThread"))
	}

	return thread, nil
}

// Message resolves the `message` model
func (r *queryResolver) Message(ctx context.Context, id *string) (*models.Message, error) {
	if id == nil {
		return nil, nil
	}
	var message models.Message
	messageFields := GetSelectFieldsFromRequestFields(MessageFields(), graphql.CollectAllFields(ctx))

	if err := message.FindByUUID(*id, messageFields...); err != nil {
		c := models.GetBuffaloContextFromGqlContext(ctx)
		extras := map[string]interface{}{
			"query":  graphql.GetRequestContext(ctx).RawQuery,
			"fields": messageFields,
		}
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "GetMessage"))
	}

	return &message, nil
}

func convertGqlCreateMessageInputToDBMessage(gqlMessage CreateMessageInput, user models.User) (models.Message, error) {

	var thread models.Thread

	threadUuid := domain.ConvertStrPtrToString(gqlMessage.ThreadID)
	if threadUuid != "" {
		var err error
		err = thread.FindByUUID(threadUuid)
		if err != nil {
			return models.Message{}, err
		}

	} else {
		var err error
		err = thread.CreateWithParticipants(gqlMessage.PostID, user)
		if err != nil {
			return models.Message{}, err
		}
	}

	dbMessage := models.Message{}
	dbMessage.Uuid = domain.GetUuid()
	dbMessage.Content = gqlMessage.Content
	dbMessage.ThreadID = thread.ID
	dbMessage.SentByID = user.ID

	return dbMessage, nil
}

// CreateMessage is a mutation resolver for creating a new message
func (r *mutationResolver) CreateMessage(ctx context.Context, input CreateMessageInput) (*models.Message, error) {
	c := models.GetBuffaloContextFromGqlContext(ctx)
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	extras := map[string]interface{}{
		"user":  cUser,
		"query": graphql.GetRequestContext(ctx).RawQuery,
	}
	message, err := convertGqlCreateMessageInputToDBMessage(input, cUser)
	if err != nil {
		domain.Error(c, err.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "CreateMessage.ParseInput"))
	}

	if err2 := message.Create(); err2 != nil {
		domain.Error(c, err2.Error(), extras)
		return nil, errors.New(domain.T.Translate(c, "CreateMessage"))
	}

	return &message, nil
}
