//go:build !prod

package main

import (
	"io/fs"
	"os"
)


// Get Frontend file system
func getFrontendDistFs() fs.FS {
	return os.DirFS("dist")
}

