package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gobuffalo/validate/validators"
	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"
)

type Image struct {
	ID            int          `json:"id" db:"id"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	UUID          uuid.UUID    `json:"uuid" db:"uuid"`
	PostID        int          `json:"post_id" db:"post_id"`
	URL           nulls.String `json:"url" db:"url"`
	URLExpiration time.Time    `json:"url_expiration" db:"url_expiration"`
	Post          Post         `belongs_to:"posts"`
}

// String is not required by pop and may be deleted
func (i Image) String() string {
	ji, _ := json.Marshal(i)
	return string(ji)
}

// Images is not required by pop and may be deleted
type Images []Image

// String is not required by pop and may be deleted
func (i Images) String() string {
	ji, _ := json.Marshal(i)
	return string(ji)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (i *Image) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: i.UUID, Name: "UUID"},
		&validators.IntIsPresent{Field: i.PostID, Name: "post_id"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (i *Image) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (i *Image) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Store takes a byte slice and stores it into S3 and saves the metadata in the database image table.
// None of the struct members of i are used as input, but are updated if the function is successful.
func (i *Image) Store(postUUID string, content []byte) error {
	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return err
	}

	imageUUID := domain.GetUuid()

	if len(content) > domain.MaxFileSize {
		return fmt.Errorf("file too large (%d bytes), max is %d bytes", len(content), domain.MaxFileSize)
	}

	contentType, err := detectContentType(content)
	if err != nil {
		return err
	}

	url, err := domain.StoreFile(postUUID+"/"+imageUUID.String(), contentType, content)
	if err != nil {
		return err
	}

	image := Image{
		UUID:          imageUUID,
		PostID:        post.ID,
		URL:           nulls.NewString(url.Url),
		URLExpiration: url.Expiration,
	}
	if err := DB.Save(&image); err != nil {
		return err
	}

	*i = image
	return nil
}

// FindByUUID locates an image by Post UUID and Image UUID and returns the result, including a valid URL.
// None of the struct members of i are used as input, but are updated if the function is successful.
func (i *Image) FindByUUID(postUUID, imageUUID string) error {
	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return err
	}

	var image Image
	if err := DB.Where("post_id = ? AND uuid = ?", post.ID, imageUUID).First(&image); err != nil {
		return err
	}

	if err := image.RefreshURL(); err != nil {
		return err
	}

	*i = image
	return nil
}

// RefreshURL ensures the image URL is good for at least a few minutes
func (i *Image) RefreshURL() error {
	if i.URLExpiration.After(time.Now().Add(time.Minute * 5)) {
		return nil
	}

	newURL, err := domain.GetFileURL(i.Post.Uuid.String() + "/" + i.UUID.String())
	if err != nil {
		return err
	}
	i.URL = nulls.NewString(newURL.Url)
	i.URLExpiration = newURL.Expiration
	if err = DB.Save(i); err != nil {
		return err
	}
	return nil
}

func detectContentType(content []byte) (string, error) {
	allowedTypes := []string{
		"image/bmp",
		"image/gif",
		"image/jpeg",
		"image/png",
		"image/webp",
	}

	detectedType := http.DetectContentType(content)
	for _, t := range allowedTypes {
		if detectedType == t {
			return t, nil
		}
	}
	return "", fmt.Errorf("invalid file type %s", detectedType)
}
