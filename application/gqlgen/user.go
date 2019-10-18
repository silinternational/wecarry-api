package gqlgen

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
)

var PostRoleMap = map[PostRole]string{
	PostRoleCreatedby: models.PostRoleCreatedby,
	PostRoleReceiving: models.PostRoleReceiving,
	PostRoleProviding: models.PostRoleProviding,
}

// UserFields maps GraphQL fields to their equivalent database fields. For related types, the
// foreign key field name is provided.
func UserFields() map[string]string {
	return map[string]string{
		"id":          "uuid",
		"email":       "email",
		"nickname":    "nickname",
		"accessToken": "access_token",
		"createdAt":   "created_at",
		"updatedAt":   "updated_at",
		"adminRole":   "admin_role",
		"photoURL":    "photo_url",
		"photoFile":   "photo_file_id",
		"location":    "location_id",
	}
}

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.Uuid.String(), nil
}

func (r *userResolver) AdminRole(ctx context.Context, obj *models.User) (*Role, error) {
	if obj == nil {
		return nil, nil
	}
	a := Role(obj.AdminRole.String)
	return &a, nil
}

func (r *userResolver) Organizations(ctx context.Context, obj *models.User) ([]*models.Organization, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetOrganizations()
}

func (r *userResolver) Posts(ctx context.Context, obj *models.User, role PostRole) ([]*models.Post, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetPosts(PostRoleMap[role])
}

// PhotoURL retrieves a URL for the user profile photo or avatar. It can either be an attached photo or
// a photo belonging to an external profile such as Gravatar or Google.
func (r *userResolver) PhotoURL(ctx context.Context, obj *models.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.GetPhotoURL()
}

func (r *userResolver) Location(ctx context.Context, obj *models.User) (*models.Location, error) {
	if obj == nil {
		return nil, nil
	}
	return obj.GetLocation()
}

func (r *queryResolver) Users(ctx context.Context) ([]*models.User, error) {
	db := models.DB
	var dbUsers []*models.User

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin {
		err := errors.New("not authorized")
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return []*models.User{}, err
	}

	if err := db.Select(GetSelectFieldsForUsers(ctx)...).All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return []*models.User{}, err
	}

	return dbUsers, nil
}

func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	dbUser := models.User{}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if id == nil {
		return &currentUser, nil
	}

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && currentUser.Uuid.String() != *id {
		err := errors.New("not authorized")
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return &dbUser, err
	}

	if err := models.DB.Select(GetSelectFieldsForUsers(ctx)...).Where("uuid = ?", id).First(&dbUser); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting user: %v", err.Error()))
		domain.Warn(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return &dbUser, err
	}

	return &dbUser, nil
}

// GetSelectFieldsForUsers returns a list of database fields appropriate for the current query. Foreign keys
// will be included as needed.
func GetSelectFieldsForUsers(ctx context.Context) []string {
	selectFields := GetSelectFieldsFromRequestFields(UserFields(), graphql.CollectAllFields(ctx))
	selectFields = append(selectFields, "id")
	if domain.IsStringInSlice("photoURL", graphql.CollectAllFields(ctx)) {
		selectFields = append(selectFields, "photo_file_id")
	}
	return selectFields
}

// UpdateUser takes data from the GraphQL `UpdateUser` mutation and updates the database. If the
// user ID is provided and the current user is allowed to edit profiles, that user will be updated.
// Otherwise, the current authenticated user is updated.
func (r *mutationResolver) UpdateUser(ctx context.Context, input UpdateUserInput) (*models.User, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	var user models.User

	if input.ID != nil {
		err := user.FindByUUID(*(input.ID))
		if err != nil {
			domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
			return &models.User{}, err
		}
	} else {
		user = cUser
	}

	if cUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && cUser.ID != user.ID {
		err := errors.New("user not allowed to edit user profiles")
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
		return &models.User{}, err
	}

	if input.PhotoID != nil {
		var file models.File
		err := file.FindByUUID(*input.PhotoID)
		if err != nil {
			domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
			return &models.User{}, err
		}
		user.PhotoFileID = nulls.NewInt(file.ID)
	}

	if input.Location != nil {
		err := user.SetLocation(convertGqlLocationInputToDBLocation(*input.Location))
		if err != nil {
			domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error())
			return &models.User{}, err
		}
	}

	err := user.Save()

	return &user, err
}
