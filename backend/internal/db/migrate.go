package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func Migrate(conn *sql.DB, migrationsDir string) error {
	if err := ensureSchemaMigrations(conn); err != nil {
		return err
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		applied, err := migrationApplied(conn, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		path := filepath.Join(migrationsDir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := conn.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if err := markMigrationApplied(conn, name); err != nil {
			return err
		}
	}
	return nil
}

func ensureSchemaMigrations(conn *sql.DB) error {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(64) PRIMARY KEY,
			applied_at DATETIME NOT NULL
		)
	`)
	return err
}

func migrationApplied(conn *sql.DB, version string) (bool, error) {
	var existing string
	err := conn.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", version).Scan(&existing)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func markMigrationApplied(conn *sql.DB, version string) error {
	_, err := conn.Exec(
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		version,
		time.Now().UTC(),
	)
	return err
}
