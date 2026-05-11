package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ApplyMigrations(db *sql.DB) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		applied, err := hasAppliedMigration(db, entry.Name())
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := applyMigrationFile(db, filepath.Join("migrations", entry.Name())); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func hasAppliedMigration(db *sql.DB, name string) (bool, error) {
	hasTable, err := hasMigrationsTable(db)
	if err != nil {
		return false, err
	}
	if !hasTable {
		return false, nil
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", name).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func hasMigrationsTable(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = 'migrations'
	`).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func applyMigrationFile(db *sql.DB, path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	for _, statement := range splitSQLStatements(string(contents)) {
		if _, err := transaction.Exec(statement); err != nil {
			transaction.Rollback()
			return err
		}
	}

	return transaction.Commit()
}

func splitSQLStatements(contents string) []string {
	parts := strings.Split(contents, ";")
	statements := make([]string, 0, len(parts))

	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement == "" {
			continue
		}

		statements = append(statements, statement)
	}

	return statements
}
