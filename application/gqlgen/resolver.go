package gqlgen

import (
	"context"
	"log"

	"github.com/silinternational/handcarry-api/domain"

	"github.com/gobuffalo/buffalo"

	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func GetBuffaloContext(ctx context.Context) buffalo.Context {
	return ctx.Value("BuffaloContext").(buffalo.Context)
}

func (r *queryResolver) Users(ctx context.Context) ([]*User, error) {
	db := models.DB
	dbUsers := models.Users{}

	if err := db.All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		return []*User{}, err
	}

	gqlUsers := []*User{}
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
	buffaloContext := GetBuffaloContext(ctx)
	currentUser := domain.GetCurrentUser(buffaloContext)
	log.Printf("\n\nresolver current user: %+v\n\n", currentUser)
	// orgIds := currentUser.GetOrgIDs()
	// log.Printf("\n\norg ids: %v", orgIds)
	if err := db.Where("org_id IN (?)", currentUser.AuthOrgID).All(&dbPosts); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting posts: %v", err.Error()))
		return []*Post{}, err
	}

	var gqlPosts []*Post
	for _, dbPost := range dbPosts {
		newGqlPost, err := ConvertDBPostToGqlPost(dbPost)
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

	if err := models.DB.Where("uuid = ?", id).First(&dbPost); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting post: %v", err.Error()))
		return &Post{}, err
	}

	newGqlPost, err := ConvertDBPostToGqlPost(dbPost)
	if err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error converting post: %v", err.Error()))
		return &newGqlPost, err
	}

	return &newGqlPost, nil
}
