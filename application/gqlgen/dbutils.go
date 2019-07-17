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

// ConvertDBUserToGqlUser does what its name says, but also converts the user's
// spouse (if there is one) and cars (if there are any)
// spouse is a "belongs_to" relationship anc cars is a "many_to_many" relationship
func ConvertDBUserToGqlUser(dbUser models.User, ctx context.Context) (User, error) {
	dbID := strconv.Itoa(dbUser.ID)

	newGqlUser := User{
		ID:        &dbID,
		Nickname:  &dbUser.Nickname,
		AdminRole: GetStringFromNullsString(dbUser.AdminRole),
		CreatedAt: ConvertTimeToStringPtr(dbUser.CreatedAt),
		UpdatedAt: ConvertTimeToStringPtr(dbUser.UpdatedAt),
	}

	return newGqlUser, nil
}