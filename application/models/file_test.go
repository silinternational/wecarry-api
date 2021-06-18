package models

import (
	"testing"
	"time"

	"github.com/gobuffalo/validate/v3"

	"github.com/silinternational/wecarry-api/aws"
	"github.com/silinternational/wecarry-api/domain"
)

func (ms *ModelSuite) TestFile_Validate() {
	t := ms.T()
	tests := []struct {
		name     string
		file     File
		want     *validate.Errors
		wantErr  bool
		errField string
	}{
		{
			name: "minimum",
			file: File{
				UUID: domain.GetUUID(),
			},
			wantErr: false,
		},
		{
			name:     "missing UUID",
			file:     File{},
			wantErr:  true,
			errField: "uuid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vErr, _ := test.file.Validate(DB)
			if test.wantErr {
				if vErr.Count() == 0 {
					t.Errorf("Expected an error, but did not get one")
				} else if len(vErr.Get(test.errField)) == 0 {
					t.Errorf("Expected an error on field %v, but got none (errors: %v)", test.errField, vErr.Errors)
				}
			} else if vErr.HasAny() {
				t.Errorf("Unexpected error: %v", vErr)
			}
		})
	}
}

func (ms *ModelSuite) TestFile_Store() {
	t := ms.T()

	// This is needed in for when this test is run on its own
	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}

	maxFileSize := domain.Megabyte * 10
	biggishGifFile := make([]byte, maxFileSize)
	tooBigFile := make([]byte, maxFileSize+1)
	for i, b := range []byte("GIF87a") {
		biggishGifFile[i] = b
		tooBigFile[i] = b
	}

	type args struct {
		name    string
		content []byte
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantCode string
	}{
		{
			name: "empty file",
			args: args{
				name:    "file0.gif",
				content: []byte{},
			},
			wantErr:  true,
			wantCode: "ErrorStoreFileBadContentType",
		},
		{
			name: "GIF87a file",
			args: args{
				name:    "file.gif",
				content: []byte("GIF87a"),
			},
			wantErr: false,
		},
		{
			name: "large file",
			args: args{
				name:    "file2.gif",
				content: tooBigFile,
			},
			wantErr:  true,
			wantCode: "ErrorStoreFileTooLarge",
		},
		{
			name: "just small enough GIF87a file",
			args: args{
				name:    "file3.gif",
				content: biggishGifFile,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := File{
				Name:    tt.args.name,
				Content: tt.args.content,
			}
			fErr := f.Store(ms.DB)
			if tt.wantErr {
				ms.NotNil(fErr)
				ms.Equal(fErr.ErrorCode, tt.wantCode, "incorrect error type")
				return
			}

			ms.Nil(fErr, "unexpected error")
		})
	}
}

func (ms *ModelSuite) TestFile_FindByUUID() {
	t := ms.T()

	_ = createUserFixtures(ms.DB, 2)

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}
	files := createFileFixtures(ms.DB, 2)

	type args struct {
		fileUUID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "good",
			args: args{
				fileUUID: files[0].UUID.String(),
			},
		},
		{
			name: "needs refresh",
			args: args{
				fileUUID: files[1].UUID.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f File
			err := f.FindByUUID(ms.DB, tt.args.fileUUID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("error = %v, fileID = %s", err, tt.args.fileUUID)
				} else {
					ms.Equal(tt.args.fileUUID, f.UUID.String(), "retrieved file has wrong UUID")
					ms.Contains(f.URL, "http", "URL doesn't start with 'http'")
					ms.True(f.URLExpiration.After(time.Now()), "URLExpiration is in the past")
				}
			}
		})
	}
}

func (ms *ModelSuite) TestFiles_FindByIDs() {
	t := ms.T()

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}
	files := createFileFixtures(ms.DB, 2)

	tests := []struct {
		name string
		ids  []int
		want []string
	}{
		{
			name: "good",
			ids:  []int{files[0].ID, files[1].ID, files[0].ID},
			want: []string{files[0].Name, files[1].Name},
		},
		{
			name: "missing",
			ids:  []int{99999},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Files
			err := f.FindByIDs(ms.DB, tt.ids)
			ms.NoError(err)

			fileNames := make([]string, len(f))
			for i, ff := range f {
				fileNames[i] = ff.Name
			}
			ms.ElementsMatch(tt.want, fileNames, "incorrect file names")
		})
	}
}

func (ms *ModelSuite) Test_detectContentType() {
	t := ms.T()
	tests := []struct {
		name    string
		content []byte
		want    string
		wantErr bool
	}{
		{
			name:    "BMP",
			content: []byte("BM"),
			want:    "image/bmp",
		},
		{
			name:    "GIF87a",
			content: []byte("GIF87a"),
			want:    "image/gif",
		},
		{
			name:    "GIF89a",
			content: []byte("GIF89a"),
			want:    "image/gif",
		},
		{
			name:    "WebP",
			content: []byte("RIFFxxxxWEBPVP"),
			want:    "image/webp",
		},
		{
			name:    "PNG",
			content: []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a},
			want:    "image/png",
		},
		{
			name:    "JPEG",
			content: []byte{0xff, 0xd8, 0xff},
			want:    "image/jpeg",
		},
		{
			name:    "pdf",
			content: []byte("%PDF-"),
			want:    "application/pdf",
		},
		{
			name:    "GZIP",
			content: []byte{0x1f, 0x8b, 0x08},
			wantErr: true,
		},
		{
			name:    "ZIP",
			content: []byte{0x50, 0x4b, 0x03, 0x04},
			wantErr: true,
		},
		{
			name:    "EXE", // detected as application/octet-stream
			content: []byte{0x4d, 0x5a, 0x00},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateContentType(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateContentType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func (ms *ModelSuite) TestFiles_DeleteUnlinked() {
	const (
		nOldUnlinkedFiles = 2
		nNewUnlinkedFiles = 2
		nRequests         = 2
		nMeetings         = 2
		nOrganizations    = 2
		nUsers            = 2
	)

	_ = createFileFixtures(ms.DB, nOldUnlinkedFiles)

	ms.NoError(DB.RawQuery("UPDATE files set updated_at = ?", time.Now().Add(-5*domain.DurationWeek)).Exec())

	_ = createFileFixtures(ms.DB, nNewUnlinkedFiles)

	requests := createRequestFixtures(ms.DB, nRequests, true)

	requestFiles := createFileFixtures(ms.DB, nRequests)
	for i, r := range requestFiles {
		_, err := requests[i].AttachFile(ms.DB, r.UUID.String())
		ms.NoError(err)
	}

	_ = createMeetingFixtures(ms.DB, nMeetings)

	_ = createOrganizationFixtures(ms.DB, nOrganizations)

	users := createUserFixtures(ms.DB, nUsers).Users
	userPhotos := createFileFixtures(ms.DB, nUsers)
	for i, u := range users {
		_, err := u.AttachPhoto(ms.DB, userPhotos[i].UUID.String())
		ms.NoError(err)
	}

	f := Files{}

	domain.Env.MaxFileDelete = 1
	ms.Error(f.DeleteUnlinked(ms.DB))

	domain.Env.MaxFileDelete = 2
	ms.NoError(f.DeleteUnlinked(ms.DB))
	n, _ := DB.Count(&f)
	ms.Equal(nRequests*2+nMeetings+nOrganizations+nUsers+nNewUnlinkedFiles, n, "wrong number of files remain")
}

func (ms *ModelSuite) Test_changeFileExtension() {
	tests := []struct {
		name        string
		filename    string
		contentType string
		want        string
	}{
		{
			name:        "png to gif",
			filename:    "file.png",
			contentType: "image/gif",
			want:        "file.gif",
		},
		{
			name:        "webp to png",
			filename:    "file.webp",
			contentType: "image/png",
			want:        "file.png",
		},
		{
			name:        "bad type",
			filename:    "file.webp",
			contentType: "file/xyz",
			want:        "file.webp",
		},
	}
	for _, tt := range tests {
		ms.T().Run(tt.name, func(t *testing.T) {
			f := File{Name: tt.filename, ContentType: tt.contentType}
			f.changeFileExtension()
			ms.Equal(tt.want, f.Name)
		})
	}
}
