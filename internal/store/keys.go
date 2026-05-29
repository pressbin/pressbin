package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jaevor/go-nanoid"
	"golang.org/x/crypto/bcrypt"
)

type APIKey struct {
	ID          string
	Label       string
	Prefix      string
	Hash        string
	Permissions []string
	CreatedAt   time.Time
	LastUsedAt  *time.Time
}

func (k APIKey) HasPermission(required string) bool {
	for _, p := range k.Permissions {
		if p == required {
			return true
		}
	}
	return false
}

func (s *Store) countKeys() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM api_keys`).Scan(&n)
	return n, err
}

// Bootstrap creates the default admin key when no keys exist (legacy/dev).
func (s *Store) Bootstrap() (string, error) {
	n, err := s.countKeys()
	if err != nil {
		return "", err
	}
	if n > 0 {
		return "", nil
	}
	return s.CreateAdminKey("default admin")
}

func (s *Store) HasKeysWithPrefix(prefix string) (bool, error) {
	keys, err := s.getKeysByPrefix(prefix)
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

func (s *Store) CreateAdminKey(label string) (string, error) {
	return s.createKey("pb_admin_", label, []string{"admin"})
}

func (s *Store) CreateSyncKey(label string) (string, error) {
	return s.createKey("pb_sync_", label, []string{"posts:write"})
}

func (s *Store) createKey(prefix, label string, perms []string) (string, error) {
	raw := prefix + RandomKeySuffix(32)
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	idGen, err := nanoid.Standard(21)
	if err != nil {
		return "", err
	}
	if err := s.CreateKey(APIKey{
		ID:          idGen(),
		Label:       label,
		Prefix:      prefix,
		Hash:        string(hash),
		Permissions: perms,
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		return "", err
	}
	return raw, nil
}

func RandomKeySuffix(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		for i := range b {
			b[i] = letters[i%len(letters)]
		}
		return string(b)
	}
	for i := range b {
		b[i] = letters[int(buf[i])%len(letters)]
	}
	return string(b)
}

func (s *Store) CreateKey(k APIKey) error {
	perms, err := json.Marshal(k.Permissions)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO api_keys (id, label, prefix, hash, permissions, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, k.ID, k.Label, k.Prefix, k.Hash, string(perms), k.CreatedAt, nil)
	return err
}

func (s *Store) ListKeys() ([]APIKey, error) {
	rows, err := s.db.Query(`
		SELECT id, label, prefix, hash, permissions, created_at, last_used_at
		FROM api_keys ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func scanAPIKey(sc interface {
	Scan(dest ...any) error
}) (APIKey, error) {
	var k APIKey
	var last sql.NullTime
	pj := permsJSON{dst: &k.Permissions}
	err := sc.Scan(&k.ID, &k.Label, &k.Prefix, &k.Hash, &pj, &k.CreatedAt, &last)
	if err != nil {
		return APIKey{}, err
	}
	if last.Valid {
		t := last.Time
		k.LastUsedAt = &t
	}
	return k, nil
}

type permsJSON struct {
	dst *[]string
}

func (p *permsJSON) Scan(src any) error {
	if p.dst == nil {
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return nil
	}
	return json.Unmarshal(data, p.dst)
}

func (s *Store) FindKeyByRaw(raw string) (APIKey, error) {
	prefix := extractPrefix(raw)
	if prefix == "" {
		return APIKey{}, errors.New("key not found")
	}
	keys, err := s.getKeysByPrefix(prefix)
	if err != nil {
		return APIKey{}, err
	}
	for _, k := range keys {
		if err := bcrypt.CompareHashAndPassword([]byte(k.Hash), []byte(raw)); err == nil {
			return k, nil
		}
	}
	return APIKey{}, errors.New("key not found")
}

func extractPrefix(raw string) string {
	if strings.HasPrefix(raw, "pb_admin_") {
		return "pb_admin_"
	}
	if strings.HasPrefix(raw, "pb_sync_") {
		return "pb_sync_"
	}
	return ""
}

func (s *Store) getKeysByPrefix(prefix string) ([]APIKey, error) {
	rows, err := s.db.Query(`
		SELECT id, label, prefix, hash, permissions, created_at, last_used_at
		FROM api_keys WHERE prefix = ?
	`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (s *Store) RevokeKey(id string) error {
	res, err := s.db.Exec(`DELETE FROM api_keys WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) TouchKey(id string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET last_used_at = ? WHERE id = ?`, time.Now().UTC(), id)
	return err
}

func KeyPrefixForType(keyType string) (prefix string, perms []string, err error) {
	switch strings.ToLower(strings.TrimSpace(keyType)) {
	case "sync":
		return "pb_sync_", []string{"posts:write"}, nil
	case "admin":
		return "pb_admin_", []string{"admin"}, nil
	default:
		return "", nil, fmt.Errorf("type must be sync or admin")
	}
}
