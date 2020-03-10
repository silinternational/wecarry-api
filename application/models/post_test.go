package models

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate"
)

func (ms *ModelSuite) TestPost_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		post     Post
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			post: Post{
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "created_by",
		},
		{
			name: "missing organization_id",
			post: Post{
				CreatedByID: 1,
				Title:       "A Request",
				Size:        PostSizeMedium,
				Status:      PostStatusOpen,
				UUID:        domain.GetUUID(),
			},
			wantErr:  true,
			errField: "organization_id",
		},
		{
			name: "missing title",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "missing size",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "size",
		},
		{
			name: "missing status",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "missing uuid",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
			},
			wantErr:  true,
			errField: "uuid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.post.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateCreate() {
	t := ms.T()

	tests := []struct {
		name     string
		post     Post
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "good - open",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name: "bad status - accepted",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusAccepted,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - delivered",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusDelivered,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - received",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusReceived,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - completed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusCompleted,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - removed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusRemoved,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad neededBefore (today)",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				NeededBefore:   nulls.NewTime(time.Now()),
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr:  true,
			errField: "needed_before",
		},
		{
			name: "good neededBefore (tommorrow)",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				NeededBefore:   nulls.NewTime(time.Now().Add(domain.DurationDay)),
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				UUID:           domain.GetUUID(),
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.post.ValidateCreate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if (test.wantErr == false) && (vErr.HasAny()) {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateUpdate_OpenRequest() {
	t := ms.T()

	post := CreateFixturesValidateUpdate_RequestStatus(PostStatusOpen, ms, t)

	tests := []struct {
		name    string
		post    Post
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from open to open",
			post: Post{
				Status: PostStatusOpen,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to accepted",
			post: Post{
				Title:  "New Title",
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to removed",
			post: Post{
				Status: PostStatusRemoved,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to accepted",
			post: Post{
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from open to delivered",
			post: Post{
				Status: PostStatusDelivered,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from open to received",
			post: Post{
				Status: PostStatusReceived,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from open to completed",
			post: Post{
				Status: PostStatusCompleted,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"
			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateUpdate_AcceptedRequest() {
	t := ms.T()

	post := CreateFixturesValidateUpdate_RequestStatus(PostStatusAccepted, ms, t)

	tests := []struct {
		name    string
		post    Post
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from accepted to accepted",
			post: Post{
				Title:  "New Title",
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to open",
			post: Post{
				Status: PostStatusOpen,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to delivered",
			post: Post{
				Status: PostStatusDelivered,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to received",
			post: Post{
				Status: PostStatusReceived,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to completed",
			post: Post{
				Status: PostStatusCompleted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to removed",
			post: Post{
				Status: PostStatusRemoved,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"
			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateUpdate_DeliveredRequest() {
	t := ms.T()

	post := CreateFixturesValidateUpdate_RequestStatus(PostStatusDelivered, ms, t)

	tests := []struct {
		name    string
		post    Post
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from delivered to accepted",
			post: Post{
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from delivered to completed",
			post: Post{
				Status: PostStatusCompleted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from delivered to open",
			post: Post{
				Status: PostStatusOpen,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from delivered to received",
			post: Post{
				Title:  "New Title",
				Status: PostStatusReceived,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from delivered to removed",
			post: Post{
				Status: PostStatusRemoved,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"
			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateUpdate_ReceivedRequest() {
	t := ms.T()

	post := CreateFixturesValidateUpdate_RequestStatus(PostStatusReceived, ms, t)

	tests := []struct {
		name    string
		post    Post
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from received to received",
			post: Post{
				Title:  "New Title",
				Status: PostStatusReceived,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to accepted",
			post: Post{
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to completed",
			post: Post{
				Status: PostStatusCompleted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from received to open",
			post: Post{
				Status: PostStatusOpen,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from received to removed",
			post: Post{
				Status: PostStatusRemoved,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"

			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateUpdate_CompletedRequest() {
	t := ms.T()

	post := CreateFixturesValidateUpdate_RequestStatus(PostStatusCompleted, ms, t)

	tests := []struct {
		name    string
		post    Post
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from completed to completed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCompleted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from completed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from completed to delivered",
			post: Post{
				Status: PostStatusDelivered,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "good status - from completed to received",
			post: Post{
				Status: PostStatusReceived,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from completed to open",
			post: Post{
				Status: PostStatusOpen,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from completed to removed",
			post: Post{
				Status: PostStatusRemoved,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"

			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_ValidateUpdate_RemovedRequest() {
	t := ms.T()

	post := CreateFixturesValidateUpdate_RequestStatus(PostStatusRemoved, ms, t)

	tests := []struct {
		name    string
		post    Post
		want    *validate.Errors
		wantErr bool
	}{
		{
			name: "good status - from removed to removed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusRemoved,
				UUID:   post.UUID,
			},
			wantErr: false,
		},
		{
			name: "bad status - from removed to open",
			post: Post{
				Status: PostStatusOpen,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to delivered",
			post: Post{
				Status: PostStatusDelivered,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to received",
			post: Post{
				Status: PostStatusReceived,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
		{
			name: "bad status - from removed to completed",
			post: Post{
				Status: PostStatusCompleted,
				UUID:   post.UUID,
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errField := "status"

			vErr, _ := test.post.ValidateUpdate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", errField, vErr.Errors)
				}
				return
			}

			if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_Create() {
	t := ms.T()
	f := createFixturesForTestPostCreate(ms)

	tests := []struct {
		name    string
		post    Post
		wantErr string
	}{
		{
			name:    "no uuid",
			post:    f.Posts[0],
			wantErr: "",
		},
		{
			name:    "uuid given",
			post:    f.Posts[1],
			wantErr: "",
		},
		{
			name:    "validation error",
			post:    f.Posts[2],
			wantErr: "Title can not be blank.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.post.Create()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.True(test.post.UUID.Version() != 0)
			var p Post
			ms.NoError(p.FindByID(test.post.ID))

			pHistories := PostHistories{}
			err = ms.DB.Where("post_id = ?", p.ID).All(&pHistories)
			ms.NoError(err)

			ms.Equal(1, len(pHistories), "incorrect number of PostHistories")
			ms.Equal(PostStatusOpen, pHistories[0].Status, "incorrect status on PostHistory")
		})
	}
}

func (ms *ModelSuite) TestPost_Update() {
	t := ms.T()
	f := createFixturesForTestPostUpdate(ms)

	tests := []struct {
		name    string
		post    Post
		wantErr string
	}{
		{
			name:    "good",
			post:    f.Posts[0],
			wantErr: "",
		},
		{
			name:    "validation error",
			post:    f.Posts[1],
			wantErr: "Title can not be blank.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.post.Update()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ms.True(test.post.UUID.Version() != 0)
			var p Post
			ms.NoError(p.FindByID(test.post.ID))
		})
	}
}

func (ms *ModelSuite) TestPost_manageStatusTransition_forwardProgression() {
	t := ms.T()
	f := createFixturesForTestPost_manageStatusTransition_forwardProgression(ms)

	tests := []struct {
		name            string
		post            Post
		newStatus       PostStatus
		providerID      nulls.Int
		wantCompletedOn bool
	}{
		{
			name:      "open to open - no change",
			post:      f.Posts[0],
			newStatus: PostStatusOpen,
		},
		{
			name:       "open to accepted - new history with provider",
			post:       f.Posts[0],
			newStatus:  PostStatusAccepted,
			providerID: nulls.NewInt(f.Users[1].ID),
		},
		{
			name:            "accepted to completed - CompletedOn added",
			post:            f.Posts[2],
			newStatus:       PostStatusCompleted,
			wantCompletedOn: true,
		},
		{
			name:            "delivered to completed - CompletedOn added",
			post:            f.Posts[3],
			newStatus:       PostStatusCompleted,
			wantCompletedOn: true,
		},
		{
			name:       "open to accepted - new history with provider",
			post:       f.Posts[0],
			newStatus:  PostStatusAccepted,
			providerID: nulls.NewInt(f.Users[1].ID),
		},
		{
			name:       "get error",
			post:       f.Posts[1],
			newStatus:  "BadStatus",
			providerID: f.Posts[1].ProviderID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.post.Status = test.newStatus
			test.post.ProviderID = test.providerID
			err := test.post.manageStatusTransition()
			ms.NoError(err)

			ph := PostHistory{}
			err = ph.getLastForPost(test.post)
			ms.NoError(err)

			ms.Equal(test.newStatus, ph.Status, "incorrect Status ")
			ms.Equal(test.providerID, ph.ProviderID, "incorrect ProviderID ")

			if test.wantCompletedOn {
				ms.True(test.post.CompletedOn.Valid, "expected a valid CompletedOn date")
			} else {
				ms.False(test.post.CompletedOn.Valid, "expected a null CompletedOn date")
			}
		})
	}
}

func (ms *ModelSuite) TestPost_manageStatusTransition_backwardProgression() {
	t := ms.T()
	f := createFixturesForTestPost_manageStatusTransition_backwardProgression(ms)

	tests := []struct {
		name            string
		post            Post
		newStatus       PostStatus
		providerID      nulls.Int
		wantCompletedOn bool
		wantErr         string
	}{
		{
			name:       "accepted to accepted - no change",
			post:       f.Posts[0],
			newStatus:  PostStatusAccepted,
			providerID: f.Posts[0].ProviderID,
			wantErr:    "",
		},
		{
			name:       "accepted to open",
			post:       f.Posts[1],
			newStatus:  PostStatusOpen,
			providerID: nulls.Int{},
			wantErr:    "",
		},
		{
			name:            "completed to accepted - CompletedOn Dropped",
			post:            f.Posts[2],
			newStatus:       PostStatusAccepted,
			providerID:      f.Posts[2].ProviderID,
			wantCompletedOn: false,
		},
		{
			name:            "completed to delivered - CompletedOn Dropped",
			post:            f.Posts[3],
			newStatus:       PostStatusDelivered,
			providerID:      f.Posts[3].ProviderID,
			wantCompletedOn: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.post.Status = test.newStatus
			test.post.ProviderID = test.providerID
			err := test.post.manageStatusTransition()
			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error message")
				return
			}
			ms.NoError(err)

			ph := PostHistory{}
			err = ph.getLastForPost(test.post)
			ms.NoError(err)

			ms.Equal(test.newStatus, ph.Status, "incorrect Status ")
			ms.Equal(test.providerID, ph.ProviderID, "incorrect ProviderID ")
			ms.Equal(test.wantCompletedOn, test.post.CompletedOn.Valid, "incorrect CompletedOn valuie")
		})
	}
}

func (ms *ModelSuite) TestPost_FindByID() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	posts := createPostFixtures(ms.DB, 2, false)

	tests := []struct {
		name          string
		id            int
		eagerFields   []string
		wantPost      Post
		wantCreatedBy User
		wantProvider  User
		wantErr       bool
	}{
		{name: "good with no related fields",
			id:       posts[0].ID,
			wantPost: posts[0],
		},
		{name: "good with a related field",
			id:            posts[1].ID,
			eagerFields:   []string{"CreatedBy"},
			wantPost:      posts[1],
			wantCreatedBy: users[0],
		},
		{name: "zero ID", id: 0, wantErr: true},
		{name: "wrong id", id: 99999, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			err := post.FindByID(test.id, test.eagerFields...)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantPost.ID, post.ID, "bad post id")
				if test.wantCreatedBy.ID != 0 {
					ms.Equal(test.wantCreatedBy.ID, post.CreatedBy.ID, "bad post createdby id")
				}
				if test.wantProvider.ID != 0 {
					ms.Equal(test.wantProvider.ID, post.Provider.ID, "bad post provider id")
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_FindByUUID() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)
	posts := createPostFixtures(ms.DB, 1, false)

	tests := []struct {
		name    string
		uuid    string
		want    Post
		wantErr bool
	}{
		{name: "good", uuid: posts[0].UUID.String(), want: posts[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUUID().String(), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			err := post.FindByUUID(test.uuid)
			if test.wantErr {
				if (err != nil) != test.wantErr {
					t.Errorf("FindByUUID() did not return expected error")
				}
			} else {
				if err != nil {
					t.Errorf("FindByUUID() error = %v", err)
				} else if post.UUID != test.want.UUID {
					t.Errorf("FindByUUID() got = %s, want %s", post.UUID, test.want.UUID)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetCreator() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 2)
	posts := createPostFixtures(ms.DB, 1, false)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: uf.Users[0].UUID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetCreator()
			if err != nil {
				t.Errorf("GetCreator() error = %v", err)
			} else if user.UUID != test.want {
				t.Errorf("GetCreator() got = %s, want %s", user.UUID, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetProvider() {
	t := ms.T()

	uf := createUserFixtures(ms.DB, 2)
	posts := createPostFixtures(ms.DB, 2, false)
	posts[1].ProviderID = nulls.NewInt(uf.Users[1].ID)

	tests := []struct {
		name string
		post Post
		want *uuid.UUID
	}{
		{name: "good", post: posts[1], want: &uf.Users[1].UUID},
		{name: "nil", post: posts[0], want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetProvider()
			if err != nil {
				t.Errorf("GetProvider() error = %v", err)
			} else if test.want == nil {
				if user != nil {
					t.Errorf("expected nil, got %s", user.UUID.String())
				}
			} else if user == nil {
				t.Errorf("received nil, expected %v", test.want.String())
			} else if user.UUID != *test.want {
				t.Errorf("GetProvider() got = %s, want %s", user.UUID, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetOrganization() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)
	posts := createPostFixtures(ms.DB, 1, false)
	ms.NoError(ms.DB.Load(&posts, "Organization"))

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].Organization.UUID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			org, err := test.post.GetOrganization()
			if err != nil {
				t.Errorf("GetOrganization() error = %v", err)
			} else if org.UUID != test.want {
				t.Errorf("GetOrganization() got = %s, want %s", org.UUID, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetThreads() {
	t := ms.T()

	users := createUserFixtures(ms.DB, 2).Users
	posts := createPostFixtures(ms.DB, 2, false)
	threadFixtures := CreateThreadFixtures(ms, posts[0])
	threads := threadFixtures.Threads

	tests := []struct {
		name string
		post Post
		want []uuid.UUID
	}{
		{name: "no threads", post: posts[1], want: []uuid.UUID{}},
		{name: "two threads", post: posts[0], want: []uuid.UUID{threads[1].UUID, threads[0].UUID}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.post.GetThreads(users[0])
			if err != nil {
				t.Errorf("GetThreads() error: %v", err)
			} else {
				ids := make([]uuid.UUID, len(got))
				for i := range got {
					ids[i] = got[i].UUID
				}
				if !reflect.DeepEqual(ids, test.want) {
					t.Errorf("GetThreads() got = %s, want %s", ids, test.want)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_AttachFile() {
	t := ms.T()

	user := User{}
	createFixture(ms, &user)

	organization := Organization{AuthConfig: "{}"}
	createFixture(ms, &organization)

	location := Location{}
	createFixture(ms, &location)

	post := Post{
		CreatedByID:    user.ID,
		OrganizationID: organization.ID,
		DestinationID:  location.ID,
	}
	createFixture(ms, &post)

	var fileFixture File
	const filename = "photo.gif"
	ms.Nil(fileFixture.Store(filename, []byte("GIF89a")), "failed to create file fixture")

	if attachedFile, err := post.AttachFile(fileFixture.UUID.String()); err != nil {
		t.Errorf("failed to attach file to post, %s", err)
	} else {
		ms.Equal(filename, attachedFile.Name)
		ms.True(attachedFile.ID != 0)
		ms.True(attachedFile.UUID.Version() != 0)
	}

	if err := ms.DB.Load(&post); err != nil {
		t.Errorf("failed to load relations for test post, %s", err)
	}

	ms.Equal(1, len(post.Files))

	if err := ms.DB.Load(&(post.Files[0])); err != nil {
		t.Errorf("failed to load files relations for test post, %s", err)
	}

	ms.Equal(filename, post.Files[0].File.Name)
}

func (ms *ModelSuite) TestPost_GetFiles() {
	f := CreateFixturesForPostsGetFiles(ms)

	files, err := f.Posts[0].GetFiles()
	ms.NoError(err, "failed to get files list for post, %s", err)

	ms.Equal(len(f.Files), len(files))

	// sort most recently updated first
	expectedFilenames := []string{
		f.Files[2].Name,
		f.Files[1].Name,
		f.Files[0].Name,
	}

	receivedFilenames := make([]string, len(files))
	for i := range files {
		receivedFilenames[i] = files[i].Name
	}

	ms.Equal(expectedFilenames, receivedFilenames, "incorrect list of files")
}

func (ms *ModelSuite) TestPost_AttachPhoto() {
	posts := createPostFixtures(ms.DB, 3, false)

	files := createFileFixtures(3)
	posts[1].PhotoFileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&posts[1], "photo_file_id"))

	tests := []struct {
		name     string
		post     Post
		oldImage *File
		newImage string
		want     File
		wantErr  string
	}{
		{
			name:     "no previous file",
			post:     posts[0],
			newImage: files[1].UUID.String(),
			want:     files[1],
		},
		{
			name:     "previous file",
			post:     posts[1],
			oldImage: &files[0],
			newImage: files[2].UUID.String(),
			want:     files[2],
		},
		{
			name:     "bad ID",
			post:     posts[2],
			newImage: uuid.UUID{}.String(),
			wantErr:  "no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.post.AttachPhoto(tt.newImage)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(tt.want.UUID.String(), got.UUID.String(), "wrong file returned")
			ms.Equal(true, got.Linked, "new post photo file is not marked as linked")
			if tt.oldImage != nil {
				ms.Equal(false, tt.oldImage.Linked, "old post photo file is not marked as unlinked")
			}
		})
	}
}

func (ms *ModelSuite) TestPost_RemovePhoto() {
	posts := createPostFixtures(ms.DB, 3, false)

	files := createFileFixtures(3)
	posts[1].PhotoFileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&posts[1], "photo_file_id"))

	tests := []struct {
		name     string
		post     Post
		oldImage *File
		wantErr  string
	}{
		{
			name: "no previous file",
			post: posts[0],
		},
		{
			name:     "previous file",
			post:     posts[1],
			oldImage: &files[0],
		},
		{
			name:    "bad ID",
			post:    Post{},
			wantErr: "invalid Post ID",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			err := tt.post.RemovePhoto()
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			if tt.oldImage != nil {
				ms.Equal(false, tt.oldImage.Linked, "old post photo file is not marked as unlinked")
			}
		})
	}
}

// TestPost_GetPhoto tests the GetPhoto method of models.Post
func (ms *ModelSuite) TestPost_GetPhotoID() {
	posts := createPostFixtures(ms.DB, 1, false)
	post := posts[0]

	var photoFixture File
	const filename = "photo.gif"
	ms.Nil(photoFixture.Store(filename, []byte("GIF89a")), "failed to create file fixture")

	attachedFile, err := post.AttachPhoto(photoFixture.UUID.String())
	ms.NoError(err, "failed to attach photo to post")
	ms.Equal(filename, attachedFile.Name)
	ms.True(attachedFile.ID != 0)
	ms.True(attachedFile.UUID.Version() != 0)

	ms.NoError(DB.Load(&post), "failed to load photo relation for test post")

	ms.Equal(filename, post.PhotoFile.Name)

	got, err := post.GetPhotoID()
	ms.NoError(err, "unexpected error")
	attachedFileUUID := attachedFile.UUID.String()
	ms.Equal(&attachedFileUUID, got)
}

// TestPost_GetPhoto tests the GetPhoto method of models.Post
func (ms *ModelSuite) TestPost_GetPhoto() {
	posts := createPostFixtures(ms.DB, 1, false)
	post := posts[0]

	var photoFixture File
	const filename = "photo.gif"
	ms.Nil(photoFixture.Store(filename, []byte("GIF89a")), "failed to create file fixture")

	attachedFile, err := post.AttachPhoto(photoFixture.UUID.String())
	ms.NoError(err, "failed to attach photo to post")
	ms.Equal(filename, attachedFile.Name)
	ms.True(attachedFile.ID != 0)
	ms.True(attachedFile.UUID.Version() != 0)

	ms.NoError(DB.Load(&post), "failed to load photo relation for test post")

	ms.Equal(filename, post.PhotoFile.Name)

	if got, err := post.GetPhoto(); err == nil {
		ms.Equal(attachedFile.UUID.String(), got.UUID.String())
		ms.True(got.URLExpiration.After(time.Now().Add(time.Minute)))
		ms.Equal(filename, got.Name)
	} else {
		ms.Fail("post.GetPhoto failed, %s", err)
	}
}

func (ms *ModelSuite) TestPost_FindByUserAndUUID() {
	t := ms.T()
	f := createFixturesForPostFindByUserAndUUID(ms)

	tests := []struct {
		name    string
		user    User
		post    Post
		wantErr string
	}{
		{name: "user 0, post 0", user: f.Users[0], post: f.Posts[0]},
		{name: "user 0, post 1", user: f.Users[0], post: f.Posts[1]},
		{name: "user 0, post 2 Removed", user: f.Users[0], post: f.Posts[2], wantErr: "no rows in result set"},
		{name: "user 1, post 0", user: f.Users[1], post: f.Posts[0]},
		{name: "user 1, post 1", user: f.Users[1], post: f.Posts[1], wantErr: "no rows in result set"},
		{name: "non-existent user", post: f.Posts[1], wantErr: "no rows in result set"},
		{name: "non-existent post", user: f.Users[1], wantErr: "no rows in result set"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			var c context.Context
			err := post.FindByUserAndUUID(c, test.user, test.post.UUID.String())

			if test.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), test.wantErr, "unexpected error")
				return
			}

			ms.NoError(err)
			ms.Equal(test.post.ID, post.ID)
		})
	}
}

func (ms *ModelSuite) TestPost_GetSetDestination() {
	t := ms.T()

	user := User{UUID: domain.GetUUID(), Email: t.Name() + "_user@example.com", Nickname: t.Name() + "_User"}
	createFixture(ms, &user)

	organization := Organization{UUID: domain.GetUUID(), AuthConfig: "{}"}
	createFixture(ms, &organization)

	locations := Locations{
		{
			Description: "a place",
			Country:     "XY",
			Latitude:    nulls.NewFloat64(1.1),
			Longitude:   nulls.NewFloat64(2.2),
		},
		{
			Description: "another place",
			Country:     "AB",
			Latitude:    nulls.Float64{},
			Longitude:   nulls.Float64{},
		},
	}
	createFixture(ms, &locations[0]) // only save the first record for now

	post := Post{CreatedByID: user.ID, OrganizationID: organization.ID, DestinationID: locations[0].ID}
	createFixture(ms, &post)

	err := post.SetDestination(locations[1])
	ms.NoError(err, "unexpected error from post.SetDestination()")

	locationFromDB, err := post.GetDestination()
	ms.NoError(err, "unexpected error from post.GetDestination()")
	locations[1].ID = locationFromDB.ID
	ms.Equal(locations[1], *locationFromDB, "destination data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)
}

func (ms *ModelSuite) TestPost_Origin() {
	posts := createPostFixtures(ms.DB, 1, false)
	post := posts[0]
	post.OriginID = nulls.Int{}
	ms.NoError(ms.DB.Save(&post))

	locationFixtures := Locations{
		{
			Description: "a place",
			Country:     "XY",
			Latitude:    nulls.NewFloat64(1.1),
			Longitude:   nulls.NewFloat64(2.2),
		},
		{
			Description: "another place",
			Country:     "AB",
			Latitude:    nulls.Float64{},
			Longitude:   nulls.Float64{},
		},
	}

	err := post.SetOrigin(locationFixtures[0])
	ms.NoError(err, "unexpected error from post.SetOrigin()")

	locationFromDB, err := post.GetOrigin()
	ms.NoError(err, "unexpected error from post.GetOrigin()")

	locationFixtures[0].ID = locationFromDB.ID
	ms.Equal(locationFixtures[0], *locationFromDB, "origin data doesn't match new location")

	err = post.SetOrigin(locationFixtures[1])
	ms.NoError(err, "unexpected error from post.SetOrigin()")

	locationFromDB, err = post.GetOrigin()
	ms.NoError(err, "unexpected error from post.GetOrigin()")
	ms.Equal(locationFixtures[0].ID, locationFromDB.ID,
		"Location ID doesn't match -- location record was probably not reused")

	locationFixtures[1].ID = locationFromDB.ID
	ms.Equal(locationFixtures[1], *locationFromDB, "origin data doesn't match after update")

	// These are redundant checks, but here to document the fact that a null overwrites previous data.
	ms.False(locationFromDB.Latitude.Valid)
	ms.False(locationFromDB.Longitude.Valid)

	ms.NoError(post.RemoveOrigin())
	ms.False(post.OriginID.Valid, "expected the origin to have been removed")
	err = ms.DB.Find(locationFromDB, locationFromDB.ID)
	ms.Error(err, "expected error when looking for removed origin")
	ms.False(domain.IsOtherThanNoRows(err), "unexpected error type finding old origin, "+err.Error())
}

func (ms *ModelSuite) TestPost_NewWithUser() {
	t := ms.T()
	user := createUserFixtures(ms.DB, 1).Users[0]

	tests := []struct {
		name           string
		wantPostStatus PostStatus
		wantProviderID int
		wantErr        bool
	}{
		{name: "Good Request", wantPostStatus: PostStatusOpen},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			err := post.NewWithUser(user)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(user.ID, post.CreatedByID)
				ms.Equal(test.wantPostStatus, post.Status)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_SetProviderWithStatus() {
	t := ms.T()
	user := createUserFixtures(ms.DB, 1).Users[0]

	tests := []struct {
		name           string
		status         PostStatus
		wantProviderID nulls.Int
	}{
		{name: "Accepted Request", status: PostStatusAccepted, wantProviderID: nulls.NewInt(user.ID)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			userID := user.UUID.String()
			err := post.SetProviderWithStatus(test.status, &userID)
			ms.NoError(err)

			ms.Equal(test.wantProviderID, post.ProviderID)
			ms.Equal(test.status, post.Status)
		})
	}
}

func (ms *ModelSuite) TestPosts_FindByUser() {
	t := ms.T()

	f := CreateFixtures_Posts_FindByUser(ms)

	tests := []struct {
		name        string
		user        User
		dest        *Location
		orig        *Location
		wantPostIDs []int
		wantErr     bool
	}{
		{name: "user 0", user: f.Users[0],
			wantPostIDs: []int{f.Posts[6].ID, f.Posts[5].ID, f.Posts[4].ID, f.Posts[1].ID, f.Posts[0].ID}},
		{name: "user 1", user: f.Users[1], wantPostIDs: []int{f.Posts[5].ID, f.Posts[4].ID, f.Posts[0].ID}},
		{name: "user 2", user: f.Users[2], wantPostIDs: []int{f.Posts[7].ID, f.Posts[6].ID, f.Posts[5].ID}},
		{name: "user 3", user: f.Users[3], wantPostIDs: []int{f.Posts[6].ID, f.Posts[5].ID, f.Posts[1].ID}},
		{name: "non-existent user", user: User{}, wantErr: true},
		{name: "destination", user: f.Users[0], dest: &Location{Country: "AU"}, wantPostIDs: []int{f.Posts[0].ID}},
		{name: "origin", user: f.Users[0], orig: &Location{Country: "AU"}, wantPostIDs: []int{f.Posts[1].ID}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			posts := Posts{}
			var c context.Context
			err := posts.FindByUser(c, test.user, test.dest, test.orig, nil)

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			postIDs := make([]int, len(posts))
			for i := range posts {
				postIDs[i] = posts[i].ID
			}
			ms.Equal(test.wantPostIDs, postIDs)
		})
	}
}

func (ms *ModelSuite) TestPosts_GetPotentialProviders() {
	t := ms.T()

	f := createPotentialProvidersFixtures(ms)
	posts := f.Posts
	pps := f.PotentialProviders

	tests := []struct {
		name      string
		post      Post
		wantPPIDs []int
	}{
		{name: "pps for first post", post: posts[0], wantPPIDs: []int{pps[0].UserID, pps[1].UserID, pps[2].UserID}},
		{name: "no pps for third post", post: posts[2], wantPPIDs: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			post := test.post
			pps, err := post.GetPotentialProviders()
			ms.NoError(err, "unexpected error")

			ids := make([]int, len(pps))
			for i, pp := range pps {
				ids[i] = pp.ID
			}
			ms.Equal(test.wantPPIDs, ids)
		})
	}
}

func (ms *ModelSuite) TestPosts_FindByUser_SearchText() {
	t := ms.T()
	f := createFixtures_Posts_FindByUser_SearchText(ms)

	tests := []struct {
		name        string
		user        User
		matchText   string
		wantPostIDs []int
		wantErr     bool
	}{
		{name: "user 0 matching case request", user: f.Users[0], matchText: "Match",
			wantPostIDs: []int{f.Posts[5].ID, f.Posts[1].ID, f.Posts[0].ID}},
		{name: "user 0 lower case request", user: f.Users[0], matchText: "match",
			wantPostIDs: []int{f.Posts[5].ID, f.Posts[1].ID, f.Posts[0].ID}},
		{name: "user 1", user: f.Users[1], matchText: "Match",
			wantPostIDs: []int{f.Posts[5].ID, f.Posts[1].ID}},
		{name: "non-existent user", user: User{}, matchText: "Match",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			posts := Posts{}
			var c context.Context
			err := posts.FindByUser(c, test.user, nil, nil, &test.matchText)

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			postIDs := make([]int, len(posts))
			for i := range posts {
				postIDs[i] = posts[i].ID
			}
			ms.Equal(test.wantPostIDs, postIDs)
		})
	}
}

func (ms *ModelSuite) TestPost_IsEditable() {
	t := ms.T()

	f := CreateFixtures_Post_IsEditable(ms)

	tests := []struct {
		name    string
		user    User
		post    Post
		want    bool
		wantErr bool
	}{
		{name: "user 0, post 0", user: f.Users[0], post: f.Posts[0], want: true},
		{name: "user 0, post 1", user: f.Users[0], post: f.Posts[1], want: false},
		{name: "user 1, post 0", user: f.Users[1], post: f.Posts[0], want: false},
		{name: "user 1, post 1", user: f.Users[1], post: f.Posts[1], want: false},
		{name: "non-existent user", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			editable, err := test.post.IsEditable(test.user)

			if test.wantErr {
				ms.Error(err)
				return
			}

			ms.NoError(err)
			ms.Equal(test.want, editable)
		})
	}
}

func (ms *ModelSuite) TestPost_isPostEditable() {
	t := ms.T()

	tests := []struct {
		status PostStatus
		want   bool
	}{
		{status: PostStatusOpen, want: true},
		{status: PostStatusAccepted, want: true},
		{status: PostStatusReceived, want: true},
		{status: PostStatusDelivered, want: true},
		{status: PostStatusCompleted, want: false},
		{status: PostStatusRemoved, want: false},
		{status: PostStatus(""), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			p := Post{Status: tt.status}
			if got := p.isPostEditable(); got != tt.want {
				t.Errorf("isStatusEditable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_canUserChangeStatus() {
	t := ms.T()

	tests := []struct {
		name      string
		post      Post
		user      User
		newStatus PostStatus
		want      bool
	}{
		{
			name: "Creator",
			post: Post{CreatedByID: 1},
			user: User{ID: 1},
			want: true,
		},
		{
			name: "SuperAdmin",
			post: Post{},
			user: User{AdminRole: UserAdminRoleSuperAdmin},
			want: true,
		},
		{
			name:      "Open",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusOpen,
			want:      false,
		},
		{
			name:      "Open to Accepted",
			post:      Post{CreatedByID: 1, Status: PostStatusOpen},
			user:      User{ID: 1},
			newStatus: PostStatusAccepted,
			want:      true,
		},
		{
			name:      "Accepted",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusAccepted,
			want:      false,
		},
		{
			name:      "Request Received",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusReceived,
			want:      false,
		},
		{
			name:      "Request From Delivered to Accepted By Requester",
			newStatus: PostStatusAccepted,
			post:      Post{CreatedByID: 1, ProviderID: nulls.NewInt(2), Status: PostStatusDelivered},
			user:      User{ID: 1},
			want:      false,
		},
		{
			name:      "Request From Delivered to Accepted By Provider",
			newStatus: PostStatusAccepted,
			post:      Post{CreatedByID: 1, ProviderID: nulls.NewInt(2), Status: PostStatusDelivered},
			user:      User{ID: 2},
			want:      true,
		},
		{
			name:      "Request From Delivered to Accepted By non-Provider",
			newStatus: PostStatusAccepted,
			post:      Post{CreatedByID: 1, ProviderID: nulls.NewInt(2), Status: PostStatusDelivered},
			user:      User{ID: 3},
			want:      false,
		},
		{
			name:      "Request Delivered By Provider",
			newStatus: PostStatusDelivered,
			post:      Post{CreatedByID: 1, ProviderID: nulls.NewInt(2)},
			user:      User{ID: 2},
			want:      true,
		},
		{
			name:      "Request Delivered By non-Provider",
			newStatus: PostStatusDelivered,
			post:      Post{CreatedByID: 1, ProviderID: nulls.NewInt(2)},
			user:      User{ID: 3},
			want:      false,
		},
		{
			name:      "Completed",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusCompleted,
			want:      false,
		},
		{
			name:      "Completed by Requester",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusCompleted,
			user:      User{ID: 1},
			want:      true,
		},
		{
			name:      "Removed",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusRemoved,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.post.canUserChangeStatus(tt.user, tt.newStatus)
			ms.Equal(tt.want, got)
		})
	}
}

func (ms *ModelSuite) TestPost_GetAudience() {
	t := ms.T()
	f := createFixturesForPostGetAudience(ms)

	tests := []struct {
		name    string
		post    Post
		want    []int
		wantErr string
	}{
		{
			name: "basic",
			post: f.Posts[0],
			want: []int{f.Users[0].ID, f.Users[1].ID},
		},
		{
			name: "no users",
			post: f.Posts[1],
			want: []int{},
		},
		{
			name:    "invalid post",
			post:    Post{},
			want:    []int{},
			wantErr: "invalid post ID in GetAudience",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.post.GetAudience()
			if tt.wantErr != "" {
				ms.Error(err)
				ms.Contains(err.Error(), tt.wantErr)
				return
			}

			ms.NoError(err)

			ids := make([]int, len(got))
			for i := range got {
				ids[i] = got[i].ID
			}
			if !reflect.DeepEqual(ids, tt.want) {
				t.Errorf("GetAudience()\ngot = %v\nwant %v", ids, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_Meeting() {
	t := ms.T()
	posts := createPostFixtures(ms.DB, 2, false)
	meeting := Meeting{
		UUID:        domain.GetUUID(),
		Name:        "a meeting",
		CreatedByID: posts[0].CreatedByID,
		LocationID:  posts[0].DestinationID,
	}
	createFixture(ms, &meeting)
	posts[0].MeetingID = nulls.NewInt(meeting.ID)
	ms.NoError(ms.DB.Save(&posts[0]))

	tests := []struct {
		name string
		post Post
		want *uuid.UUID
	}{
		{
			name: "has meeting",
			post: posts[0],
			want: &meeting.UUID,
		},
		{
			name: "no meeting",
			post: posts[1],
			want: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.post.Meeting()
			ms.NoError(err)
			if test.want == nil {
				ms.Nil(got)
				return
			}
			ms.Equal(*test.want, got.UUID)
		})
	}
}

func (ms *ModelSuite) TestPost_DestroyPotentialProviders() {
	f := createPotentialProvidersFixtures(ms)
	posts := f.Posts
	users := f.Users
	pps := f.PotentialProviders
	t := ms.T()
	tests := []struct {
		name        string
		currentUser User
		post        Post
		status      PostStatus
		wantIDs     []int
		wantErr     string
	}{
		{
			name:        "no change: wrong status",
			currentUser: users[0],
			post:        posts[1],
			status:      PostStatusAccepted,
			wantIDs:     []int{pps[0].ID, pps[1].ID, pps[2].ID, pps[3].ID, pps[4].ID},
		},
		{
			name:        "good: Post Creator as current user",
			currentUser: users[0],
			post:        posts[0],
			status:      PostStatusCompleted,
			wantIDs:     []int{pps[3].ID, pps[4].ID},
		},
		{
			name:        "bad: current user is potential provider but not Post Creator",
			currentUser: users[2],
			post:        posts[0],
			status:      PostStatusCompleted,
			wantErr: fmt.Sprintf(`user %v has insufficient permissions to destroy PotentialProviders for Post %v`,
				users[2].ID, posts[0].ID),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.post.DestroyPotentialProviders(test.status, test.currentUser)

			if test.wantErr != "" {
				ms.Error(err, "did not get error as expected")
				ms.Equal(test.wantErr, err.Error(), "wrong error message")
				return
			}

			ms.NoError(err, "unexpected error")

			var provs PotentialProviders
			err = DB.All(&provs)
			ms.NoError(err, "error just getting PotentialProviders back out of the DB.")

			pIDs := make([]int, len(provs))
			for i, p := range provs {
				pIDs[i] = p.ID
			}

			ms.Equal(test.wantIDs, pIDs)
		})
	}
}
