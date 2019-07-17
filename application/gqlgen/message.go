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
