package templates

import (
	"embed"
	"io/fs"

	"github.com/gobuffalo/buffalo"
)

//go:embed * */* mail/*
var files embed.FS

func FS() fs.FS {
	return buffalo.NewFS(files, "templates")
}
