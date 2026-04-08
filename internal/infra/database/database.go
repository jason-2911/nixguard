// Package database provides SQLite/PostgreSQL database access for NixGuard.
package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the standard sql.DB with NixGuard-specific helpers.
type DB struct {
	*sql.DB
	log *slog.Logger
}

// Open creates a new database connection.
func Open(driver, dsn string, log *slog.Logger) (*DB, error) {
	// Ensure directory exists for SQLite
	if driver == "sqlite" || driver == "sqlite3" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
		driver = "sqlite3"
		dsn = dsn + "?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("database connected", slog.String("driver", driver))

	return &DB{DB: db, log: log}, nil
}

// Migrate runs all SQL migration files.
func (db *DB) Migrate(migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Create migrations tracking table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		// Check if already applied
		var count int
		db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", entry.Name()).Scan(&count)
		if count > 0 {
			continue
		}

		// Read and execute migration
		path := filepath.Join(migrationsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", entry.Name(), err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", entry.Name()); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}

		db.log.Info("migration applied", slog.String("file", entry.Name()))
	}

	return nil
}
