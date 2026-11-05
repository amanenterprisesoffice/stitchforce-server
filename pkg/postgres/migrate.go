package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	embeddedmigrations "stitchforce-server/migrations"

	migrate "github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Migrate(db *sql.DB, migrationsPath string) error {
	// migrationsPath is retained for backward compatibility.
	// Migrations are now embedded in the binary.

	fmt.Println("Running embedded database migrations...")

	source, err := iofs.New(embeddedmigrations.FS, ".")
	if err != nil {
		return fmt.Errorf("create embedded migration source: %w", err)
	}
	defer source.Close()

	driver, err := migratepostgres.WithInstance(
		db,
		&migratepostgres.Config{},
	)
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	var host, dbName, schema string

	err = db.QueryRow(`
		SELECT
			current_database(),
			current_schema(),
			inet_server_addr()::text
	`).Scan(&dbName, &schema, &host)

	if err != nil {
		return fmt.Errorf("identify connected database: %w", err)
	}

	log.Println(
		"Connected database",
		"database", dbName,
		"schema", schema,
		"server", host,
	)

	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}

	err = m.Up()

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("Database schema is already up to date.")
		return nil
	}

	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	fmt.Println("Database migrations completed successfully.")

	return nil
}