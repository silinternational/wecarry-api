package gqlgen

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
)

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

func (r *Resolver) Message() MessageResolver {
	return &messageResolver{r}
}

type messageResolver struct{ *Resolver }

func (r *messageResolver) ID(ctx context.Context, obj *models.Message) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *messageResolver) Sender(ctx context.Context, obj *models.Message) (*models.User, error) {
	if obj == nil {
		return nil, nil
	}
	user, err := obj.GetSender(GetSelectFieldsForUsers(ctx))
	if err != nil {
		c := models.GetBuffaloContextFromGqlContext(ctx)
		domain.Error(c, err.Error())
		return nil, errors.New(domain.T.Translate(c, "MessageGetSender"))
	}
	return user, nil
}

func (r *messageResolver) Thread(ctx context.Context, obj *models.Message) (*models.Thread, error) {
	if obj == nil {
		return nil, nil
	}
	selectFields := getSelectFieldsForThreads(ctx)
	return obj.GetThread(selectFields)
}

func (r *queryResolver) Message(ctx context.Context, id *string) (*models.Message, error) {
	if id == nil {
		return nil, nil
	}
	var message models.Message
	messageFields := GetSelectFieldsFromRequestFields(MessageFields(), graphql.CollectAllFields(ctx))

	if err := message.FindByUUID(*id, messageFields...); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error getting message: %v", err.Error()))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return &models.Message{}, err
	}

	return &message, nil
}

func ConvertGqlCreateMessageInputToDBMessage(gqlMessage CreateMessageInput, user models.User) (models.Message, error) {

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
