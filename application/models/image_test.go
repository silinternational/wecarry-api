package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/silinternational/handcarry-api/domain"
)

func (ms *ModelSuite) TestImage_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		image    Image
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			image: Image{
				UUID:   domain.GetUuid(),
				PostID: 1,
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			image: Image{
				PostID: 1,
			},
			wantErr:  true,
			errField: "UUID",
		},
		{
			name: "missing post_id",
			image: Image{
				UUID: domain.GetUuid(),
			},
			wantErr:  true,
			errField: "post_id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.image.Validate(DB)
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

func (ms *ModelSuite) TestImage_Store() {
	t := ms.T()
	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)

	type args struct {
		postUUID string
		content  []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty file",
			args: args{
				postUUID: posts[0].Uuid.String(),
				content:  []byte{},
			},
			wantErr: false,
		},
		{
			name: "small file",
			args: args{
				postUUID: posts[0].Uuid.String(),
				content:  []byte{'t', 'e', 's', 't'},
			},
			wantErr: false,
		},
		{
			name: "bad post ID",
			args: args{
				postUUID: "92311213-081e-4286-96b9-599609676552",
				content:  []byte{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i Image
			if err := i.Store(tt.args.postUUID, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func CreateImageFixtures(t *testing.T, posts Posts) Images {
	const n = 2
	images := make(Images, n)

	for i := 0; i < n; i++ {
		var image Image
		if err := image.Store(posts[i].Uuid.String(), []byte{}); err != nil {
			t.Errorf("failed to create image fixture %d", i)
		}
		images[i] = image
	}

	images[1].URLExpiration = time.Now().Add(-time.Minute)
	if err := DB.Save(&images[1]); err != nil {
		t.Errorf("failed to update image fixture")
	}

	return images
}

func (ms *ModelSuite) TestImage_FindByUUID() {
	t := ms.T()
	_, users, _ := CreateUserFixtures(t)
	posts := CreatePostFixtures(t, users)
	images := CreateImageFixtures(t, posts)

	type args struct {
		postUUID  string
		imageUUID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good",
			args: args{
				postUUID:  posts[0].Uuid.String(),
				imageUUID: images[0].UUID.String(),
			},
		},
		{
			name: "needs refresh",
			args: args{
				postUUID:  posts[1].Uuid.String(),
				imageUUID: images[1].UUID.String(),
			},
		},
		{
			name: "wrong image ID",
			args: args{
				postUUID:  posts[0].Uuid.String(),
				imageUUID: images[1].UUID.String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var i Image
			err := i.FindByUUID(tt.args.postUUID, tt.args.imageUUID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("error = %v, postID = %s, imageID = %s", err, tt.args.postUUID, tt.args.imageUUID)
				} else {
					ms.Equal(tt.args.imageUUID, i.UUID.String(), "retrieved image has wrong UUID")
					ms.Contains(i.URL.String, "http", "URL doesn't start with 'http'")
					ms.True(i.URLExpiration.After(time.Now()), "URLExpiration is in the past")
				}
			}
		})
	}
}
