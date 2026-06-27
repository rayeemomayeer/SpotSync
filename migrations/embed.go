// Package migrations holds versioned SQL schema files embedded for runtime migrate.
package migrations

import "embed"

// Files contains golang-migrate up/down SQL scripts.
//
//go:embed *.sql
var Files embed.FS
