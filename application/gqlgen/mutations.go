package gqlgen

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gobuffalo/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
)

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreatePost(ctx context.Context, input NewPost) (*models.Post, error) {
	cUser := models.GetCurrentUserFromGqlContext(ctx, TestUser)
	post, err := ConvertGqlNewPostToDBPost(input, cUser)
	if err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	if err := models.DB.Create(&post); err != nil {
		return &models.Post{}, err
	}

	return &post, err
}

func (r *mutationResolver) UpdatePost(ctx context.Context, input UpdatedPost) (*models.Post, error) {
	var post models.Post
	if err := post.FindByUUID(input.ID); err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}

	post.Status = input.Status.String()
	post.ProviderID = nulls.NewInt(models.GetCurrentUserFromGqlContext(ctx, TestUser).ID)
	if err := models.DB.Update(&post); err != nil {
		domain.Error(models.GetBuffaloContextFromGqlContext(ctx), err.Error(), domain.NoExtras)
		return &models.Post{}, err
	}
	if input.Title != nil {
		post.Title = *input.Title
	}
	if input.Description != nil {
		post.Description = models.ConvertStringPtrToNullsString(input.Description)
	}
	if input.Destination != nil {
		post.Destination = models.ConvertStringPtrToNullsString(input.Destination)
	}
	if input.Origin != nil {
		post.Origin = models.ConvertStringPtrToNullsString(input.Origin)
	}
	if input.Size != nil {
		post.Size = *input.Size
	}
	if input.NeededAfter != nil {
		neededAfter, err := domain.ConvertStringPtrToDate(input.NeededAfter)
		if err != nil {
			err = fmt.Errorf("error converting NeededAfter %v ... %v", input.NeededAfter, err.Error())
			return &models.Post{}, err
		}
		post.NeededAfter = neededAfter
	}
	if input.NeededBefore != nil {
		neededBefore, err := domain.ConvertStringPtrToDate(input.NeededBefore)
		if err != nil {
			err = fmt.Errorf("error converting NeededBefore %v ... %v", input.NeededBefore, err.Error())
			return &models.Post{}, err
		}
		post.NeededBefore = neededBefore
	}
	if input.Category != nil {
		post.Category = *input.Category
	}
	if input.URL != nil {
		post.URL = nulls.NewString(*input.URL)
	}
	if input.Cost != nil {
		c, err := strconv.ParseFloat(*input.Cost, 64)
		if err != nil {
			err = fmt.Errorf("error converting cost %v ... %v", input.Cost, err.Error())
			return &models.Post{}, err
		}
		post.Cost = nulls.NewFloat64(c)
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
