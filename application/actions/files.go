package actions

import (
	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/models"
)

func convertFile(file models.File) api.File {
	return api.File{
		ID:            file.UUID,
		URL:           file.URL,
		URLExpiration: file.URLExpiration,
		Name:          file.Name,
		Size:          file.Size,
		ContentType:   file.ContentType,
	}
}
