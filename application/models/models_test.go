package models

import (
	"context"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/suite"

	"github.com/silinternational/wecarry-api/domain"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	model := suite.NewModel()

	as := &ModelSuite{
		Model: model,
	}
	suite.Run(t, as)
}

func createFixture(ms *ModelSuite, f interface{}) {
	err := ms.DB.Create(f)
	if err != nil {
		ms.T().Errorf("error creating %T fixture, %s", f, err)
		ms.T().FailNow()
	}
}

type testBuffaloContext struct {
	buffalo.DefaultContext
	params map[interface{}]interface{}
}

func (b *testBuffaloContext) Value(key interface{}) interface{} {
	return b.params[key]
}

func (b *testBuffaloContext) Set(key string, val interface{}) {
	b.params[key] = val
}

func createTestContext(user User) buffalo.Context {
	ctx := &testBuffaloContext{
		params: map[interface{}]interface{}{},
	}
	ctx.Set("current_user", user)
	return ctx
}

func (ms *ModelSuite) TestCurrentUser() {
	// setup
	user := createUserFixtures(ms.DB, 1).Users[0]
	ctx := createTestContext(user)

	tests := []struct {
		name     string
		context  context.Context
		wantUser User
	}{
		{
			name:     "buffalo context",
			context:  ctx,
			wantUser: user,
		},
		{
			name:     "gql context",
			context:  context.WithValue(ctx, domain.BuffaloContext, ctx),
			wantUser: user,
		},
		{
			name:     "empty context",
			context:  &testBuffaloContext{params: map[interface{}]interface{}{}},
			wantUser: User{},
		},
	}

	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			// execute
			got := CurrentUser(tt.context)

			// verify
			ms.Equal(tt.wantUser.ID, got.ID)
		})
	}

	// teardown
}

func (ms *ModelSuite) Test_removeImage() {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	files := createFileFixtures(1)
	users[1].PhotoFileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&users[1], "photo_file_id"))
	files[0].Linked = true
	ms.NoError(ms.DB.UpdateColumns(&files[0], "linked"))

	tests := []struct {
		name     string
		user     User
		oldImage *File
		want     File
		wantErr  string
	}{
		{
			name: "no file",
			user: users[0],
		},
		{
			name:     "has a file",
			user:     users[1],
			oldImage: &files[0],
		},
		{
			name:    "bad ID",
			user:    User{},
			wantErr: "invalid ID",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			err := removeImage(&tt.user)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.NoError(ms.DB.Reload(&tt.user))
			ms.False(tt.user.PhotoFileID.Valid, "image was not removed, %+v", tt.user.PhotoFileID)
			if tt.oldImage != nil {
				ms.NoError(ms.DB.Reload(tt.oldImage))
				ms.Equal(false, tt.oldImage.Linked, "old file is not marked as unlinked")
			}
		})
	}
}
