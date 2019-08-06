package gqlgen

import (
	"fmt"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

func ThreadSimpleFields() map[string]string {
	return map[string]string{
		"id":     "uuid",
		"postID": "post_id",
	}
}

func getSelectFieldsForThreads(requestFields []string) []string {
	selectFields := GetSelectFieldsFromRequestFields(ThreadSimpleFields(), requestFields)

	// Ensure we can get participants via the thread ID
	if domain.IsStringInSlice(ParticipantsField, requestFields) {
		selectFields = append(selectFields, "id")
	}

	// Ensure we can get the post via the post ID
	if domain.IsStringInSlice(PostField, requestFields) {
		selectFields = append(selectFields, "post_id")
	}

	return selectFields
}

func addParticipantsToThread(gqlThread *Thread, dbThread models.Thread, requestFields []string) error {
	if !domain.IsStringInSlice(ParticipantsField, requestFields) {
		return nil
	}

	selectedFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)
	participants, err := models.GetThreadParticipants(dbThread.ID, selectedFields)
	if err != nil {
		return err
	}

	var users []*models.User

	for _, p := range participants {
		users = append(users, &p)
	}

	gqlThread.Participants = users
	return nil
}

func addPostToThread(gqlThread *Thread, postID int, requestFields []string) error {
	if !domain.IsStringInSlice(PostField, requestFields) && !domain.IsStringInSlice(PostIDField, requestFields) {
		return nil
	}

	if postID <= 0 {
		return fmt.Errorf("error: postID must be positive, got %v", postID)
	}

	post := models.Post{}
	if err := models.DB.Find(&post, postID); err != nil {
		return fmt.Errorf("error loading post %v %s", postID, err)
	}

	if domain.IsStringInSlice(PostField, requestFields) {
		gqlThread.Post = &post
	}

	gqlThread.PostID = post.Uuid.String()

	return nil
}

func addMessagesToThread(gqlThread *Thread, dbThread models.Thread, requestFields []string) error {
	if !domain.IsStringInSlice(MessagesField, requestFields) {
		return nil
	}

	selectFields := GetSelectFieldsFromRequestFields(MessageSimpleFields(), requestFields)

	dbMessages := []models.Message{}
	if err := models.DB.Select(selectFields...).Where("thread_id = ?", dbThread.ID).All(&dbMessages); err != nil {
		return fmt.Errorf("error getting messages for thread id %v ... %v", dbThread.ID, err)
	}

	messages := []*Message{}

	for _, m := range dbMessages {
		message, err := ConvertDBMessageToGqlMessageWithSender(m, requestFields)
		if err != nil {
			return err
		}
		messages = append(messages, &message)
	}

	gqlThread.Messages = messages
	return nil
}

// ConvertDBThreadToGqlThread does what its name says, but also gets the
// thread participants
func ConvertDBThreadToGqlThread(dbThread models.Thread, requestFields []string) (Thread, error) {
	gqlThread := Thread{
		ID:        dbThread.Uuid.String(),
		CreatedAt: domain.ConvertTimeToStringPtr(dbThread.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbThread.UpdatedAt),
	}

	if err := addParticipantsToThread(&gqlThread, dbThread, requestFields); err != nil {
		err = fmt.Errorf("error adding participants to thread id %v ... %v", dbThread.ID, err)
		return gqlThread, err
	}

	if err := addMessagesToThread(&gqlThread, dbThread, requestFields); err != nil {
		err = fmt.Errorf("error adding messages to thread id %v ... %v", dbThread.ID, err)
		return gqlThread, err
	}

	if err := addPostToThread(&gqlThread, dbThread.PostID, requestFields); err != nil {
		err = fmt.Errorf("error adding post to thread id %v ... %v", dbThread.ID, err)
		return gqlThread, err
	}

	return gqlThread, nil
}
