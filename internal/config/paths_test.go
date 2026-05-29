package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/config"
)

func TestLoad_rejectsAssetsInDatabaseDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "pressbin.db")
	path := writeConfig(t, dir, "config.yml", `
database:
  path: `+dbPath+`
assets:
  path: `+dir+`
`)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when assets.path is database directory")
	}
}

func TestLoad_rejectsAssetsContainingDatabase(t *testing.T) {
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	dbPath := filepath.Join(assetsDir, "pressbin.db")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath, []byte("sqlite"), 0o644); err != nil {
		t.Fatal(err)
	}

	path := writeConfig(t, dir, "config.yml", `
database:
  path: `+dbPath+`
assets:
  path: `+assetsDir+`
`)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when assets.path contains database file")
	}
}

func TestLoad_allowsPressbinHomeStylePaths(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, ".pressbin")
	dbPath := filepath.Join(home, "data", "pressbin.db")
	assetsPath := filepath.Join(home, "assets")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(assetsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	path := writeConfig(t, dir, "config.yml", `
database:
  path: `+dbPath+`
assets:
  path: `+assetsPath+`
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Assets.Path != assetsPath {
		t.Errorf("assets.path = %q", cfg.Assets.Path)
	}
}

func TestLoad_allowsSeparateAssetsDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "data", "pressbin.db")
	assetsPath := filepath.Join(dir, "assets")
	path := writeConfig(t, dir, "config.yml", `
database:
  path: `+dbPath+`
assets:
  path: `+assetsPath+`
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Assets.Path != assetsPath {
		t.Errorf("assets.path = %q, want %q", cfg.Assets.Path, assetsPath)
	}
}
