package models

import (
	"testing"

	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
)

type PostFileFixtures struct {
	Posts
	Files
	PostFiles
}

func createPostFileFixtures(tx *pop.Connection, n int) PostFileFixtures {
	posts := createPostFixtures(tx, 1, 0, false)
	files := createFileFixtures(n)
	pf := make(PostFiles, n)
	for i := range pf {
		pf[i].FileID = files[i].ID
		pf[i].PostID = posts[0].ID
		mustCreate(tx, &pf[i])
	}
	return PostFileFixtures{
		Posts:     posts,
		Files:     files,
		PostFiles: pf,
	}
}

func (ms *ModelSuite) TestPostFile_AttachFile() {
	f := createPostFileFixtures(ms.DB, 2)
	files := f.Files
	newFile := createFileFixtures(1)[0]
	postfiles := f.PostFiles

	tests := []struct {
		name     string
		postfile PostFile
		oldFile  *File
		newFile  string
		want     File
		wantErr  string
	}{
		{
			name:     "new file",
			postfile: postfiles[0],
			oldFile:  &files[0],
			newFile:  newFile.UUID.String(),
			want:     newFile,
		},
		{
			name:     "bad ID",
			postfile: postfiles[1],
			newFile:  uuid.UUID{}.String(),
			wantErr:  "no rows in result set",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.postfile.AttachFile(tt.newFile)
			if tt.wantErr != "" {
				ms.Error(err, "did not get expected error")
				ms.Contains(err.Error(), tt.wantErr)
				return
			}
			ms.NoError(err, "unexpected error")
			ms.Equal(tt.want.UUID.String(), got.UUID.String(), "wrong file returned")
			ms.Equal(true, got.Linked, "new file is not marked as linked")
			if tt.oldFile != nil {
				ms.Equal(false, tt.oldFile.Linked, "old file is not marked as unlinked")
			}
		})
	}
}
