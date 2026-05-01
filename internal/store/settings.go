package store

import (
	"database/sql"
	"strconv"
	"time"
)

func (s *Store) GetSetting(key string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	return v, err
}

func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

func (s *Store) GetAllSettings() (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func (s *Store) UpdateSettings(updates map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for k, v := range updates {
		if _, err := tx.Exec(`
			INSERT INTO settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) LastPostUpdate() (*time.Time, error) {
	var raw sql.NullString
	err := s.db.QueryRow(`SELECT MAX(updated_at) FROM posts`).Scan(&raw)
	if err != nil {
		return nil, err
	}
	if !raw.Valid || raw.String == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, raw.String)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", raw.String)
	}
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05-07:00", raw.String)
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) PostsPerPageFromSettings(defaultN int) int {
	v, err := s.GetSetting("posts_per_page")
	if err != nil {
		return defaultN
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultN
	}
	return n
}
