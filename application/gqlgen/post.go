package gqlgen

import (
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

// ConvertDBPostToGqlPost does what its name says, but also ...
func ConvertDBPostToGqlPost(dbPost models.Post) (Post, error) {
	dbID := strconv.Itoa(dbPost.ID)

	stubAfter := "0000"
	stubBefore := "9999"
	stubCategory := "Eats"

	newGqlPost := Post{
		ID:           dbID,
		Type:         PostType(dbPost.Type),
		Title:        dbPost.Title,
		Description:  GetStringFromNullsString(dbPost.Description),
		Destination:  GetStringFromNullsString(dbPost.Destination),
		Origin:       GetStringFromNullsString(dbPost.Origin),
		Size:         dbPost.Size,
		NeededAfter:  &stubAfter,    //GetStringFromNullsString(dbPost.NeededAfter),
		NeededBefore: &stubBefore,   //GetStringFromNullsString(dbPost.NeededBefore),
		Category:     &stubCategory, // GetStringFromNullsString(dbPost.Category),
		Status:       dbPost.Status,
		CreatedAt:    ConvertTimeToStringPtr(dbPost.CreatedAt),
		UpdatedAt:    ConvertTimeToStringPtr(dbPost.UpdatedAt),
	}

	return newGqlPost, nil
}
