package models

import (
	"testing"

	"github.com/silinternational/wecarry-api/domain"
)

type MessageFixtures struct {
	Users
	Messages
	Threads
}

func Fixtures_GetSender(ms *ModelSuite, t *testing.T) MessageFixtures {
	org := &Organization{AuthConfig: "{}", Uuid: domain.GetUuid()}
	createFixture(t, org)

	unique := domain.GetUuid().String()
	users := Users{
		{Email: unique + "user1@example.com", Nickname: unique + "User1", Uuid: domain.GetUuid()},
		{Email: unique + "user2@example.com", Nickname: unique + "User2", Uuid: domain.GetUuid()},
	}
	for i := range users {
		createFixture(t, &users[i])
	}

	posts := Posts{
		{Uuid: domain.GetUuid(), CreatedByID: users[0].ID, OrganizationID: org.ID},
	}
	for i := range posts {
		createFixture(t, &posts[i])
	}

	threads := Threads{
		{Uuid: domain.GetUuid(), PostID: posts[0].ID},
	}
	for i := range threads {
		createFixture(t, &threads[i])
	}

	messages := Messages{
		{
			Uuid:     domain.GetUuid(),
			ThreadID: threads[0].ID,
			SentByID: users[0].ID,
			Content:  "I can being chocolate if you bring PB",
		},
	}
	for i := range messages {
		createFixture(t, &messages[i])
	}

	return MessageFixtures{
		Users:    users,
		Messages: messages,
		Threads:  threads,
	}

}
