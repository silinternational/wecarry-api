package gqlgen

import (
	"strconv"

	"github.com/gobuffalo/nulls"
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
		NeededAfter:  domain.ConvertTimeToStringPtr(dbPost.NeededAfter),
		NeededBefore: domain.ConvertTimeToStringPtr(dbPost.NeededBefore),
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

	dbPost.Description = nulls.NewString(*gqlPost.Description)
	dbPost.Destination = nulls.NewString(*gqlPost.Destination)
	dbPost.Origin = nulls.NewString(*gqlPost.Origin)
	dbPost.Size = gqlPost.Size

	//dbPost.NeededAfter = domain.ConvertStringPtrToTime(gqlPost.NeededAfter)
	//dbPost.NeededBefore = domain.ConvertStringPtrToTime(gqlPost.NeededBefore)
	dbPost.Category = domain.ConvertStrPtrToString(gqlPost.Category)

	return dbPost, nil
}
