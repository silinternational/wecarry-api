package actions

import (
	"io/ioutil"
	"net/http"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/models"
)

const FileTagName = "file"

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
	f, err := c.File(FileTagName)
	if err != nil {
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{Code: "ErrorReceivingFile", Message: err.Error()},
		}))
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{Code: "UnableToReadFile", Message: err.Error()},
		}))
	}

	var fileObject models.File
	if err := fileObject.Store(f.Filename, content); err != nil {
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{Code: "UnableToStoreFile", Message: err.Error()},
		}))
	}

	resp := UploadResponse{
		Name:        fileObject.Name,
		UUID:        fileObject.UUID.String(),
		URL:         fileObject.URL.String,
		ContentType: fileObject.ContentType,
		Size:        fileObject.Size,
	}

	return c.Render(200, render.JSON(resp))
}
