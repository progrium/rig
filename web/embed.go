package web

import "embed"

//go:embed system vscode index.html workbench.js wanix.min.js
var FS embed.FS
