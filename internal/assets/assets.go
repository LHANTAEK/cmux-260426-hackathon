package assets

import "embed"

// Harness contains the project-local Claude Code plugin and harness files.
//
//go:embed all:templates
var Harness embed.FS
