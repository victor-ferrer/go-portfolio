package webui

import "embed"

// StaticFiles holds the embedded web UI static files.
//
//go:embed dist
var StaticFiles embed.FS
