package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
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

func DeleteSpecificGroups() error {
	_, err := DB.Exec(`
		DELETE FROM groups 
		WHERE title IN ('group one', 'test group', 'My group')
	`)
	if err != nil {
		return fmt.Errorf("failed to delete specific groups: %v", err)
	}
	
	log.Println("Successfully deleted specified groups")
	return nil
}

func PrintDatabaseContent() {
	// Print Groups
	rows, err := DB.Query("SELECT id, title, description, creator_id FROM groups")
	if err != nil {
		log.Printf("Error querying groups: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Groups ===")
	fmt.Printf("%-5s %-20s %-30s %-10s\n", "ID", "Title", "Description", "Creator")
	fmt.Println(strings.Repeat("-", 65))
	
	for rows.Next() {
		var id int
		var title, description string
		var creatorID int
		rows.Scan(&id, &title, &description, &creatorID)
		fmt.Printf("%-5d %-20s %-30s %-10d\n", id, title, description, creatorID)
	}

	// Print Users
	rows, err = DB.Query("SELECT id, username, email FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Users ===")
	fmt.Printf("%-5s %-20s %-30s\n", "ID", "Username", "Email")
	fmt.Println(strings.Repeat("-", 55))
	
	for rows.Next() {
		var id int
		var username, email string
		rows.Scan(&id, &username, &email)
		fmt.Printf("%-5d %-20s %-30s\n", id, username, email)
	}

	// Print Group Members
	rows, err = DB.Query(`
		SELECT gm.group_id, g.title, gm.user_id, u.username, gm.role
		FROM group_members gm
		JOIN groups g ON gm.group_id = g.id
		JOIN users u ON gm.user_id = u.id
	`)
	if err != nil {
		log.Printf("Error querying group members: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n=== Group Members ===")
	fmt.Printf("%-8s %-20s %-8s %-20s %-10s\n", "GroupID", "Group", "UserID", "Username", "Role")
	fmt.Println(strings.Repeat("-", 70))
	
	for rows.Next() {
		var groupID int
		var groupTitle string
		var userID int
		var username string
		var role string
		rows.Scan(&groupID, &groupTitle, &userID, &username, &role)
		fmt.Printf("%-8d %-20s %-8d %-20s %-10s\n", groupID, groupTitle, userID, username, role)
	}
}
