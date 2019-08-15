package gqlgen

import (
	"context"
	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/models"
)

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	post, err := ConvertGqlNewPostToDBPost(input, cUser)
	if err != nil {
		return &models.Post{}, err
	}

	if err := models.DB.Create(&post); err != nil {
		return &models.Post{}, err
	}

	return &post, err
}

func (r *mutationResolver) UpdatePostStatus(ctx context.Context, input UpdatedPostStatus) (*models.Post, error) {
	post, err := models.FindPostByUUID(input.ID)
	if err != nil {
		return &models.Post{}, err
	}

	post.Status = input.Status
	post.ProviderID = nulls.NewInt(models.GetCurrentUserFromGqlContext(ctx, TestUser).ID)
	if err := models.DB.Update(&post); err != nil {
		return &models.Post{}, err
	}

	return &post, nil
}

func (r *mutationResolver) CreateMessage(ctx context.Context, input NewMessage) (*models.Message, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	message, err := ConvertGqlNewMessageToDBMessage(input, cUser)
	if err != nil {
		return &models.Message{}, err
	}

	if err := models.DB.Create(&message); err != nil {
		return &models.Message{}, err
	}

	return &message, err
}
