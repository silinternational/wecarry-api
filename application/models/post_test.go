package models

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/silinternational/handcarry-api/domain"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate"
)

func TestPost_Validate(t *testing.T) {
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
				Type:           "Request",
				OrganizationID: 1,
				Title:          "A Request",
				Size:           "Medium",
				Status:         "New",
				Uuid:           domain.GetUuid(),
			},
			wantErr: false,
		},
		{
			name: "missing created_by",
			post: Post{
				Type:           "Request",
				OrganizationID: 1,
				Title:          "A Request",
				Size:           "Medium",
				Status:         "New",
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
				Size:           "Medium",
				Status:         "New",
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "type",
		},
		{
			name: "missing organization_id",
			post: Post{
				CreatedByID: 1,
				Type:        "Request",
				Title:       "A Request",
				Size:        "Medium",
				Status:      "New",
				Uuid:        domain.GetUuid(),
			},
			wantErr:  true,
			errField: "organization_id",
		},
		{
			name: "missing title",
			post: Post{
				CreatedByID:    1,
				Type:           "Request",
				OrganizationID: 1,
				Size:           "Medium",
				Status:         "New",
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "missing size",
			post: Post{
				CreatedByID:    1,
				Type:           "Request",
				OrganizationID: 1,
				Title:          "A Request",
				Status:         "New",
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "size",
		},
		{
			name: "missing status",
			post: Post{
				CreatedByID:    1,
				Type:           "Request",
				OrganizationID: 1,
				Title:          "A Request",
				Size:           "Medium",
				Uuid:           domain.GetUuid(),
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "missing uuid",
			post: Post{
				CreatedByID:    1,
				Type:           "Request",
				OrganizationID: 1,
				Title:          "A Request",
				Size:           "Medium",
				Status:         "New",
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

func CreatePostFixtures(t *testing.T, users Users) []Post {
	if err := DB.Load(&users[0], "Organizations"); err != nil {
		t.Errorf("failed to load organizations on users[0] fixture, %s", err)
	}

	// Load UserOrganization test fixtures
	posts := []Post{
		{
			CreatedByID:    users[0].ID,
			Type:           "Request",
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "A Request",
			Size:           "Medium",
			Status:         "New",
			Uuid:           domain.GetUuid(),
			ProviderID:     nulls.NewInt(users[1].ID),
		},
		{
			CreatedByID:    users[0].ID,
			Type:           "Offer",
			OrganizationID: users[0].Organizations[0].ID,
			Title:          "An Offer",
			Size:           "Medium",
			Status:         "New",
			Uuid:           domain.GetUuid(),
			ReceiverID:     nulls.NewInt(users[1].ID),
		},
	}
	for i := range posts {
		if err := DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
		if err := DB.Load(&posts[i], "CreatedBy", "Provider", "Receiver", "Organization"); err != nil {
			t.Errorf("Error loading post associations: %s", err)
			t.FailNow()
		}
	}
	return posts
}

func TestFindPostByUUID(t *testing.T) {
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

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

func TestPost_GetCreator(t *testing.T) {
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].CreatedBy.Uuid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			user, err := test.post.GetCreator([]string{"uuid"})
			if err != nil {
				t.Errorf("GetCreator() error = %v", err)
			} else if user.Uuid != test.want {
				t.Errorf("GetCreator() got = %s, want %s", user.Uuid, test.want)
			}
		})
	}
}

func TestPost_GetProvider(t *testing.T) {
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

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
			user, err := test.post.GetProvider([]string{"uuid"})
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

func TestPost_GetReceiver(t *testing.T) {
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

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
			user, err := test.post.GetReceiver([]string{"uuid"})
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

func TestPost_GetOrganization(t *testing.T) {
	resetTables(t)

	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

	tests := []struct {
		name string
		post Post
		want uuid.UUID
	}{
		{name: "good", post: posts[0], want: posts[0].Organization.Uuid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			org, err := test.post.GetOrganization([]string{"uuid"})
			if err != nil {
				t.Errorf("GetOrganization() error = %v", err)
			} else if org.Uuid != test.want {
				t.Errorf("GetOrganization() got = %s, want %s", org.Uuid, test.want)
			}
		})
	}
}
