package gqlgen

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gobuffalo/pop/nulls"
	"github.com/silinternational/handcarry-api/domain"
	"github.com/silinternational/handcarry-api/models"
	"github.com/vektah/gqlparser/gqlerror"
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

	gqlPost, err := ConvertDBPostToGqlPost(dbPost, &cUser)

	return &gqlPost, err
}

func (r *mutationResolver) UpdatePostStatus(ctx context.Context, input UpdatedPostStatus) (*Post, error) {
	post, err := models.FindPostByUUID(input.ID)
	if err != nil {
		return &Post{}, err
	}

	post.Status = input.Status
	post.ProviderID = nulls.NewInt(domain.GetCurrentUserFromGqlContext(ctx).ID)
	if err := models.DB.Update(&post); err != nil {
		return &Post{}, err
	}

	updatedPost, err := ConvertDBPostToGqlPost(post, nil)
	if err != nil {
		graphql.AddError(ctx, gqlerror.Errorf("Error converting post: %v", err.Error()))
		return &updatedPost, err
	}

	return &updatedPost, nil
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
