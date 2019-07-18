package gqlgen

import (
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertDBMessageToGqlMessage(dbMessage models.Message) (Message, error) {
	dbID := strconv.Itoa(dbMessage.ID)

	gqlMessage := Message{
		ID:        dbID,
		Content:   dbMessage.Content,
		CreatedAt: domain.ConvertTimeToStringPtr(dbMessage.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbMessage.UpdatedAt),
	}

	return gqlMessage, nil
}

func ConvertGqlNewMessageToDBMessage(gqlMessage NewMessage) (models.Message, error) {

	thread, err := models.FindThreadByUUID(domain.ConvertStrPtrToString(gqlMessage.ThreadID))
	if err != nil {
		return models.Message{}, err
	}

	user, err := models.FindUserByUUID(gqlMessage.SenderID)
	if err != nil {
		return models.Message{}, err
	}

	dbMessage := models.Message{}
	dbMessage.Uuid = domain.GetUuid()
	dbMessage.Content = gqlMessage.Content
	dbMessage.ThreadID = thread.ID
	dbMessage.SentByID = user.ID

	return dbMessage, nil
}
