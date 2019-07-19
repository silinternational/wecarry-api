package gqlgen

import (
	"fmt"
	"strconv"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

// ConvertDBPostToGqlPost does what its name says, but also adds the ID
// of the first thread that is associated with the Post and the current user ...
func ConvertDBPostToGqlPost(dbPost models.Post, currentUser *models.User) (Post, error) {
	dbID := strconv.Itoa(dbPost.ID)

	gqlPost := Post{
		ID:           dbID,
		UUID:         dbPost.Uuid.String(),
		Type:         PostType(dbPost.Type),
		Title:        dbPost.Title,
		Description:  GetStringFromNullsString(dbPost.Description),
		Destination:  GetStringFromNullsString(dbPost.Destination),
		Origin:       GetStringFromNullsString(dbPost.Origin),
		Size:         dbPost.Size,
		NeededAfter:  domain.ConvertDateToStringPtr(dbPost.NeededAfter),
		NeededBefore: domain.ConvertDateToStringPtr(dbPost.NeededBefore),
		Category:     dbPost.Category,
		Status:       dbPost.Status,
		CreatedAt:    domain.ConvertTimeToStringPtr(dbPost.CreatedAt),
		UpdatedAt:    domain.ConvertTimeToStringPtr(dbPost.UpdatedAt),
	}

	if currentUser == nil {
		return gqlPost, nil
	}

	thread, err := models.FindThreadByPostIDAndUserID(dbPost.ID, currentUser.ID)
	if err != nil {
		return gqlPost, err
	}

	threadUuid := thread.Uuid.String()
	s := ""
	gqlPost.MyThreadID = &s
	if threadUuid != domain.EmptyUUID {
		gqlPost.MyThreadID = &threadUuid
	}

	return gqlPost, nil
}

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
