package gqlgen

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

// ConvertDBPostToGqlPost does what its name says, but also ...
func ConvertDBPostToGqlPost(dbPost models.Post) (Post, error) {
	dbID := strconv.Itoa(dbPost.ID)

	newGqlPost := Post{
		ID:           dbID,
		UUID:         dbPost.Uuid.String(),
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

func ConvertGqlNewPostToDBPost(gqlPost NewPost) (models.Post, error) {

	createdByUser, err := models.FindUserByUUID(gqlPost.CreatedByID)
	if err != nil {
		return models.Post{}, err
	}

	org, err := models.FindOrgByUUID(gqlPost.OrgID)
	if err != nil {
		return models.Post{}, err
	}

	dbPost := models.Post{}

	dbPost.CreatedByID = createdByUser.ID
	dbPost.OrgID = org.ID
	dbPost.Type = gqlPost.Type.String()
	dbPost.Title = gqlPost.Title

	dbPost.Description = nulls.NewString(*gqlPost.Description)
	dbPost.Origin = nulls.NewString(*gqlPost.Origin)
	dbPost.Size = gqlPost.Size

	//dbPost.NeededAfter = domain.ConvertStringPtrToTime(gqlPost.NeededAfter)
	//dbPost.NeededBefore = domain.ConvertStringPtrToTime(gqlPost.NeededBefore)
	dbPost.Category = domain.ConvertStrPtrToString(gqlPost.Category)

	return dbPost, nil

}
