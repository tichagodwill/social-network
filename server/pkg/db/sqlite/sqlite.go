package sqlite

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func OpenDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// test the database connection
	if err := DB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to database")

	if err := runMigrations(); err != nil {
		log.Fatalf("Migration Error: %v", err)
	}

	return nil
}

func runMigrations() error {
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs("/pkg/db/migrations/sqlite")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file:/"+absPath, "sqlite3", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migration applied succesffuly")
	return nil
}
