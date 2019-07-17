package gqlgen

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Users(ctx context.Context) ([]*User, error) {
	db := models.DB
	dbUsers := models.Users{}

	if err := db.All(&dbUsers); err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error getting users: %v", err.Error()))
		return []*User{}, err
	}

	gqlUsers := []*User{}
	for _, dbUser := range dbUsers {
		newGqlUser, err := ConvertDBUserToGqlUser(dbUser, ctx)
		if err != nil {
			graphql.AddError(ctx, gqlerror.Errorf("Error converting users: %v", err.Error()))
			return gqlUsers, err
		}
		gqlUsers = append(gqlUsers, &newGqlUser)

	}
	return gqlUsers, nil
}


func (r *queryResolver) User(ctx context.Context, id *string) (*User, error) {
	panic("not implemented")
}
