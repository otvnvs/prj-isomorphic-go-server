// Package assets_embedded embeds static files into the binary at compile time.
// Any file added to this directory is automatically included — no code changes
// needed. Access files via FS.ReadFile("name") or serve the whole directory
// with http.FileServerFS(assets_embedded.FS).
package assets_embedded

import "embed"

//go:embed *
var FS embed.FS
