package gqlgen

import (
	"context"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (*Post, error) {
	cUser := domain.GetCurrentUserFromGqlContext(ctx)
	dbPost, err := ConvertGqlNewPostToDBPost(input, cUser)
	if err != nil {
		return &Post{}, err
	}

	if err := models.DB.Create(&dbPost); err != nil {
		return &Post{}, err
	}

	gqlPost, err := ConvertDBPostToGqlPost(dbPost)

	return &gqlPost, err
}

func (r *mutationResolver) UpdatePost(ctx context.Context, input UpdatedPostStatus) (*Post, error) {
	panic("UpdatePost not implemented")
}

func (r *mutationResolver) CreateMessage(ctx context.Context, input NewMessage) (*Message, error) {
	cUser := domain.GetCurrentUserFromGqlContext(ctx)
	dbMessage, err := ConvertGqlNewMessageToDBMessage(input, cUser)
	if err != nil {
		return &Message{}, err
	}

	if err := models.DB.Create(&dbMessage); err != nil {
		return &Message{}, err
	}

	gqlMessage, err := ConvertDBMessageToGqlMessage(dbMessage)

	return &gqlMessage, err
}
