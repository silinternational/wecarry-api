package gqlgen

import (
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

func UserSimpleFields() map[string]string {
	return map[string]string{
		"id":          "uuid",
		"email":       "email",
		"firstName":   "first_name",
		"lastName":    "last_name",
		"nickname":    "nickname",
		"accessToken": "access_token",
		"createdAt":   "created_at",
		"updatedAt":   "updated_at",
		"adminRole":   "admin_role",
	}
}

// ConvertDBUserToGqlUser does what its name says, but also ...
func ConvertDBUserToGqlUser(dbUser models.User) (User, error) {

	r := GetStringFromNullsString(dbUser.AdminRole)
	gqlRole := Role(*r)

	newGqlUser := User{
		ID:        dbUser.Uuid.String(),
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
