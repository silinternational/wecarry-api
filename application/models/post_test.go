package models

import (
	"context"
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
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			post: Post{
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "created_by",
		},
		{
			name: "missing type",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "type",
		},
		{
			name: "missing organization_id",
			post: Post{
				CreatedByID: 1,
				Type:        PostTypeRequest,
				Title:       "A Request",
				Size:        PostSizeMedium,
				Status:      PostStatusOpen,
				Uuid:        domain.GetUuid(),
			},
			wantErr:  true,
			errField: "organization_id",
		},
		{
			name: "missing title",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "missing size",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "size",
		},
		{
			name: "missing status",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "missing uuid",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
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
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusOpen,
				Uuid:           domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "bad status - accepted",
			post: Post{
				CreatedByID:    1,
				Type:           PostTypeRequest,
				OrganizationID: 1,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusAccepted,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - committed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusCommitted,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - delivered",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusDelivered,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - received",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusReceived,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - completed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusCompleted,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
		{
			name: "bad status - removed",
			post: Post{
				CreatedByID:    1,
				OrganizationID: 1,
				Type:           PostTypeRequest,
				Title:          "A Request",
				Size:           PostSizeMedium,
				Status:         PostStatusRemoved,
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "create_status",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.post.ValidateCreate(DB)
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

func (ms *ModelSuite) TestPost_ValidateUpdate() {
	t := ms.T()

	posts := CreateFixturesValidateUpdate(ms, t)

	tests := []struct {
		name     string
		post     Post
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "good status - from open to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[0].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to committed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCommitted,
				Uuid:   posts[0].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from open to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[0].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from open to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from open to delivered",
			post: Post{
				Status: PostStatusDelivered,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from open to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from open to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[0].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from committed to committed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCommitted,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to delivered",
			post: Post{
				Status: PostStatusDelivered,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from committed to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[1].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from committed to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[1].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from committed to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[1].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from accepted to accepted",
			post: Post{
				Title:  "New Title",
				Status: PostStatusAccepted,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to delivered",
			post: Post{
				Status: PostStatusDelivered,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from accepted to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[2].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from accepted to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[2].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from received to received",
			post: Post{
				Title:  "New Title",
				Status: PostStatusReceived,
				Uuid:   posts[3].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[3].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[3].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from received to delivered",
			post: Post{
				Status: PostStatusDelivered,
				Uuid:   posts[3].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from received to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[3].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from received to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[3].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from received to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[3].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from completed to completed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusCompleted,
				Uuid:   posts[4].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from completed to delivered",
			post: Post{
				Status: PostStatusDelivered,
				Uuid:   posts[4].Uuid,
			},
			wantErr: false,
		},
		{
			name: "good status - from completed to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[4].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from completed to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from completed to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from completed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from completed to removed",
			post: Post{
				Status: PostStatusRemoved,
				Uuid:   posts[4].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "good status - from removed to removed",
			post: Post{
				Title:  "New Title",
				Status: PostStatusRemoved,
				Uuid:   posts[5].Uuid,
			},
			wantErr: false,
		},
		{
			name: "bad status - from removed to open",
			post: Post{
				Status: PostStatusOpen,
				Uuid:   posts[5].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from removed to committed",
			post: Post{
				Status: PostStatusCommitted,
				Uuid:   posts[5].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from removed to accepted",
			post: Post{
				Status: PostStatusAccepted,
				Uuid:   posts[5].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from removed to delivered",
			post: Post{
				Status: PostStatusDelivered,
				Uuid:   posts[5].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from removed to received",
			post: Post{
				Status: PostStatusReceived,
				Uuid:   posts[5].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "bad status - from removed to completed",
			post: Post{
				Status: PostStatusCompleted,
				Uuid:   posts[5].Uuid,
			},
			wantErr:  true,
			errField: "status",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.post.Type = PostTypeRequest // only Requests have been implemented thus far
			vErr, _ := test.post.ValidateUpdate(DB)
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

			ms.True(test.post.Uuid.Version() != 0)
			var p Post
			ms.NoError(p.FindByID(test.post.ID))
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

			ms.True(test.post.Uuid.Version() != 0)
			var p Post
			ms.NoError(p.FindByID(test.post.ID))
		})
	}
}

func (ms *ModelSuite) TestPost_FindByID() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

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
		{name: "good with two related fields",
			id:            posts[0].ID,
			eagerFields:   []string{"CreatedBy", "Provider"},
			wantPost:      posts[0],
			wantCreatedBy: users[0],
			wantProvider:  users[1],
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
					ms.Equal(test.wantProvider.ID, post.Provider.ID, "bod post provider id")
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_FindByUUID() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name    string
		uuid    string
		want    Post
		wantErr bool
	}{
		{name: "good", uuid: posts[0].Uuid.String(), want: posts[0]},
		{name: "blank uuid", uuid: "", wantErr: true},
		{name: "wrong uuid", uuid: domain.GetUuid().String(), wantErr: true},
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
				} else if post.Uuid != test.want.Uuid {
					t.Errorf("FindByUUID() got = %s, want %s", post.Uuid, test.want.Uuid)
				}
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetCreator() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].CreatedBy.Uuid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetCreator()
			if err != nil {
				t.Errorf("GetCreator() error = %v", err)
			} else if user.Uuid != test.want {
				t.Errorf("GetCreator() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetProvider() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want *uuid.UUID
	}{
		{name: "good", post: posts[0], want: &posts[0].Provider.Uuid},
		{name: "nil", post: posts[1], want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetProvider()
			if err != nil {
				t.Errorf("GetProvider() error = %v", err)
			} else if test.want == nil {
				if user != nil {
					t.Errorf("expected nil, got %s", user.Uuid.String())
				}
			} else if user == nil {
				t.Errorf("received nil, expected %v", test.want.String())
			} else if user.Uuid != *test.want {
				t.Errorf("GetProvider() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetReceiver() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want *uuid.UUID
	}{
		{name: "good", post: posts[1], want: &posts[1].Receiver.Uuid},
		{name: "nil", post: posts[0], want: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetReceiver()
			if err != nil {
				t.Errorf("GetReceiver() error = %v", err)
			} else if test.want == nil {
				if user != nil {
					t.Errorf("expected nil, got %s", user.Uuid.String())
				}
			} else if user == nil {
				t.Errorf("received nil, expected %v", test.want.String())
			} else if user.Uuid != *test.want {
				t.Errorf("GetProvider() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetOrganization() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].Organization.Uuid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			org, err := test.post.GetOrganization()
			if err != nil {
				t.Errorf("GetOrganization() error = %v", err)
			} else if org.Uuid != test.want {
				t.Errorf("GetOrganization() got = %s, want %s", org.Uuid, test.want)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_GetThreads() {
	t := ms.T()

	_, users, _ := CreateUserFixtures(ms, t)
	posts := CreatePostFixtures(ms, t, users)
	threadFixtures := CreateThreadFixtures(ms, posts[0])
	threads := threadFixtures.Threads

	tests := []struct {
		name string
		post Post
		want []uuid.UUID
	}{
		{name: "no threads", post: posts[1], want: []uuid.UUID{}},
		{name: "two threads", post: posts[0], want: []uuid.UUID{threads[1].Uuid, threads[0].Uuid}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.post.GetThreads(users[0])
			if err != nil {
				t.Errorf("GetThreads() error: %v", err)
			} else {
				ids := make([]uuid.UUID, len(got))
				for i := range got {
					ids[i] = got[i].Uuid
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
	if err := fileFixture.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

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

// TestPost_AttachPhoto_GetPhoto tests the AttachPhoto and GetPhoto methods of models.Post
func (ms *ModelSuite) TestPost_AttachPhoto_GetPhoto() {
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

	var photoFixture File
	const filename = "photo.gif"
	if err := photoFixture.Store(filename, []byte("GIF89a")); err != nil {
		t.Errorf("failed to create file fixture, %s", err)
	}

	attachedFile, err := post.AttachPhoto(photoFixture.UUID.String())
	if err != nil {
		t.Errorf("failed to attach photo to post, %s", err)
	} else {
		ms.Equal(filename, attachedFile.Name)
		ms.True(attachedFile.ID != 0)
		ms.True(attachedFile.UUID.Version() != 0)
	}

	if err := DB.Load(&post); err != nil {
		t.Errorf("failed to load photo relation for test post, %s", err)
	}

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
		{name: "user 1, post 0", user: f.Users[1], post: f.Posts[0]},
		{name: "user 1, post 1", user: f.Users[1], post: f.Posts[1], wantErr: "no rows in result set"},
		{name: "non-existent user", post: f.Posts[1], wantErr: "no rows in result set"},
		{name: "non-existent post", user: f.Users[1], wantErr: "no rows in result set"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			var c context.Context
			err := post.FindByUserAndUUID(c, test.user, test.post.Uuid.String())

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

	user := User{Uuid: domain.GetUuid(), Email: t.Name() + "_user@example.com", Nickname: t.Name() + "_User"}
	createFixture(ms, &user)

	organization := Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
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

func (ms *ModelSuite) TestPost_GetSetOrigin() {
	t := ms.T()

	user := User{Uuid: domain.GetUuid(), Email: t.Name() + "_user@example.com", Nickname: t.Name() + "_User"}
	createFixture(ms, &user)

	organization := Organization{Uuid: domain.GetUuid(), AuthConfig: "{}"}
	createFixture(ms, &organization)

	location := Location{}
	createFixture(ms, &location)

	post := Post{CreatedByID: user.ID, OrganizationID: organization.ID, DestinationID: location.ID}
	createFixture(ms, &post)

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
}

func (ms *ModelSuite) TestPost_NewWithUser() {
	t := ms.T()
	_, users, _ := CreateUserFixtures(ms, t)
	user := users[0]

	tests := []struct {
		name           string
		pType          PostType
		wantPostType   PostType
		wantPostStatus PostStatus
		wantReceiverID int
		wantProviderID int
		wantErr        bool
	}{
		{name: "Good Request", pType: PostTypeRequest,
			wantPostType: PostTypeRequest, wantPostStatus: PostStatusOpen,
			wantReceiverID: user.ID,
		},
		{name: "Good Offer", pType: PostTypeRequest,
			wantPostType: PostTypeRequest, wantPostStatus: PostStatusOpen,
			wantProviderID: user.ID,
		},
		{name: "Bad Type", pType: "BADTYPE", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			err := post.NewWithUser(test.pType, user)

			if test.wantErr {
				ms.Error(err)
			} else {
				ms.NoError(err)
				ms.Equal(test.wantPostType, post.Type)
				ms.Equal(user.ID, post.CreatedByID)
				ms.Equal(test.wantPostStatus, post.Status)
			}
		})
	}
}

func (ms *ModelSuite) TestPost_SetProviderWithStatus() {
	t := ms.T()
	_, users, _ := CreateUserFixtures(ms, t)
	user := users[0]

	tests := []struct {
		name           string
		status         PostStatus
		pType          PostType
		wantProviderID nulls.Int
	}{
		{name: "Committed Request", status: PostStatusCommitted,
			pType: PostTypeRequest, wantProviderID: nulls.NewInt(user.ID)},
		{name: "Not Committed Request", status: PostStatusAccepted,
			pType: PostTypeRequest, wantProviderID: nulls.Int{}},
		{name: "Committed Offer", status: PostStatusCommitted,
			pType: PostTypeOffer, wantProviderID: nulls.Int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var post Post
			post.Type = test.pType
			post.SetProviderWithStatus(test.status, user)

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
		wantPostIDs []int
		wantErr     bool
	}{
		{name: "user 0", user: f.Users[0], wantPostIDs: []int{f.Posts[2].ID, f.Posts[1].ID, f.Posts[0].ID}},
		{name: "user 1", user: f.Users[1], wantPostIDs: []int{f.Posts[2].ID, f.Posts[0].ID}},
		{name: "non-existent user", user: User{}, wantPostIDs: []int{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			posts := Posts{}
			var c context.Context
			err := posts.FindByUser(c, test.user)

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
		{status: PostStatusCommitted, want: true},
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
			name:      "Committed",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusCommitted,
			want:      true,
		},
		{
			name:      "Accepted",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusAccepted,
			want:      false,
		},
		{
			name:      "Offer Received",
			post:      Post{Type: PostTypeOffer, CreatedByID: 1},
			newStatus: PostStatusReceived,
			want:      true,
		},
		{
			name:      "Request Received",
			post:      Post{Type: PostTypeRequest, CreatedByID: 1},
			newStatus: PostStatusReceived,
			want:      false,
		},
		{
			name:      "Offer Delivered",
			newStatus: PostStatusDelivered,
			post:      Post{Type: PostTypeOffer, CreatedByID: 1},
			want:      false,
		},
		{
			name:      "Request Delivered",
			newStatus: PostStatusDelivered,
			post:      Post{Type: PostTypeRequest, CreatedByID: 1},
			want:      true,
		},
		{
			name:      "Completed",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusCompleted,
			want:      false,
		},
		{
			name:      "Removed",
			post:      Post{CreatedByID: 1},
			newStatus: PostStatusRemoved,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.newStatus.String(), func(t *testing.T) {
			if got := tt.post.canUserChangeStatus(tt.user, tt.newStatus); got != tt.want {
				t.Errorf("isStatusEditable() = %v, want %v", got, tt.want)
			}
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

func (ms *ModelSuite) TestPost_GetLocationForNotifications() {
	t := ms.T()
	f := createFixturesForGetLocationForNotifications(ms)

	tests := []struct {
		name string
		post Post
		want string
	}{
		{
			name: "offer",
			post: f.Posts[0],
			want: f.Locations[0].Description,
		},
		{
			name: "request",
			post: f.Posts[1],
			want: f.Locations[3].Description,
		},
		{
			name: "request with no origin",
			post: f.Posts[2],
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.post.GetLocationForNotifications()
			ms.NoError(err)
			ms.Equal(tt.want, got.Description)
		})
	}
}

func (ms *ModelSuite) TestPost_createNewHistory() {
	t := ms.T()
	f := createFixturesForTestPost_createNewHistory(ms)

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
				f.PostHistories[0],
				{
					Status:     PostStatusCommitted,
					ReceiverID: f.PostHistories[0].ReceiverID,
					ProviderID: nulls.NewInt(f.Users[1].ID),
				},
			},
		},
		{
			name:   "null to open",
			post:   f.Posts[1],
			status: PostStatusOpen,
			want:   PostHistories{{Status: PostStatusOpen, ReceiverID: nulls.NewInt(f.Users[1].ID)}},
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

			err := test.post.createNewHistory()
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
				ms.Equal(test.want[i].Status, histories[i].Status, "incorrect status")
				ms.Equal(test.want[i].ReceiverID, histories[i].ReceiverID, "incorrect receiver id")

				if test.providerID > 0 {
					ms.Equal(test.want[i].ProviderID, histories[i].ProviderID, "incorrect provider id")
				}
			}
		})
	}
}
