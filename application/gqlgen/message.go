package gqlgen

import (
	"fmt"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertSimpleDBMessageToGqlMessage(dbMessage models.Message) (Message, error) {
	dbID := strconv.Itoa(dbMessage.ID)

	gqlMessage := Message{
		ID:        dbID,
		Content:   dbMessage.Content,
		CreatedAt: domain.ConvertTimeToStringPtr(dbMessage.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbMessage.UpdatedAt),
	}

	return gqlMessage, nil
}

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertDBMessageToGqlMessage(dbMessage models.Message) (Message, error) {

	// TODO this fetching of related objects is all quick and dirty.  Rewrite when there is time.
	dbID := strconv.Itoa(dbMessage.ID)

	dbThread := models.Thread{}
	if err := models.DB.Find(&dbThread, dbMessage.ThreadID); err != nil {
		return Message{}, err
	}

	thread, err := ConvertDBThreadToGqlThread(dbThread)
	if err != nil {
		return Message{}, err
	}

	dbMessages := models.Messages{}
	queryString := fmt.Sprintf("thread_id = '%s'", thread.ID)
	if err := models.DB.Where(queryString).All(&dbMessages); err != nil {
		return Message{}, err
	}

	for _, m := range dbMessages {
		gqlMsg, err := ConvertSimpleDBMessageToGqlMessage(m)
		if err != nil {
			return Message{}, err
		}

		thread.Messages = append(thread.Messages, &gqlMsg)
	}

	dbUser := models.User{}
	if err := models.DB.Find(&dbUser, dbMessage.SentByID); err != nil {
		return Message{}, err
	}

	sender, err := ConvertDBUserToGqlUser(dbUser)
	if err != nil {
		return Message{}, err
	}

	gqlMessage := Message{
		ID:        dbID,
		Content:   dbMessage.Content,
		Thread:    &thread,
		Sender:    &sender,
		CreatedAt: domain.ConvertTimeToStringPtr(dbMessage.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbMessage.UpdatedAt),
	}

	return gqlMessage, err
}

func ConvertGqlNewMessageToDBMessage(gqlMessage NewMessage, user models.User) (models.Message, error) {

	senderID := domain.ConvertStrPtrToString(gqlMessage.SenderID)
	if senderID != "" {
		var err error
		user, err = models.FindUserByUUID(senderID)
		if err != nil {
			return models.Message{}, err
		}
	}
	var thread models.Thread

	threadUuid := domain.ConvertStrPtrToString(gqlMessage.ThreadID)
	if threadUuid != "" {
		var err error
		thread, err = models.FindThreadByUUID(threadUuid)
		if err != nil {
			return models.Message{}, err
		}
	} else {
		post, err := models.FindPostByUUID(gqlMessage.PostID)
		if err != nil {
			return models.Message{}, err
		}

		thread = models.Thread{
			PostID:       post.ID,
			Uuid:         domain.GetUuid(),
			Participants: []models.User{user},
		}

		if err = models.DB.Save(&thread); err != nil {
			err = fmt.Errorf("error saving new thread for message: %v", err.Error())
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
