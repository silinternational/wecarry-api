package actions

import (
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
		PhotoID:       user.PhotoFile.UUID,
		AvatarURL:     user.AuthPhotoURL,
		Organizations: []api.Organization{org},
	}
	got, _ := convertUserToPrivateAPIType(test.Ctx(), user)
	as.Equal(want, got)
}
