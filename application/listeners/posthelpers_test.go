package listeners

import (
	"bytes"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"os"
	"testing"
)

func (ms *ModelSuite) TestGetPostUsers() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_GetPostRecipients(ms, t)
	users := orgUserPostFixtures.users
	posts := orgUserPostFixtures.posts

	tests := []struct {
		name          string
		id            int
		wantRequester PostUser
		wantProvider  PostUser
		wantErr       bool
	}{
		{name: "Request by User0 with User1 as Provider",
			id: posts[0].ID,
			wantRequester: PostUser{
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
			wantProvider: PostUser{
				Nickname: users[1].Nickname,
				Email:    users[1].Email,
			},
		},
		{name: "Request by User0 with no Provider",
			id: posts[1].ID,
			wantRequester: PostUser{
				Nickname: users[0].Nickname,
				Email:    users[0].Email,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var post models.Post
			err := post.FindByID(test.id)
			ms.NoError(err, "error finding post for test")

			postUsers := GetPostUsers(post)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantRequester, postUsers.Requester)
				ms.Equal(test.wantProvider, postUsers.Provider)
			}
		})
	}
}

func (ms *ModelSuite) TestRequestStatusUpdatedNotifications() {
	t := ms.T()

	orgUserPostFixtures := CreateFixtures_RequestStatusUpdatedNotifications(ms, t)
	//users := orgUserPostFixtures.users
	posts := orgUserPostFixtures.posts

	postStatusEData := models.PostStatusEventData{
		OldStatus: models.PostStatusOpen,
		NewStatus: models.PostStatusCommitted,
		PostID:    posts[0].ID,
	}

	var buf bytes.Buffer
	domain.ErrLogger.SetOutput(&buf)

	defer func() {
		domain.ErrLogger.SetOutput(os.Stderr)
	}()

	// No logging message expected
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got := buf.String()
	ms.Equal("", got, "Got an unexpected error log entry")

	buf.Reset()

	// Logging message expected about bad transition
	postStatusEData.NewStatus = models.PostStatusAccepted
	requestStatusUpdatedNotifications(posts[0], postStatusEData)

	got = buf.String()
	ms.Contains(got, "unexpected status transition 'OPEN-ACCEPTED'", "Got an unexpected error log entry")

}
