package gqlgen

import (
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

// ConvertDBPostToGqlPost does what its name says, but also ...
func ConvertDBPostToGqlPost(dbPost models.Post) (Post, error) {
	dbID := strconv.Itoa(dbPost.ID)

	newGqlPost := Post{
		ID:           dbID,
		Type:         PostType(dbPost.Type),
		Title:        dbPost.Title,
		Description:  GetStringFromNullsString(dbPost.Description),
		Destination:  GetStringFromNullsString(dbPost.Destination),
		Origin:       GetStringFromNullsString(dbPost.Origin),
		Size:         dbPost.Size,
		NeededAfter:  domain.ConvertTimeToStringPtr(dbPost.NeededAfter),
		NeededBefore: domain.ConvertTimeToStringPtr(dbPost.NeededBefore),
		Category:     dbPost.Category,
		Status:       dbPost.Status,
		CreatedAt:    domain.ConvertTimeToStringPtr(dbPost.CreatedAt),
		UpdatedAt:    domain.ConvertTimeToStringPtr(dbPost.UpdatedAt),
	}

	return newGqlPost, nil
}
