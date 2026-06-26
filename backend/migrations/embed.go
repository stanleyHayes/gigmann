// Package migrations embeds the SQL schema migrations so the binary can apply
// them at startup without shipping the .sql files separately. Only *.up.sql files
// are embedded (and applied); *.down.sql files are kept for manual rollback.
package migrations

import "embed"

// Files holds the embedded up-migrations, applied in filename order.
//
//go:embed *.up.sql
var Files embed.FS
