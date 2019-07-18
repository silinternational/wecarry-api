package gqlgen

import (
	"context"
	"github.com/silinternational/handcarry-api/models"
)

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (*Post, error) {
	dbPost, err := ConvertGqlNewPostToDBPost(input)
	if err != nil {
		return &Post{}, err
	}

	if err := models.DB.Create(&dbPost); err != nil {
		return &Post{}, err
	}

	gqlPost, err := ConvertDBPostToGqlPost(dbPost)

	return &gqlPost, err
}
