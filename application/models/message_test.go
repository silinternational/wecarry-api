package models

import ()

func (as *ModelSuite) TestMessage_GetSender() {
	t := as.T()

	messageFixtures := Fixtures_GetSender(t)

	messages := messageFixtures.Messages
	users := messageFixtures.Users

	userResults, err := messages[1].GetSender([]string{"nickname", "last_name"})

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	as.Equal(users[1].ID, userResults.ID, "Bad user ID")

	as.Equal(users[1].Nickname, userResults.Nickname, "Bad user Nickname")

}
