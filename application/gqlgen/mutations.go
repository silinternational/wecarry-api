package gqlgen

import (
	"context"

	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

type mutationResolver struct{ *Resolver }

type postInput struct {
	ID           *string
	Status       *PostStatus
	OrgID        *string
	Type         *PostType
	Title        *string
	Description  *string
	Destination  *string
	Origin       *string
	Size         *string
	NeededAfter  *string
	NeededBefore *string
	Category     *string
	URL          *string
	Cost         *string
}

func (r *mutationResolver) CreatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	post, err := convertGqlPostInputToDBPost(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	if err := models.DB.Create(&post); err != nil {
		return &models.Post{}, err
	}

	return &post, nil
}

func (r *mutationResolver) UpdatePost(ctx context.Context, input postInput) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	post, err := convertGqlPostInputToDBPost(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	if err := models.DB.Update(&post); err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	return &post, nil
}

func (r *mutationResolver) CreateMessage(ctx context.Context, input NewMessage) (*models.Message, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	message, err := ConvertGqlNewMessageToDBMessage(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Message{}, err
	}

	if err := models.DB.Create(&message); err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Message{}, err
	}

	return &message, err
}
