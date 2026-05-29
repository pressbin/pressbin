package testutil

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jaevor/go-nanoid"
	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/server"
	"pressbin.dev/pressbin/internal/store"
)

// NewStore opens a temporary SQLite database with migrations applied.
func NewStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "pressbin-test.db")
	st, err := store.New(path)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// TestConfig returns minimal site configuration for handlers.
func TestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{Port: 8080, Host: "127.0.0.1"},
		Database: config.DatabaseConfig{Path: "./test.db"},
		Site: config.SiteConfig{
			Title:          "Test Blog",
			Description:    "Test description",
			URL:            "http://test.local",
			PostsPerPage:   5,
		},
		Log: config.LogConfig{Level: "error"},
	}
}

// TestAssets is a minimal embedded asset tree for HTTP tests.
func TestAssets() fs.FS {
	return fstest.MapFS{
		"assets/style.css": &fstest.MapFile{Data: []byte("body { margin: 0; }")},
	}
}

// NewServer wires a test server with the given store.
func NewServer(t *testing.T, st *store.Store) *server.Server {
	t.Helper()
	return NewServerWithVersion(t, st, "dev")
}

// NewServerWithVersion wires a test server with an explicit API version string.
func NewServerWithVersion(t *testing.T, st *store.Store, version string) *server.Server {
	t.Helper()
	return server.New(st, TestConfig(), TestAssets(), version)
}

// MustAdminKey creates an admin API key and returns the raw secret.
func MustAdminKey(t *testing.T, st *store.Store) string {
	t.Helper()
	raw, err := st.Bootstrap()
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	if raw != "" {
		return raw
	}
	return insertKey(t, st, "pb_admin_", []string{"admin"})
}

// MustSyncKey creates a sync API key and returns the raw secret.
func MustSyncKey(t *testing.T, st *store.Store) string {
	t.Helper()
	return insertKey(t, st, "pb_sync_", []string{"posts:write"})
}

func insertKey(t *testing.T, st *store.Store, prefix string, perms []string) string {
	t.Helper()
	raw := prefix + store.RandomKeySuffix(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	idGen, err := nanoid.Standard(12)
	if err != nil {
		t.Fatalf("nanoid: %v", err)
	}
	k := store.APIKey{
		ID:          idGen(),
		Label:       "test",
		Prefix:      prefix,
		Hash:        string(hash),
		Permissions: perms,
		CreatedAt:   time.Now().UTC(),
	}
	if err := st.CreateKey(k); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	return raw
}

// SamplePost returns a published post suitable for UpsertPost.
func SamplePost(slug, title string) store.Post {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	md := "---\ntitle: " + title + "\ndate: 2026-05-01\ntags: [go, test]\n---\n\n# " + title + "\n\nBody text.\n"
	return store.Post{
		Slug:        slug,
		Title:       title,
		ContentMD:   md,
		ContentHTML: "<h1>" + title + "</h1><p>Body text.</p>",
		Summary:     "Summary for " + title,
		Tags:        []string{"go", "test"},
		Status:      "published",
		PublishedAt: now,
		UpdatedAt:   now,
	}
}

// DoRequest performs an HTTP request against h and returns the recorder.
func DoRequest(t *testing.T, h http.Handler, method, path string, body []byte, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
