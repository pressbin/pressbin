package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"pressbin.dev/pressbin/internal/config"
)

func writeConfig(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad_defaultsAndRelativeDB(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, "config.yml", `
server:
  port: 9090
site:
  title: My Blog
  url: https://example.com
database:
  path: ./data/blog.db
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("port = %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("host = %q", cfg.Server.Host)
	}
	if cfg.Site.Title != "My Blog" {
		t.Errorf("title = %q", cfg.Site.Title)
	}
	if cfg.Site.PostsPerPage != 10 {
		t.Errorf("posts_per_page = %d", cfg.Site.PostsPerPage)
	}
	wantDB := filepath.Join(dir, "data", "blog.db")
	if cfg.Database.Path != wantDB {
		t.Errorf("database path = %q, want %q", cfg.Database.Path, wantDB)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("log level = %q", cfg.Log.Level)
	}
}

func TestLoad_absoluteDBPath(t *testing.T) {
	dir := t.TempDir()
	absDB := filepath.Join(dir, "pressbin.db")
	path := writeConfig(t, dir, "config.yml", "database:\n  path: "+absDB+"\n")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Path != absDB {
		t.Errorf("path = %q, want %q", cfg.Database.Path, absDB)
	}
}

func TestLoad_missingFileUsesDefaults(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.yml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d", cfg.Server.Port)
	}
	if cfg.Database.Path == "" {
		t.Error("expected default database path")
	}
}

func TestLoad_envOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, "config.yml", `
server:
  port: 9090
site:
  title: From YAML
`)

	t.Setenv("PRESSBIN_SERVER_PORT", "3001")
	t.Setenv("PRESSBIN_SITE_TITLE", "From Env")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 3001 {
		t.Errorf("port = %d, want env override", cfg.Server.Port)
	}
	if cfg.Site.Title != "From Env" {
		t.Errorf("title = %q", cfg.Site.Title)
	}
}

func TestLoad_envOnly(t *testing.T) {
	t.Setenv("PRESSBIN_SERVER_PORT", "4000")
	t.Setenv("PRESSBIN_SITE_URL", "https://env.example")
	t.Setenv("PRESSBIN_DATABASE_PATH", "./env.db")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 4000 {
		t.Errorf("port = %d", cfg.Server.Port)
	}
	if cfg.Site.URL != "https://env.example" {
		t.Errorf("url = %q", cfg.Site.URL)
	}
	if !filepath.IsAbs(cfg.Database.Path) && cfg.Database.Path != filepath.Join(mustWd(t), "env.db") {
		t.Errorf("database path = %q", cfg.Database.Path)
	}
}

func mustWd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}

func TestConfig_LogLevel(t *testing.T) {
	tests := []struct {
		level string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}
	for _, tc := range tests {
		cfg := &config.Config{Log: config.LogConfig{Level: tc.level}}
		if got := cfg.LogLevel(); got != tc.want {
			t.Errorf("LogLevel(%q) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

func TestConfig_ListenAddr(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 3000},
	}
	if got := cfg.ListenAddr(); got != "127.0.0.1:3000" {
		t.Errorf("ListenAddr = %q", got)
	}
}
