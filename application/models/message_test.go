package models

func (ms *ModelSuite) TestMessage_GetSender() {
	t := ms.T()

	ResetTables(t, ms.DB)
	messageFixtures := Fixtures_GetSender(t)

	messages := messageFixtures.Messages
	users := messageFixtures.Users

	userResults, err := messages[1].GetSender([]string{"nickname", "last_name"})

	if err != nil {
		t.Errorf("unexpected error ... %v", err)
		t.FailNow()
	}

	ms.Equal(users[1].ID, userResults.ID, "Bad user ID")

	ms.Equal(users[1].Nickname, userResults.Nickname, "Bad user Nickname")
}
