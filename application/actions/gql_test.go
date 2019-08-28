package actions

import (
)



func (as *ActionSuite) TestQueryAUser() {
	t := as.T()

	queryFixtures := Fixtures_QueryAUser(t)
	userFixtures := queryFixtures.Users

	tUuid := userFixtures[1].Uuid.String()

	uq := map[string]string {
		"query": `{user(id: "` + tUuid + `") {id nickname}}`,
	}

	bearer := queryFixtures.ClientID + queryFixtures.AccessToken
	headers := map[string]string{
		"Content-Type": "application/json",
		"Authorization": "Bearer " + bearer,
	}

	hj := as.JSON("/gql")
	hj.Headers = headers
	res := hj.Post(uq)

	as.Equal(200, res.Code)


	u2Uuid := userFixtures[1].Uuid.String()
	u2Nname := userFixtures[1].Nickname 
	expectedBody := `{"data":{"user":{"id":"` + u2Uuid + `","nickname":"` + u2Nname + `"}}}`
	as.Equal(expectedBody, res.Body.String())

}

