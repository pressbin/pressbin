package preflight_test

import (
	"os"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/preflight"
	"pressbin.dev/pressbin/internal/store"
)

func TestRun_ok(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	dbDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dbDir, "pressbin.db")

	st, err := store.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateAdminKey("admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateSyncKey("sync"); err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: dbPath},
		Assets:   config.AssetsConfig{Path: assets},
	}
	if err := preflight.Run(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestRun_missingKeys(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	_ = os.MkdirAll(assets, 0o755)
	dbPath := filepath.Join(dir, "pressbin.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{Path: dbPath},
		Assets:   config.AssetsConfig{Path: assets},
	}
	if err := preflight.Run(cfg); err == nil {
		t.Fatal("expected error")
	}
}
