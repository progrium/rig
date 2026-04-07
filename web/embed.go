package web

import "embed"

//go:embed editors system vscode index.html workbench.js wanix.min.js
var FS embed.FS
