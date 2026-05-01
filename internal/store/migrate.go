package store

import (
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationSQL embed.FS

func (s *Store) migrate() error {
	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
)`); err != nil {
		return fmt.Errorf("schema_migrations table: %w", err)
	}

	entries, err := migrationSQL.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		ver, err := parseMigrationVersion(name)
		if err != nil {
			return err
		}
		var n int
		if err := s.db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, ver).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}

		body, err := migrationSQL.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, ver, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		slog.Info("migration applied", "version", ver, "name", name)
	}
	return nil
}

func parseMigrationVersion(filename string) (int, error) {
	base := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) < 1 || parts[0] == "" {
		return 0, fmt.Errorf("migration name must be like 001_description.sql, got %q", filename)
	}
	return strconv.Atoi(parts[0])
}
