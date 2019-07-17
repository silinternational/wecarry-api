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


	newGqlPost := Post{
		ID:        dbID,
		Type: PostType(dbPost.Type),
		CreatedAt: ConvertTimeToStringPtr(dbPost.CreatedAt),
		UpdatedAt: ConvertTimeToStringPtr(dbPost.UpdatedAt),
	}

	return newGqlPost, nil
}
