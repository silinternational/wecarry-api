package gqlgen

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
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
		CreatedAt: domain.ConvertTimeToStringPtr(dbUser.CreatedAt),
		UpdatedAt: domain.ConvertTimeToStringPtr(dbUser.UpdatedAt),
	}

	return newGqlUser, nil
}

func ConvertGqlUserToDbUser(user User) (models.User, error) {
	role := user.AdminRole
	var dbRole string
	if role == nil {
		dbRole = ""
	} else {
		dbRole = role.String()
	}

	newDbUser := models.User{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		AdminRole: nulls.NewString(dbRole),
	}

	return newDbUser, nil
}
