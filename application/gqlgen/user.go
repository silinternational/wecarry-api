package gqlgen

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
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

func (r *userResolver) CreatedAt(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *userResolver) UpdatedAt(ctx context.Context, obj *models.User) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *userResolver) AdminRole(ctx context.Context, obj *models.User) (*Role, error) {
	if obj == nil {
		return nil, nil
	}
	a := Role(obj.AdminRole.String)
	return &a, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*models.User, error) {
	db := models.DB
	var dbUsers []*models.User

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin {
		return []*models.User{}, fmt.Errorf("not authorized")
	}

	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), graphql.CollectAllFields(ctx))
	if err := db.Select(selectFields...).All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		return []*models.User{}, err
	}

	return dbUsers, nil
}

func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	dbUser := models.User{}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && currentUser.Uuid.String() != *id {
		return &dbUser, fmt.Errorf("not authorized")
	}

	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), graphql.CollectAllFields(ctx))
	if err := models.DB.Select(selectFields...).Where("uuid = ?", id).First(&dbUser); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting user: %v", err.Error()))
		return &dbUser, err
	}

	return &dbUser, nil
}
