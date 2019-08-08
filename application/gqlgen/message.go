package gqlgen

import (
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

func MessageSimpleFields() map[string]string {
	return map[string]string{
		"id":        "uuid",
		"content":   "content",
		"sender":    "sent_by_id",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
	}
}

func ConvertGqlNewMessageToDBMessage(gqlMessage NewMessage, user models.User, requestFields []string) (models.Message, error) {

	var thread models.Thread

	threadUuid := domain.ConvertStrPtrToString(gqlMessage.ThreadID)
	if threadUuid != "" {
		var err error
		selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)
		thread, err = models.GetThreadAndParticipants(threadUuid, user, selectFields)
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
