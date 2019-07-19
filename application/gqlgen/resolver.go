//go:generate go run github.com/99designs/gqlgen

package gqlgen

import (
	"context"

	"github.com/silinternational/handcarry-api/domain"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Users(ctx context.Context) ([]*User, error) {
	db := models.DB
	dbUsers := models.Users{}

	currentUser := domain.GetCurrentUserFromGqlContext(ctx)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin {
		return []*User{}, nil

	}

	if err := db.All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		return []*User{}, err
	}

	var gqlUsers []*User
	for _, dbUser := range dbUsers {
		newGqlUser, err := ConvertDBUserToGqlUser(dbUser)
		if err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error converting users: %v", err.Error()))
			return gqlUsers, err
		}
		gqlUsers = append(gqlUsers, &newGqlUser)

	}
	return gqlUsers, nil
}

func (r *queryResolver) User(ctx context.Context, id *string) (*User, error) {
	dbUser := models.User{}

	currentUser := domain.GetCurrentUserFromGqlContext(ctx)

	if currentUser.AdminRole.String != domain.AdminRoleSuperDuperAdmin && currentUser.Uuid != *id {
		return &User{}, nil
	}

	if err := models.DB.Where("uuid = ?", id).First(&dbUser); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting user: %v", err.Error()))
		return &User{}, err
	}

	newGqlUser, err := ConvertDBUserToGqlUser(dbUser)
	if err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error converting user: %v", err.Error()))
		return &newGqlUser, err
	}

	return &newGqlUser, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*Post, error) {
	db := models.DB
	dbPosts := models.Posts{}
	cUser := domain.GetCurrentUserFromGqlContext(ctx)
	if err := db.Where("organization_id IN (?)", cUser.GetOrgIDs()...).All(&dbPosts); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting posts: %v", err.Error()))
		return []*Post{}, err
	}

	var gqlPosts []*Post
	for _, dbPost := range dbPosts {
		newGqlPost, err := ConvertDBPostToGqlPost(dbPost, &cUser)
		if err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error converting posts: %v", err.Error()))
			return gqlPosts, err
		}
		gqlPosts = append(gqlPosts, &newGqlPost)

	}

	return gqlPosts, nil
}

func (r *queryResolver) Post(ctx context.Context, id *string) (*Post, error) {
	dbPost := models.Post{}
	cUser := domain.GetCurrentUserFromGqlContext(ctx)

	if err := models.DB.Where("organization_id IN (?)", cUser.GetOrgIDs()...).Where("uuid = ?", id).First(&dbPost); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting post: %v", err.Error()))
		return &Post{}, err
	}

	newGqlPost, err := ConvertDBPostToGqlPost(dbPost, &cUser)
	if err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error converting post: %v", err.Error()))
		return &newGqlPost, err
	}

	return &newGqlPost, nil
}

func (r *queryResolver) Threads(ctx context.Context) ([]*Thread, error) {
	db := models.DB
	dbThreads := models.Threads{}

	db = CallDBEagerWithRelatedFields(ThreadRelatedFields(), db, ctx)
	if err := db.All(&dbThreads); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting threads: %v", err.Error()))
		return []*Thread{}, err
	}

	var gqlThreads []*Thread
	for _, dbThread := range dbThreads {
		newGqlThread, err := ConvertDBThreadToGqlThread(dbThread)

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
	currentUser := domain.GetCurrentUserFromGqlContext(ctx)

	query := db.Q().LeftJoin("thread_participants tp", "threads.id = tp.thread_id")
	query = query.Where("tp.user_id = ?", currentUser.ID)
	if err := query.All(&dbThreads); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting threads: %v", err.Error()))
		return []*Thread{}, err
	}

	var gqlThreads []*Thread
	for _, dbThread := range dbThreads {
		newGqlThread, err := ConvertDBThreadToGqlThread(dbThread)

		if err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error converting users: %v", err.Error()))
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

	gqlMessage, err := ConvertDBMessageToGqlMessage(dbMsg)
	if err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("error converting message: %v", err.Error()))
		return &gqlMessage, err
	}

	return &gqlMessage, nil
}
