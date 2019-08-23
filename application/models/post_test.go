package models

import (
	"testing"

	"github.com/silinternational/handcarry-api/domain"

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

func CreatePostFixtures(t *testing.T, user User) []Post {
	if err := DB.Load(&user, "Organizations"); err != nil {
		t.Errorf("failed to load organizations on user fixture, %s", err)
	}

	// Load UserOrganization test fixtures
	posts := []Post{
		{
			CreatedByID:    user.ID,
			Type:           "Request",
			OrganizationID: user.Organizations[0].ID,
			Title:          "A Request",
			Size:           "Medium",
			Status:         "New",
			Uuid:           domain.GetUuid(),
		},
		{
			CreatedByID:    user.ID,
			Type:           "Offer",
			OrganizationID: user.Organizations[0].ID,
			Title:          "An Offer",
			Size:           "Medium",
			Status:         "New",
			Uuid:           domain.GetUuid(),
		},
	}
	for i := range posts {
		if err := DB.Create(&posts[i]); err != nil {
			t.Errorf("could not create test user org ... %v", err)
			t.FailNow()
		}
	}

	return posts
}

func TestFindPostByUUID(t *testing.T) {
	_, user, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, user)

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

	resetTables(t)
}
