package migrations

import (
	"embed"
	"io/fs"
)

//go:embed sql/*.sql
var EmbeddedMigrations embed.FS

func GetMigrations() (fs.FS, error) {
	return fs.Sub(EmbeddedMigrations, "sql")
}
