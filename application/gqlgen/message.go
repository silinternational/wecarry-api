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

func getThreadAndParticipants(threadUuid string, user models.User) (models.Thread, error) {

	thread, err := models.FindThreadByUUID(threadUuid)
	if err != nil {
		return models.Thread{}, err
	}

	if thread.ID == 0 {
		return thread, fmt.Errorf("could not find thread with uuid %v", threadUuid)
	}

	users, err := models.GetThreadParticipants(thread.ID)
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

//TODO Make this good and call it from below
func createThreadWithParticipants(postUuid string, user models.User) (models.Thread, error) {
	post, err := models.FindPostByUUID(postUuid)
	if err != nil {
		return models.Thread{}, err
	}

	participants := models.Users{user}

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
}

func ConvertGqlNewMessageToDBMessage(gqlMessage NewMessage, user models.User) (models.Message, error) {

	var thread models.Thread

	threadUuid := domain.ConvertStrPtrToString(gqlMessage.ThreadID)
	if threadUuid != "" {
		var err error
		thread, err = getThreadAndParticipants(threadUuid, user)
		if err != nil {
			return models.Message{}, err
		}

	} else {
		post, err := models.FindPostByUUID(gqlMessage.PostID)
		if err != nil {
			return models.Message{}, err
		}

		participants := models.Users{user}

		if post.CreatedBy.ID != 0 && post.CreatedBy.ID != user.ID {
			participants = append(participants, post.CreatedBy)
		}

		thread = models.Thread{
			PostID:       post.ID,
			Uuid:         domain.GetUuid(),
			Participants: participants,
		}

		//TODO make it actually create thread-participant records where needed

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
