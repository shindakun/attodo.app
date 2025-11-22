package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps sql.DB with our custom methods
type DB struct {
	*sql.DB
}

// New creates a new database connection and runs all pending migrations
// dbPath: path to the SQLite database file (e.g., "./data/notifications.db")
// migrationsDir: path to directory containing migration SQL files (e.g., "./migrations")
func New(dbPath string, migrationsDir string) (*DB, error) {
	// Ensure the directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Run migrations automatically
	if err := runMigrations(db, migrationsDir); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{db}, nil
}

// runMigrations applies all pending SQL migration files from the migrations directory
// Migration files are named like: 0001-description.sql, 0002-another.sql, etc.
// They are applied in alphabetical order and tracked in schema_migrations table
func runMigrations(db *sql.DB, migrationsDir string) error {
	// Create migrations tracking table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of already applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading migration rows: %w", err)
	}

	// Read migration files from directory
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		// If migrations directory doesn't exist, that's OK - no migrations to run
		if os.IsNotExist(err) {
			log.Printf("Migrations directory %s does not exist, skipping migrations", migrationsDir)
			return nil
		}
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, filename := range migrationFiles {
		// Extract version from filename (e.g., "0001-initial-db-setup.sql" -> "0001")
		parts := strings.Split(filename, "-")
		if len(parts) == 0 {
			log.Printf("Skipping migration file with invalid name: %s", filename)
			continue
		}
		version := parts[0]

		if appliedMigrations[version] {
			continue // Already applied
		}

		// Read migration file
		migrationPath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", filename, err)
		}

		// Execute the migration SQL
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record migration as applied
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("Applied migration: %s", filename)
	}

	return nil
}
