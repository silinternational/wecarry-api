package gqlgen

import (
	"fmt"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

// ConvertGqlNewPostToDBPost does what its name says, but also ...
func ConvertGqlNewPostToDBPost(gqlPost NewPost, createdByUser models.User) (models.Post, error) {
	org, err := models.FindOrgByUUID(gqlPost.OrgID)
	if err != nil {
		return models.Post{}, err
	}

	dbPost := models.Post{}
	dbPost.Uuid = domain.GetUuid()

	dbPost.CreatedByID = createdByUser.ID
	dbPost.OrganizationID = org.ID
	dbPost.Type = gqlPost.Type.String()
	dbPost.Title = gqlPost.Title

	dbPost.Description = models.ConvertStringPtrToNullsString(gqlPost.Description)
	dbPost.Destination = models.ConvertStringPtrToNullsString(gqlPost.Destination)
	dbPost.Origin = models.ConvertStringPtrToNullsString(gqlPost.Origin)

	dbPost.Size = gqlPost.Size

	neededAfter, err := domain.ConvertStringPtrToDate(gqlPost.NeededAfter)
	if err != nil {
		err = fmt.Errorf("error converting NeededAfter %v ... %v", gqlPost.NeededAfter, err.Error())
		return models.Post{}, err
	}

	dbPost.NeededAfter = neededAfter

	neededBefore, err := domain.ConvertStringPtrToDate(gqlPost.NeededBefore)
	if err != nil {
		err = fmt.Errorf("error converting NeededBefore %v ... %v", gqlPost.NeededBefore, err.Error())
		return models.Post{}, err
	}

	dbPost.NeededBefore = neededBefore
	dbPost.Category = domain.ConvertStrPtrToString(gqlPost.Category)

	return dbPost, nil
}
