package gqlgen

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
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
	selectFields := GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx))
	return obj.GetSender(selectFields)
}

func (r *messageResolver) Thread(ctx context.Context, obj *models.Message) (*models.Thread, error) {
	if obj == nil {
		return nil, nil
	}
	selectFields := getSelectFieldsForThreads(graphql.CollectAllFields(ctx))
	return obj.GetThread(selectFields)
}

func (r *messageResolver) CreatedAt(ctx context.Context, obj *models.Message) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *messageResolver) UpdatedAt(ctx context.Context, obj *models.Message) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *queryResolver) Message(ctx context.Context, id *string) (*models.Message, error) {
	message := models.Message{}
	messageFields := GetSelectFieldsFromRequestFields(MessageFields(), graphql.CollectAllFields(ctx))

	if err := models.DB.Select(messageFields...).Where("uuid = ?", id).First(&message); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error getting message: %v", err.Error()))
		return &models.Message{}, err
	}

	return &message, nil
}

func ConvertGqlNewMessageToDBMessage(gqlMessage NewMessage, user models.User) (models.Message, error) {

	var thread models.Thread

	threadUuid := domain.ConvertStrPtrToString(gqlMessage.ThreadID)
	if threadUuid != "" {
		var err error
		thread, err = models.FindThreadByUUID(threadUuid)
		if err != nil {
			return models.Message{}, err
		}

	} else {
		var err error
		thread, err = models.CreateThreadWithParticipants(gqlMessage.PostID, user)
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
