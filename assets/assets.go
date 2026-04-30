// Package assets embeds all static files into the binary at compile time.
// Any file added to this directory is automatically included — no code changes
// needed. Access files via FS.ReadFile("name") or serve the whole directory
// with http.FileServerFS(assets.FS).
package assets

import "embed"

//go:embed *
var FS embed.FS
