package webdist

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// StaticFS returns the filesystem rooted at the Next static export tree (dist/).
func StaticFS() (fs.FS, error) {
	return fs.Sub(embedded, "dist")
}
