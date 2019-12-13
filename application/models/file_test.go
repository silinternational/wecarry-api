package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/validate"
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

	type args struct {
		name    string
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty file",
			args: args{
				name:    "empty",
				content: []byte{},
			},
			wantErr: true,
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
				name:    "large_file",
				content: make([]byte, domain.MaxFileSize+1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f File
			if err := f.Store(tt.args.name, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func CreateFileFixtures(ms *ModelSuite, posts Posts) Files {
	t := ms.T()
	const n = 2
	files := make(Files, n)

	for i := 0; i < n; i++ {
		var file File
		if err := file.Store(fmt.Sprintf("file_%d.gif", i), []byte("GIF87a")); err != nil {
			t.Errorf("failed to create file fixture %d, %s", i, err)
			t.FailNow()
		}
		files[i] = file
	}

	files[1].URLExpiration = time.Now().Add(-time.Minute)
	if err := ms.DB.Save(&files[1]); err != nil {
		t.Errorf("failed to update file fixture")
	}

	return files
}

func (ms *ModelSuite) TestFile_FindByUUID() {
	t := ms.T()

	uf := CreateUserFixtures(ms.DB, 2)
	posts := CreatePostFixtures(ms, t, uf.Users)

	if err := aws.CreateS3Bucket(); err != nil {
		t.Errorf("failed to create S3 bucket, %s", err)
		t.FailNow()
	}
	files := CreateFileFixtures(ms, posts)

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
			err := f.FindByUUID(tt.args.fileUUID)
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
			got, err := detectContentType(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectContentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectContentType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
