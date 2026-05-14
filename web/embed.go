package web

import (
	"embed"

	"tractor.dev/toolkit-go/engine/fs"
)

//go:embed editors system vscode index.html workbench.js wanix.min.js
var embedded embed.FS

var FS = fs.LiveDir(embedded)
