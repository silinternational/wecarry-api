package gqlgen

import (
	"fmt"
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

func getThreadAndParticipants(threadUuid string, user models.User, requestFields []string) (models.Thread, error) {

	thread, err := models.FindThreadByUUID(threadUuid)
	if err != nil {
		return models.Thread{}, err
	}

	if thread.ID == 0 {
		return thread, fmt.Errorf("could not find thread with uuid %v", threadUuid)
	}

	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)

	users, err := models.GetThreadParticipants(thread.ID, selectFields)
	if err != nil {
		return models.Thread{}, err
	}

	isUserAlreadyAParticipant := false
	for _, u := range users {
		if u.ID == user.ID {
			isUserAlreadyAParticipant = true
			break
		}
	}

	if !isUserAlreadyAParticipant {
		users = append(users, user)
	}

	thread.Participants = users

	return thread, nil
}

func createThreadWithParticipants(postUuid string, user models.User) (models.Thread, error) {
	post, err := models.FindPostByUUID(postUuid)
	if err != nil {
		return models.Thread{}, err
	}

	participants := models.Users{user}

	// Ensure Post Creator is one of the participants
	if post.CreatedBy.ID != 0 && post.CreatedBy.ID != user.ID {
		participants = append(participants, post.CreatedBy)
	}

	thread := models.Thread{
		PostID:       post.ID,
		Uuid:         domain.GetUuid(),
		Participants: participants,
	}

	if err = models.DB.Save(&thread); err != nil {
		err = fmt.Errorf("error saving new thread for message: %v", err.Error())
		return models.Thread{}, err
	}

	for _, p := range participants {
		threadP := models.ThreadParticipant{
			ThreadID: thread.ID,
			UserID:   p.ID,
		}
		if err := models.DB.Save(&threadP); err != nil {
			err = fmt.Errorf("error saving new thread participant %+v for message: %v", threadP, err.Error())
			return models.Thread{}, err
		}
	}

	return thread, nil
}

func ConvertGqlNewMessageToDBMessage(gqlMessage NewMessage, user models.User, requestFields []string) (models.Message, error) {

	var thread models.Thread

	threadUuid := domain.ConvertStrPtrToString(gqlMessage.ThreadID)
	if threadUuid != "" {
		var err error
		thread, err = getThreadAndParticipants(threadUuid, user, requestFields)
		if err != nil {
			return models.Message{}, err
		}

	} else {
		var err error
		thread, err = createThreadWithParticipants(gqlMessage.PostID, user)
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
