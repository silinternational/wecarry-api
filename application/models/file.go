package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/silinternational/wecarry-api/aws"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
)

type File struct {
	ID            int       `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	UUID          uuid.UUID `json:"uuid" db:"uuid"`
	URL           string    `json:"url" db:"url"`
	URLExpiration time.Time `json:"url_expiration" db:"url_expiration"`
	Name          string    `json:"name" db:"name"`
	Size          int       `json:"size" db:"size"`
	ContentType   string    `json:"content_type" db:"content_type"`
}

// String can be helpful for serializing the model
func (f File) String() string {
	ji, _ := json.Marshal(f)
	return string(ji)
}

// Files is merely for convenience and brevity
type Files []File

// String can be helpful for serializing the model
func (i Files) String() string {
	ji, _ := json.Marshal(i)
	return string(ji)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (f *File) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: f.UUID, Name: "UUID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (f *File) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (f *File) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// Store takes a byte slice and stores it into S3 and saves the metadata in the database file table.
// None of the struct members of `f` are used as input, but are updated if the function is successful.
func (f *File) Store(name string, content []byte) error {
	fileUUID := domain.GetUUID()

	if len(content) > domain.MaxFileSize {
		return fmt.Errorf("file too large (%d bytes), max is %d bytes", len(content), domain.MaxFileSize)
	}

	contentType, err := detectContentType(content)
	if err != nil {
		return err
	}

	url, err := aws.StoreFile(fileUUID.String(), contentType, content)
	if err != nil {
		return err
	}

	file := File{
		UUID:          fileUUID,
		URL:           url.Url,
		URLExpiration: url.Expiration,
		Name:          name,
		Size:          len(content),
		ContentType:   contentType,
	}
	if err := DB.Save(&file); err != nil {
		return err
	}

	*f = file
	return nil
}

// FindByUUID locates an file by UUID and returns the result, including a valid URL.
// None of the struct members of f are used as input, but are updated if the function is successful.
func (f *File) FindByUUID(fileUUID string) error {
	var file File
	if err := DB.Where("uuid = ?", fileUUID).First(&file); err != nil {
		return err
	}

	if err := file.refreshURL(); err != nil {
		return err
	}

	*f = file
	return nil
}

// refreshURL ensures the file URL is good for at least a few minutes
func (f *File) refreshURL() error {
	if f.URLExpiration.After(time.Now().Add(time.Minute * 5)) {
		return nil
	}

	newURL, err := aws.GetFileURL(f.UUID.String())
	if err != nil {
		return err
	}
	f.URL = newURL.Url
	f.URLExpiration = newURL.Expiration
	if err = DB.Update(f); err != nil {
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
		"application/pdf",
	}

	detectedType := http.DetectContentType(content)
	if domain.IsStringInSlice(detectedType, allowedTypes) {
		return detectedType, nil
	}
	return "", fmt.Errorf("invalid file type %s", detectedType)
}
