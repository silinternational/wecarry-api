package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/validate/validators"
	"github.com/silinternational/handcarry-api/domain"

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
		&validators.IntIsPresent{Field: i.PostID, Name: "ThreadID"},
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

func (i *Image) Store(postUUID string, content []byte) error {
	var post Post
	if err := post.FindByUUID(postUUID); err != nil {
		return err
	}

	imageUUID := domain.GetUuid()
	url, err := domain.StoreFile(postUUID+"/"+imageUUID.String(), "binary/octet-stream", content)
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

// RefreshURL ensures the URL is good for at least a few minutes
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
