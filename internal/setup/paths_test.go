package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"pressbin.dev/pressbin/internal/setup"
)

func TestConfigPath_nextToBinary(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "site", ".pressbin")
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(home, "config.yml")
	if err := os.WriteFile(cfg, []byte("assets:\n  path: assets\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fakeBin := filepath.Join(binDir, "pressbin")
	if err := os.WriteFile(fakeBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", filepath.Join(root, "empty-home"))

	got := setup.ConfigPathForExecutable(fakeBin)
	if got == "" {
		t.Fatal("ConfigPathForExecutable returned empty")
	}
	gotInfo, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	wantInfo, err := os.Stat(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(gotInfo, wantInfo) {
		t.Fatalf("ConfigPathForExecutable = %q, want %q", got, cfg)
	}
}
