package database

import (
	"context"
	"database/sql"
	"embed"
	"log"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(db *sql.DB, ctx context.Context) error {
	log.Println("Running migrations...")

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		content, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}

		_, err = db.ExecContext(ctx, string(content))
		if err != nil {
			return err
		}
	}

	log.Println("Migrations finished")
	return nil
}
