package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	_ "golang.org/x/image/webp" // enable decoding of WEBP images

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
)

type FileUploadError struct {
	HttpStatus int
	ErrorCode  string
	Message    string
}

func (f *FileUploadError) Error() string {
	return fmt.Sprintf("%d: %s ... %s", f.HttpStatus, f.ErrorCode, f.Message)
}

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
	Linked        bool      `json:"linked" db:"linked"`
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
func (f *File) Store(name string, content []byte) *FileUploadError {
	fileUUID := domain.GetUUID()

	if len(content) > domain.MaxFileSize {
		e := FileUploadError{
			HttpStatus: http.StatusBadRequest,
			ErrorCode:  domain.ErrorStoreFileTooLarge,
			Message:    fmt.Sprintf("file too large (%d bytes), max is %d bytes", len(content), domain.MaxFileSize),
		}
		return &e
	}

	contentType, err := validateContentType(content)
	if err != nil {
		e := FileUploadError{
			HttpStatus: http.StatusBadRequest,
			ErrorCode:  domain.ErrorStoreFileBadContentType,
			Message:    err.Error(),
		}
		return &e
	}

	removeMetadata(&contentType, &content)
	changeFileExtension(&name, contentType)

	url, err := aws.StoreFile(fileUUID.String(), contentType, content)
	if err != nil {
		e := FileUploadError{
			HttpStatus: http.StatusInternalServerError,
			ErrorCode:  domain.ErrorUnableToStoreFile,
			Message:    err.Error(),
		}
		return &e
	}

	file := File{
		UUID:          fileUUID,
		URL:           url.Url,
		URLExpiration: url.Expiration,
		Name:          name,
		Size:          len(content),
		ContentType:   contentType,
	}
	if err := file.Create(); err != nil {
		e := FileUploadError{
			HttpStatus: http.StatusInternalServerError,
			ErrorCode:  domain.ErrorUnableToStoreFile,
			Message:    err.Error(),
		}
		return &e
	}

	*f = file
	return nil
}

// removeMetadata removes, if possible, all EXIF metadata by re-encoding the image. If the encoding type changes,
// `contentType` will be modified accordingly.
func removeMetadata(contentType *string, content *[]byte) {
	img, _, err := image.Decode(bytes.NewReader(*content))
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	switch *contentType {
	case "image/jpg":
		if err := jpeg.Encode(buf, img, nil); err == nil {
			*content = buf.Bytes()
		}
	case "image/gif":
		if err := gif.Encode(buf, img, nil); err == nil {
			*content = buf.Bytes()
		}
	case "image/png":
		if err := png.Encode(buf, img); err == nil {
			*content = buf.Bytes()
		}
	case "image/webp":
		if err := png.Encode(buf, img); err == nil {
			*content = buf.Bytes()
			*contentType = "image/png"
		}
	}
}

// changeFileExtension attempts to make the file extension match the given content type
func changeFileExtension(name *string, contentType string) {
	ext, err := mime.ExtensionsByType(contentType)
	if err != nil || len(ext) < 1 {
		return
	}
	*name = strings.TrimSuffix(*name, filepath.Ext(*name)) + ext[0]

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
	if err = f.Update(); err != nil {
		return err
	}
	return nil
}

func validateContentType(content []byte) (string, error) {
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

// Create stores the File data as a new record in the database.
func (f *File) Create() error {
	return create(f)
}

// Update writes the File data to an existing database record.
func (f *File) Update() error {
	return update(f)
}

// DeleteUnlinked removes all files that are no longer linked to any database records
func (f *Files) DeleteUnlinked() error {
	var files Files
	if err := DB.Select("id", "uuid").Where("linked = 0").All(&files); err != nil {
		return err
	}

	nRemovedFromDB := 0
	nRemovedFromS3 := 0
	for _, file := range files {
		if err := aws.RemoveFile(file.UUID.String()); err != nil {
			domain.ErrLogger.Printf("error removing from S3, id='%s', %s", file.UUID.String(), err)
			continue
		}
		nRemovedFromS3++

		if err := DB.Destroy(&file); err != nil {
			domain.ErrLogger.Printf("file %d destroy error, %s", file.ID, err)
			continue
		}
		nRemovedFromDB++
	}

	if nRemovedFromDB < len(files) || nRemovedFromS3 < len(files) {
		domain.ErrLogger.Printf("not all unlinked files were removed")
	}
	domain.Logger.Printf("removed %d from S3, %d from file table", nRemovedFromS3, nRemovedFromDB)
	return nil
}

// SetLinked marks the file as linked. The struct need not be hydrated; only the ID is needed.
func (f *File) SetLinked() error {
	f.Linked = true
	return DB.UpdateColumns(f, "linked")
}

// ClearLinked marks the file as unlinked. The struct need not be hydrated; only the ID is needed.
func (f *File) ClearLinked() error {
	f.Linked = false
	return DB.UpdateColumns(f, "linked")
}
