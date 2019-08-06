//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"context"
	"fmt"

	"github.com/silinternational/handcarry-api/domain"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// TestUser is intended as a way to inject a "current User" for unit tests
var TestUser models.User

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

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

	requestFields := GetRequestFields(ctx)
	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)

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

	requestFields := GetRequestFields(ctx)
	selectFields := GetSelectFieldsFromRequestFields(UserSimpleFields(), requestFields)

	if err := models.DB.Select(selectFields...).Where("uuid = ?", id).First(&dbUser); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting user: %v", err.Error()))
		return &dbUser, err
	}

	return &dbUser, nil
}

func (r *Resolver) Post() PostResolver {
	return &postResolver{r}
}

type postResolver struct{ *Resolver }

func (r *postResolver) Type(ctx context.Context, obj *models.Post) (PostType, error) {
	if obj == nil {
		return "", nil
	}
	return PostType(obj.Type), nil
}

func (r *postResolver) Organization(ctx context.Context, obj *models.Post) (*Organization, error) {
	return nil, nil
}

func (r *postResolver) Description(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Description), nil
}

func (r *postResolver) Destination(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Destination), nil
}

func (r *postResolver) Origin(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return GetStringFromNullsString(obj.Origin), nil
}

func (r *postResolver) NeededAfter(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededAfter), nil
}

func (r *postResolver) NeededBefore(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.NeededBefore), nil
}

func (r *postResolver) Threads(ctx context.Context, obj *models.Post) ([]*Thread, error) {
	return nil, nil
}

func (r *postResolver) CreatedAt(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.CreatedAt), nil
}

func (r *postResolver) UpdatedAt(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}
	return domain.ConvertTimeToStringPtr(obj.UpdatedAt), nil
}

func (r *postResolver) MyThreadID(ctx context.Context, obj *models.Post) (*string, error) {
	if obj == nil {
		return nil, nil
	}

	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	thread, err := models.FindThreadByPostIDAndUserID(obj.ID, currentUser.ID)
	if err != nil {
		return nil, err
	}

	threadUuid := thread.Uuid.String()
	if threadUuid == domain.EmptyUUID {
		return nil, nil
	}

	return &threadUuid, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*models.Post, error) {

	db := models.DB
	dbPosts := models.Posts{}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	if err := db.Where("organization_id IN (?)", cUser.GetOrgIDs()...).All(&dbPosts); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting posts: %v", err.Error()))
		return []*models.Post{}, err
	}

	var posts []*models.Post
	for _, dbPost := range dbPosts {
		posts = append(posts, &dbPost)
	}

	return posts, nil
}

func (r *queryResolver) Post(ctx context.Context, id *string) (*models.Post, error) {
	post := models.Post{}
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	if err := models.DB.Where("organization_id IN (?)", cUser.GetOrgIDs()...).Where("uuid = ?", id).First(&post); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting post: %v", err.Error()))
		return &models.Post{}, err
	}

	e := post.QueryRelatedUsers(GetSelectFieldsFromRequestFields(UserSimpleFields(), GetRequestFields(ctx)))
	if e != nil {
		return &post, e
	}

	return &post, nil
}

func (r *queryResolver) Threads(ctx context.Context) ([]*Thread, error) {

	db := models.DB
	dbThreads := models.Threads{}

	requestFields := GetRequestFields(ctx)
	selectFields := getSelectFieldsForThreads(requestFields)

	if err := db.Select(selectFields...).All(&dbThreads); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting threads: %v", err.Error()))
		return []*Thread{}, err
	}

	var gqlThreads []*Thread
	for _, dbThread := range dbThreads {
		newGqlThread, err := ConvertDBThreadToGqlThread(dbThread, requestFields)

		if err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error converting users: %v", err.Error()))
			return gqlThreads, err
		}
		gqlThreads = append(gqlThreads, &newGqlThread)

	}

	return gqlThreads, nil
}

func (r *queryResolver) MyThreads(ctx context.Context) ([]*Thread, error) {
	db := models.DB
	dbThreads := models.Threads{}
	currentUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)

	query := db.Q().LeftJoin("thread_participants tp", "threads.id = tp.thread_id")
	query = query.Where("tp.user_id = ?", currentUser.ID)
	if err := query.All(&dbThreads); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting threads: %v", err.Error()))
		return []*Thread{}, err
	}

	var gqlThreads []*Thread
	for _, dbThread := range dbThreads {
		newGqlThread, err := ConvertDBThreadToGqlThread(dbThread, []string{})

		if err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error converting threads: %v", err.Error()))
			return gqlThreads, err
		}
		gqlThreads = append(gqlThreads, &newGqlThread)

	}
	return gqlThreads, nil
}

func (r *queryResolver) Message(ctx context.Context, id *string) (*Message, error) {
	dbMsg := models.Message{}

	if err := models.DB.Where("uuid = ?", id).First(&dbMsg); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error getting message: %v", err.Error()))
		return &Message{}, err
	}

	gqlMessage, err := ConvertDBMessageToGqlMessage(dbMsg, GetRequestFields(ctx))
	if err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error converting message: %v", err.Error()))
		return &gqlMessage, err
	}

	return &gqlMessage, nil
}
