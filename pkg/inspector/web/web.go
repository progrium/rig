package web

import (
	"embed"

	"github.com/progrium/rig/pkg/webfs"
	"tractor.dev/toolkit-go/engine/fs"
)

//go:embed *
var assets embed.FS

var Dir = webfs.New(fs.LiveDir(assets))
