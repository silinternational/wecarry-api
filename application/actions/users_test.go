package actions

import (
	"fmt"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/internal/test"
)

func (as *ActionSuite) Test_convertUserToPrivateAPIType() {
	uf := test.CreateUserFixtures(as.DB, 1)
	user := uf.Users[0]
	org, err := convertOrganizationToAPIType(uf.Organization)
	as.NoError(err)

	want := api.UserPrivate{
		ID:            user.UUID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		PhotoID:       nulls.NewUUID(user.PhotoFile.UUID),
		AvatarURL:     user.AuthPhotoURL,
		Organizations: []api.Organization{org},
	}
	got, _ := convertUserToPrivateAPIType(test.Ctx(), user)
	as.Equal(want, got)
}

func (as *ActionSuite) TestUsersUpdate() {
	f := fixturesForUserQuery(as)
	users0 := f.Users[0]

	photo := test.CreateFileFixture(as.DB)

	nickname := "new nickname"
	photoID := photo.UUID.String()
	reqBody := api.UsersInput{
		Nickname: &nickname,
		PhotoID:  &photoID,
	}

	req := as.JSON("/users/me")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", users0.Nickname)
	req.Headers["content-type"] = "application/json"
	res := req.Put(reqBody)

	body := res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains := []string{
		fmt.Sprintf(`"id":"%s"`, f.Users[0].UUID),
		fmt.Sprintf(`"photo_id":"%s"`, photoID),
		fmt.Sprintf(`"nickname":"%s"`, nickname),
	}
	for _, w := range wantContains {
		as.Contains(body, w)
	}

	// test for removing photo
	reqBody = api.UsersInput{
		// remove the photo by leaving it as nil
	}
	res = req.Put(reqBody)
	body = res.Body.String()
	as.Equal(200, res.Code, "incorrect status code returned, body: %s", body)

	wantContains = []string{
		`"photo_id":null`,
		fmt.Sprintf(`"nickname":"%s"`, nickname), // nickname should remain unchanged
	}
	for _, w := range wantContains {
		as.Contains(body, w)
	}
}
