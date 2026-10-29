package main

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var staticFS embed.FS

func frontendFiles() fs.FS {
	files, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return files
}
