package models

import (
	"context"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/silinternational/wecarry-api/domain"
)

// ModelSuite doesn't contain a buffalo suite.Model and can be used for tests that don't need access to the database
// or don't need the buffalo test runner to refresh the database
type ModelSuite struct {
	suite.Suite
	*require.Assertions
	DB *pop.Connection
}

func (ms *ModelSuite) SetupTest() {
	ms.Assertions = require.New(ms.T())
	DestroyAll()
}

// Test_ModelSuite runs the test suite
func Test_ModelSuite(t *testing.T) {
	ms := &ModelSuite{}
	c, err := pop.Connect(envy.Get("GO_ENV", "test"))
	if err == nil {
		ms.DB = c
	}
	suite.Run(t, ms)
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

func CtxWithUser(user User) buffalo.Context {
	ctx := &testBuffaloContext{
		params: map[interface{}]interface{}{},
	}
	ctx.Set(domain.ContextKeyCurrentUser, user)
	return ctx
}

func (ms *ModelSuite) TestCurrentUser() {
	// setup
	user := createUserFixtures(ms.DB, 1).Users[0]
	ctx := CtxWithUser(user)

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

func (ms *ModelSuite) Test_addFile() {
	uf := createUserFixtures(ms.DB, 3)
	users := uf.Users

	files := createFileFixtures(ms.DB, 3)
	users[1].FileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&users[1], "file_id"))

	tests := []struct {
		name     string
		user     User
		oldImage *File
		newImage string
		want     File
		wantErr  string
	}{
		{
			name:     "no previous file",
			user:     users[0],
			newImage: files[1].UUID.String(),
			want:     files[1],
		},
		{
			name:     "previous file",
			user:     users[1],
			oldImage: &files[0],
			newImage: files[2].UUID.String(),
			want:     files[2],
		},
		{
			name:     "bad ID",
			user:     users[2],
			newImage: uuid.UUID{}.String(),
			wantErr:  "no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := addFile(ms.DB, &tt.user, tt.newImage)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(tt.want.UUID.String(), got.UUID.String(), "wrong file returned")
			ms.NoError(ms.DB.Reload(&tt.user))
			ms.Equal(tt.want.ID, tt.user.FileID.Int, "user file ID isn't correct")
			ms.Equal(true, got.Linked, "new user photo file is not marked as linked")
			if tt.oldImage != nil {
				ms.NoError(ms.DB.Reload(tt.oldImage))
				ms.Equal(false, tt.oldImage.Linked, "old user photo file is not marked as unlinked")
			}
		})
	}
}

func (ms *ModelSuite) Test_removeFile() {
	uf := createUserFixtures(ms.DB, 2)
	users := uf.Users

	files := createFileFixtures(ms.DB, 1)
	users[1].FileID = nulls.NewInt(files[0].ID)
	ms.NoError(ms.DB.UpdateColumns(&users[1], "file_id"))
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
			err := removeFile(ms.DB, &tt.user)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.NoError(ms.DB.Reload(&tt.user))
			ms.False(tt.user.FileID.Valid, "image was not removed, %+v", tt.user.FileID)
			if tt.oldImage != nil {
				ms.NoError(ms.DB.Reload(tt.oldImage))
				ms.Equal(false, tt.oldImage.Linked, "old file is not marked as unlinked")
			}
		})
	}
}
