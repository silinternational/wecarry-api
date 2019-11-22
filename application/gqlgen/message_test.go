package gqlgen

type messageResponse struct {
	Message struct {
		ID     string `json:"id"`
		Sender struct {
			ID       string `json:"id"`
			Email    string `json:"email"`
			Nickname string `json:"nickname"`
			Posts    []struct {
				ID string `json:"id"`
			}
			PhotoURL string `json:"photoURL"`
			Location struct {
				Description string `json:"description"`
			} `json:"location"`
		} `json:"sender"`
		Content string `json:"content"`
		Thread  struct {
			ID           string `json:"id"`
			LastViewedAt string `json:"lastViewedAt"`
			Participants []struct {
				ID       string `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"participants"`
		} `json:"thread"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"message"`
}

func (gs *GqlgenSuite) Test_MessageQuery() {
	f := createFixtures_MessageQuery(gs)
	c := getGqlClient()

	query := `{ message(id: "` + f.Messages[0].Uuid.String() + `")
		{ id, content }}`

	var resp messageResponse

	TestUser = f.Users[0]
	err := c.Post(query, &resp)
	gs.NoError(err)

	gs.Equal(f.Messages[0].Uuid.String(), resp.Message.ID)
	gs.Equal(f.Messages[0].Content, resp.Message.Content)
}
