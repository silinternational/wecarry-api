package actions

import (
	"fmt"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/internal/test"
	"github.com/silinternational/wecarry-api/models"
)

func (as *ActionSuite) Test_convertUserPrivate() {
	uf := test.CreateUserFixtures(as.DB, 1)
	user := uf.Users[0]
	org := models.ConvertOrganization(uf.Organization)

	want := api.UserPrivate{
		ID:            user.UUID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AuthPhotoURL,
		Organizations: []api.Organization{org},
	}
	got, _ := models.ConvertUserPrivate(test.Ctx(), user)
	as.Equal(want, got)

	// with Photo
	photo := test.CreateFileFixture(as.DB)
	_, err := user.AttachPhoto(as.DB, photo.UUID.String())
	as.NoError(err)
	want = api.UserPrivate{
		ID:            user.UUID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		PhotoID:       nulls.NewUUID(photo.UUID),
		AvatarURL:     nulls.NewString(photo.URL),
		Organizations: []api.Organization{org},
	}
	got, _ = models.ConvertUserPrivate(test.Ctx(), user)
	as.Equal(want, got)
}

func (as *ActionSuite) Test_convertUser() {
	uf := test.CreateUserFixtures(as.DB, 1)
	user := uf.Users[0]

	want := api.User{
		ID:        user.UUID,
		Nickname:  user.Nickname,
		AvatarURL: user.AuthPhotoURL,
	}
	got, _ := models.ConvertUser(test.Ctx(), user)
	as.Equal(want, got)

	// with Photo
	photo := test.CreateFileFixture(as.DB)
	_, err := user.AttachPhoto(as.DB, photo.UUID.String())
	as.NoError(err)
	want = api.User{
		ID:        user.UUID,
		Nickname:  user.Nickname,
		AvatarURL: nulls.NewString(photo.URL),
	}
	got, _ = models.ConvertUser(test.Ctx(), user)
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

func (as *ActionSuite) verifyUser(user models.User, apiUser api.User, msg string) {
	as.Equal(user.UUID, apiUser.ID, msg+", ID is not correct")

	as.Equal(user.Nickname, apiUser.Nickname, msg+", Nickname is not correct")

	avatarURL, err := user.GetPhotoURL(as.DB)
	as.NoError(err)
	if avatarURL == nil {
		as.False(apiUser.AvatarURL.Valid, msg+", AvatarURL should be null but isn't")
	} else {
		as.Equal(*avatarURL, apiUser.AvatarURL.String, msg+", AvatarURL is not correct")
	}
}
