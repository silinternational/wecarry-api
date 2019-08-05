package gqlgen

import (
	"fmt"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

func addMyThreadIDToPost(gqlPost *Post, dbPost models.Post, currentUser *models.User) error {
	if currentUser == nil {
		return nil
	}

	thread, err := models.FindThreadByPostIDAndUserID(dbPost.ID, currentUser.ID)
	if err != nil {
		return err
	}

	threadUuid := thread.Uuid.String()
	s := ""
	gqlPost.MyThreadID = &s
	if threadUuid != domain.EmptyUUID {
		gqlPost.MyThreadID = &threadUuid
	}
	return nil
}

func addCreatedByToPost(gqlPost *Post, dbPost models.Post, requestFields []string) error {
	if !domain.IsStringInSlice(CreatedByField, requestFields) {
		return nil
	}

	creator := models.User{}
	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)
	if err := models.DB.Select(selectFields...).Find(&creator, dbPost.CreatedByID); err != nil {
		return err
	}

	gqlPost.CreatedBy = &creator

	return nil
}

// ConvertDBPostToGqlPost does what its name says, but also adds the ID
// of the first thread that is associated with the Post and the current user ...
func ConvertDBPostToGqlPost(dbPost models.Post, currentUser *models.User, requestFields []string) (Post, error) {
	gqlPost := Post{
		ID:           dbPost.Uuid.String(),
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

	if err := addMyThreadIDToPost(&gqlPost, dbPost, currentUser); err != nil {
		return gqlPost, err
	}

	err := addCreatedByToPost(&gqlPost, dbPost, requestFields)

	return gqlPost, err
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
