package sqlite

import (
	"database/sql"
	"log"
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

	if err := RunMigrations(); err != nil {
		log.Fatalf("Migration Error: %v", err)
	}

	return nil
}
