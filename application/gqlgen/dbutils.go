package gqlgen

import (
	"context"
	"fmt"
	"github.com/silinternational/handcarry-api/models"
	"strconv"
	"time"
)

// ConvertTimeToStringPtr is intended to convert the
// CreatedAt and UpdatedAt fields of database objects
// to pointers to strings to populate the same gqlgen fields
func ConvertTimeToStringPtr(inTime time.Time) *string {
	inTimeStr := fmt.Sprintf("%v", inTime)
	return &inTimeStr
}

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertDBUserToGqlUser(dbUser models.User, ctx context.Context) (User, error) {
	dbID := strconv.Itoa(dbUser.ID)

	r := GetStringFromNullsString(dbUser.AdminRole)
	gqlRole := Role(*r)

	newGqlUser := User{
		ID:        dbID,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		Nickname:  dbUser.Nickname,
		AdminRole: &gqlRole,
		CreatedAt: ConvertTimeToStringPtr(dbUser.CreatedAt),
		UpdatedAt: ConvertTimeToStringPtr(dbUser.UpdatedAt),
	}

	return newGqlUser, nil
}

// ConvertDBPostToGqlPost does what its name says, but also ...
func ConvertDBPostToGqlPost(dbPost models.Post, ctx context.Context) (Post, error) {
	dbID := strconv.Itoa(dbPost.ID)

	stubAfter := "0000"
	stubBefore := "9999"
	stubCategory := "Eats"

	newGqlPost := Post{
		ID:        dbID,
		Type: PostType(dbPost.Type),
		Title: dbPost.Title,
		Description: GetStringFromNullsString(dbPost.Description),
		Destination: GetStringFromNullsString(dbPost.Destination),
		Origin: GetStringFromNullsString(dbPost.Origin),
		Size: dbPost.Size,
		NeededAfter: &stubAfter, //GetStringFromNullsString(dbPost.NeededAfter),
		NeededBefore: &stubBefore, //GetStringFromNullsString(dbPost.NeededBefore),
		Category: &stubCategory, // GetStringFromNullsString(dbPost.Category),
		Status: dbPost.Status,
		CreatedAt: ConvertTimeToStringPtr(dbPost.CreatedAt),
		UpdatedAt: ConvertTimeToStringPtr(dbPost.UpdatedAt),
	}


	return newGqlPost, nil
}
