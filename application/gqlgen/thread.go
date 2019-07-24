package gqlgen

import (
	"fmt"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

func ThreadRelatedFields() map[string]string {
	return map[string]string{
		//"participants": "Participants",
		"messages": "Messages",
	}
}

// ConvertDBThreadToGqlThread does what its name says, but also gets the
// thread participants
func ConvertDBThreadToGqlThread(dbThread models.Thread) (Thread, error) {
	dbID := strconv.Itoa(dbThread.ID)

	gqlThread := Thread{
		ID:        dbID,
		CreatedAt: domain.ConvertTimeToStringPtr(dbThread.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbThread.UpdatedAt),
	}

	// TODO only get participants if they are requested
	// Get thread participants and convert them to Gql users
	participants, err := models.GetThreadParticipants(dbThread.ID)
	if err != nil {
		return gqlThread, err
	}

	gqlUsers := []*User{}

	for _, p := range participants {
		participant, err := ConvertDBUserToGqlUser(p)
		if err != nil {
			return gqlThread, err
		}
		gqlUsers = append(gqlUsers, &participant)
	}

	gqlThread.Participants = gqlUsers

	messages := []*Message{}
	for _, m := range dbThread.Messages {
		message, err := ConvertDBMessageToGqlMessage(m)
		if err != nil {
			return gqlThread, err
		}
		messages = append(messages, &message)
	}

	gqlThread.Messages = messages

	post := models.Post{}
	if err := models.DB.Find(&post, dbThread.PostID); err != nil {
		return gqlThread, fmt.Errorf("error loading post %v %s", dbThread.PostID, err)
	}

	gqlPost, err := ConvertDBPostToGqlPost(post, nil)
	if err != nil {
		return gqlThread, err
	}

	gqlThread.Post = &gqlPost

	return gqlThread, err
}
