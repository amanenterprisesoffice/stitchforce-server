package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	migrate "github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(db *sql.DB, migrationsPath string) error {
	absolutePath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migration path: %w", err)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		return fmt.Errorf("migration directory %q: %w", absolutePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("migration path %q is not a directory", absolutePath)
	}

	fmt.Printf("Running database migrations from: %s\n", absolutePath)

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(absolutePath),
	}).String()

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
		log.Println("failed to identify database", "error", err)
		return nil
	}

	log.Println("connected database",
		"database", dbName,
		"schema", schema,
		"server", host,
	)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		_ = driver.Close()
		return fmt.Errorf("initialize migrations: %w", err)
	}
// 	defer func() {
//     sourceErr, databaseErr := m.Close()

//     if sourceErr != nil {
//         log.Println("migration source close error:", sourceErr)
//     }
//     if databaseErr != nil {
//         log.Println("migration database close error:", databaseErr)
//     }
// }()

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
