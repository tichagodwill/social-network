package sqlite

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// RunMigrations executes all up migrations in order
func RunMigrations() error {
    // Get the current working directory
    currentDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %v", err)
    }

    // Path to migrations directory
    migrationsPath := filepath.Join(currentDir, "pkg", "db", "migrations", "sqlite")

    // Read all files in the migrations directory
    files, err := os.ReadDir(migrationsPath)
    if err != nil {
        return fmt.Errorf("failed to read migrations directory: %v", err)
    }

    // Filter and sort up migrations
    var upMigrations []string
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".up.sql") {
            upMigrations = append(upMigrations, file.Name())
        }
    }
    sort.Strings(upMigrations)

    // Execute each migration
    for _, fileName := range upMigrations {
        log.Printf("Running migration: %s", fileName)
        
        // Read migration file
        content, err := os.ReadFile(filepath.Join(migrationsPath, fileName))
        if err != nil {
            return fmt.Errorf("failed to read migration file %s: %v", fileName, err)
        }

        // Execute migration
        _, err = DB.Exec(string(content))
        if err != nil {
            // Check if error is about table already existing
            if strings.Contains(err.Error(), "already exists") {
                log.Printf("Table already exists in %s, continuing...", fileName)
                continue
            }
            return fmt.Errorf("failed to execute migration %s: %v", fileName, err)
        }
    }

    log.Println("All migrations completed successfully")
    return nil
}

// RollbackMigrations executes all down migrations in reverse order
func RollbackMigrations() error {
    // Get the current working directory
    currentDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %v", err)
    }

    // Path to migrations directory
    migrationsPath := filepath.Join(currentDir, "pkg", "db", "migrations", "sqlite")

    // Read all files in the migrations directory
    files, err := os.ReadDir(migrationsPath)
    if err != nil {
        return fmt.Errorf("failed to read migrations directory: %v", err)
    }

    // Filter and sort down migrations in reverse order
    var downMigrations []string
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".down.sql") {
            downMigrations = append(downMigrations, file.Name())
        }
    }
    sort.Sort(sort.Reverse(sort.StringSlice(downMigrations)))

    // Execute each rollback
    for _, fileName := range downMigrations {
        log.Printf("Rolling back: %s", fileName)
        
        // Read migration file
        content, err := os.ReadFile(filepath.Join(migrationsPath, fileName))
        if err != nil {
            return fmt.Errorf("failed to read migration file %s: %v", fileName, err)
        }

        // Execute rollback
        _, err = DB.Exec(string(content))
        if err != nil {
            // Check if error is about table not existing
            if strings.Contains(err.Error(), "no such table") {
                log.Printf("Table already dropped in %s, continuing...", fileName)
                continue
            }
            return fmt.Errorf("failed to execute rollback %s: %v", fileName, err)
        }
    }

    log.Println("All rollbacks completed successfully")
    return nil
}

// ClearDatabase drops and recreates all tables
func ClearDatabase() error {
    // First rollback all migrations
    if err := RollbackMigrations(); err != nil {
        return fmt.Errorf("failed to rollback migrations: %v", err)
    }

    // Then run them again
    if err := RunMigrations(); err != nil {
        return fmt.Errorf("failed to run migrations: %v", err)
    }

    return nil
} 