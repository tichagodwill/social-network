package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
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

	return nil
}
