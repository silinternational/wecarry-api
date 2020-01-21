package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif" // enable decoding of GIF images
	"image/jpeg"  // decode/encode JPEG images
	_ "image/png" // enable decoding of PNG images
	"net/http"
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

	contentType, err := detectContentType(content)
	if err != nil {
		e := FileUploadError{
			HttpStatus: http.StatusBadRequest,
			ErrorCode:  domain.ErrorStoreFileBadContentType,
			Message:    err.Error(),
		}
		return &e
	}

	// If possible, strip EXIF metadata by re-encoding the image. Also sets background to white.
	img, _, err := image.Decode(bytes.NewReader(content))
	if err == nil {
		dst := image.NewRGBA(img.Bounds())
		draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Over)
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, dst, nil); err == nil {
			content = buf.Bytes()
			contentType = "image/jpg"
			name = name + ".jpeg"
		}
	}

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
	if err := DB.Select("id").All(&files); err != nil {
		return err
	}

	toDelete := make(map[int]bool, len(files))
	for i := range files {
		toDelete[files[i].ID] = true
	}

	var posts Posts
	if err := DB.Select("photo_file_id").Where("photo_file_id is not null").All(&posts); err != nil {
		return err
	}
	for _, p := range posts {
		toDelete[p.PhotoFileID.Int] = false
	}

	var postFiles PostFiles
	if err := DB.Select("file_id").All(&postFiles); err != nil {
		return err
	}
	for _, p := range postFiles {
		toDelete[p.FileID] = false
	}

	var meetings Meetings
	if err := DB.Select("image_file_id").Where("image_file_id is not null").All(&meetings); err != nil {
		return err
	}
	for _, m := range meetings {
		toDelete[m.ImageFileID.Int] = false
	}

	var organizations Organizations
	if err := DB.Select("logo_file_id").Where("logo_file_id is not null").All(&organizations); err != nil {
		return err
	}
	for _, o := range organizations {
		toDelete[o.LogoFileID.Int] = false
	}

	var users Users
	if err := DB.Select("photo_file_id").Where("photo_file_id is not null").All(&users); err != nil {
		return err
	}
	for _, u := range users {
		toDelete[u.PhotoFileID.Int] = false
	}

	for id, del := range toDelete {
		if del {
			var file File
			if err := DB.Select("id", "uuid").Find(&file, id); err != nil {
				domain.ErrLogger.Printf("file %d not found, %s", id, err)
				continue
			}

			if err := aws.RemoveFile(file.UUID.String()); err != nil {
				domain.ErrLogger.Printf("error removing from S3, id='%s', %s", file.UUID.String(), err)
				continue
			}

			if err := DB.Destroy(&file); err != nil {
				domain.ErrLogger.Printf("file %d destroy error, %s", id, err)
			}
		}
	}
	return nil
}
