package gqlgen

import (
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

type publicProfileFixtures struct {
	models.User
}

func createFixturesForGetPublicProfile() publicProfileFixtures {
	unique := domain.GetUUID().String()
	user := models.User{
		UUID:         domain.GetUUID(),
		Nickname:     "user0" + unique,
		AuthPhotoURL: nulls.NewString("https://example.com/userphoto/1"),
	}
	return publicProfileFixtures{user}
}

func (gs *GqlgenSuite) Test_getPublicProfile() {
	t := gs.T()
	f := createFixturesForGetPublicProfile()
	tests := []struct {
		name    string
		user    *models.User
		want    PublicProfile
		wantNil bool
	}{
		{
			name: "fully-specified User",
			user: &f.User,
			want: PublicProfile{
				ID:        f.User.UUID.String(),
				Nickname:  f.User.Nickname,
				AvatarURL: &f.User.AuthPhotoURL.String,
			},
		},
		{
			name:    "nil user",
			user:    nil,
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ctx *buffalo.DefaultContext
			profile := getPublicProfile(ctx, test.user)

			if test.wantNil {
				gs.Nil(profile)
				return
			}
			gs.NotNil(profile)
			gs.Equal(test.want.ID, profile.ID, "ID doesn't match")
			gs.Equal(test.want.Nickname, profile.Nickname, "Nickname doesn't match")
			gs.Equal(*test.want.AvatarURL, *profile.AvatarURL, "AvatarURL doesn't match")
		})
	}
}
