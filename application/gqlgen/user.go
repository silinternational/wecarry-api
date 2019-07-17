package gqlgen

import (
	"github.com/silinternational/handcarry-api/models"
	"strconv"
)

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertDBUserToGqlUser(dbUser models.User) (User, error) {
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
