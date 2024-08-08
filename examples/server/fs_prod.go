//go:build prod

package main

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embedFS embed.FS

// Get Frontend file system
func getFrontendDistFs() fs.FS {
	f, err := fs.Sub(embedFS, "dist")
	if err != nil {
		panic(err)
	}

	return f
}
