package models

import (
	"bytes"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"os"
	"testing"
)

func (ms *ModelSuite) TestPostHistory_Load() {
	t := ms.T()
	f := createFixturesForTestPostHistory_Load(ms)

	tests := []struct {
		name        string
		postHistory PostHistory
		wantErr     string
		wantEmail   string
	}{
		{
			name:        "open to committed",
			postHistory: f.PostHistories[0],
			wantEmail:   f.Users[0].Email,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.postHistory.Load("Receiver")
			ms.NoError(err, "did not expect any error")

			ms.Equal(test.wantEmail, test.postHistory.Receiver.Email, "incorrect Receiver email")

		})
	}
}

func (ms *ModelSuite) TestPost_pop() {
	t := ms.T()
	f := createFixturesForTestPostHistory_pop(ms)

	tests := []struct {
		name          string
		post          Post
		newStatus     PostStatus
		currentStatus PostStatus
		providerID    int
		wantErr       string
		wantLog       string
		want          PostHistories
	}{
		{
			name:      "null to open - log error",
			post:      f.Posts[1],
			newStatus: PostStatusOpen,
			wantLog:   "None Found",
			want:      PostHistories{},
		},
		{
			name:          "from accepted back to committed",
			post:          f.Posts[0],
			newStatus:     PostStatusCommitted,
			currentStatus: PostStatusAccepted,
			want:          PostHistories{f.PostHistories[0], f.PostHistories[1]},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var buf bytes.Buffer
			domain.ErrLogger.SetOutput(&buf)

			defer func() {
				domain.Logger.SetOutput(os.Stdout)
			}()

			test.post.Status = test.newStatus
			var pHistory PostHistory

			err := pHistory.popForPost(test.post, test.currentStatus)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}

			ms.NoError(err, "did not expect any error")

			gotLog := buf.String()

			if test.wantLog == "" {
				ms.Equal("", gotLog, "unexpected logging message")
			} else {
				ms.Contains(gotLog, test.wantLog, "did not get expected logging message")
			}

			var histories PostHistories
			err = DB.Where("post_id = ?", test.post.ID).All(&histories)
			ms.NoError(err, "unexpected error fetching histories")

			ms.Equal(len(test.want), len(histories), "incorrect number of histories")

			for i := range test.want {
				ms.Equal(test.want[i].Status, histories[i].Status, "incorrect status")
				ms.Equal(test.want[i].ReceiverID, histories[i].ReceiverID, "incorrect receiver id")
			}
		})
	}
}

func (ms *ModelSuite) TestPostHistory_createForPost() {
	t := ms.T()
	f := createFixturesForTestPostHistory_createForPost(ms)

	tests := []struct {
		name       string
		post       Post
		status     PostStatus
		providerID int
		wantErr    string
		want       PostHistories
	}{
		{
			name:       "open to committed",
			post:       f.Posts[0],
			status:     PostStatusCommitted,
			providerID: f.Users[1].ID,
			want: PostHistories{
				{
					Status:     PostStatusOpen,
					PostID:     f.Posts[0].ID,
					ReceiverID: nulls.NewInt(f.Posts[0].CreatedByID),
				},
				{
					Status:     PostStatusCommitted,
					ReceiverID: nulls.NewInt(f.Users[0].ID),
					ProviderID: nulls.NewInt(f.Users[1].ID),
				},
			},
		},
		{
			name:   "null to open",
			post:   f.Posts[1],
			status: PostStatusOpen,
			want:   PostHistories{{Status: PostStatusOpen, ReceiverID: nulls.NewInt(f.Users[0].ID)}},
		},
		{
			name:       "bad provider id",
			post:       f.Posts[1],
			status:     PostStatusCommitted,
			providerID: 999999,
			wantErr:    `key constraint "post_histories_provider_id_fkey"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.post.Status = test.status

			if test.providerID > 0 {
				test.post.ProviderID = nulls.NewInt(test.providerID)
			}

			var pH PostHistory

			err := pH.createForPost(test.post)
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}

			ms.NoError(err, "did not expect any error")

			var histories PostHistories
			err = DB.Where("post_id = ?", test.post.ID).All(&histories)
			ms.NoError(err, "unexpected error fetching histories")

			ms.Equal(len(test.want), len(histories), "incorrect number of histories")

			for i := range test.want {
				ms.Equal(test.want[i].Status, histories[i].Status, "incorrect newStatus")
				ms.Equal(test.want[i].ReceiverID, histories[i].ReceiverID, "incorrect receiver id")

				if test.providerID > 0 {
					ms.Equal(test.want[i].ProviderID, histories[i].ProviderID, "incorrect provider id")
				}
			}
		})
	}
}
