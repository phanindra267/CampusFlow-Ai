// Package migrations embeds the SQL schema files so that the migration CLI and
// any tooling that needs the schema share a single source of truth.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS