package actions

import (
	"io/ioutil"
	"net/http"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/models"
)

// FileFieldName is the multipart field name for the file upload.
const FileFieldName = "file"

// UploadResponse is a JSON response for the /upload endpoint
type UploadResponse struct {
	Error       *domain.AppError `json:"Error,omitempty"`
	Name        string           `json:"filename,omitempty"`
	UUID        string           `json:"id,omitempty"`
	URL         string           `json:"url,omitempty"`
	ContentType string           `json:"content_type,omitempty"`
	Size        int              `json:"size,omitempty"`
}

// UploadHandler responds to POST requests at /upload
func UploadHandler(c buffalo.Context) error {
	f, err := c.File(FileFieldName)
	if err != nil {
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{Code: domain.ErrorReceivingFile, Message: err.Error()},
		}))
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{Code: domain.UnableToReadFile, Message: err.Error()},
		}))
	}

	var fileObject models.File
	if err := fileObject.Store(f.Filename, content); err != nil {
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{Code: domain.UnableToStoreFile, Message: err.Error()},
		}))
	}

	resp := UploadResponse{
		Name:        fileObject.Name,
		UUID:        fileObject.UUID.String(),
		URL:         fileObject.URL,
		ContentType: fileObject.ContentType,
		Size:        fileObject.Size,
	}

	return c.Render(200, render.JSON(resp))
}
