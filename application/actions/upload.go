package actions

import (
	"io/ioutil"
	"net/http"

	"github.com/silinternational/wecarry-api/domain"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/silinternational/wecarry-api/models"
)

// fileFieldName is the multipart field name for the file upload.
const fileFieldName = "file"

// UploadResponse is a JSON response for the /upload endpoint
type UploadResponse struct {
	Error       *domain.AppError `json:"Error,omitempty"`
	Name        string           `json:"filename,omitempty"`
	UUID        string           `json:"id,omitempty"`
	URL         string           `json:"url,omitempty"`
	ContentType string           `json:"content_type,omitempty"`
	Size        int              `json:"size,omitempty"`
}

// uploadHandler responds to POST requests at /upload
func uploadHandler(c buffalo.Context) error {
	f, err := c.File(fileFieldName)
	if err != nil {
		domain.ErrLogger.Printf("error getting uploaded file from context ... %v", err)
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{
				Code: http.StatusInternalServerError,
				Key:  domain.ErrorReceivingFile,
			},
		}))
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		domain.ErrLogger.Printf("error reading uploaded file ... %v", err)
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{
				Code: http.StatusInternalServerError,
				Key:  domain.UnableToReadFile,
			},
		}))
	}

	var fileObject models.File
	if err := fileObject.Store(f.Filename, content); err != nil {
		domain.ErrLogger.Printf("error storing uploaded file ... %v", err)
		return c.Render(http.StatusInternalServerError, render.JSON(UploadResponse{
			Error: &domain.AppError{
				Code: http.StatusInternalServerError,
				Key:  domain.UnableToStoreFile,
			},
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
