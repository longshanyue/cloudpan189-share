package embed

import (
	"embed"
	"io/fs"
)

//go:embed all:fe
var topFileFs embed.FS

func StaticFS() (fs.FS, bool) {
	feFs, err := fs.Sub(topFileFs, "fe")
	if err != nil {
		return nil, false
	}

	staticFS, err := fs.Sub(feFs, "dist")
	if err != nil {
		return nil, false
	}

	return staticFS, true
}
