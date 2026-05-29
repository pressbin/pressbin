package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigPath resolves the default config file for serve/check.
//
// Order:
//  1. ../config.yml relative to the pressbin binary (…/bin/pressbin → …/config.yml)
//     This is where `pressbin setup --home DIR` writes config.yml.
//  2. ./.pressbin-dev/config.yml (monorepo dev)
//  3. ~/.pressbin/config.yml
//  4. ./config.yml
func ConfigPath() string {
	if p := configFromExecutable(); p != "" {
		return p
	}
	if home, err := LocalDevHome(); err == nil {
		if p := configInHome(home); p != "" {
			return p
		}
	}
	if home, err := DefaultHome(); err == nil {
		if p := configInHome(home); p != "" {
			return p
		}
	}
	return "config.yml"
}

func configInHome(home string) string {
	home = strings.TrimSpace(home)
	if home == "" {
		return ""
	}
	p := Paths(home).Config
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func configFromExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return ConfigPathForExecutable(exe)
}

// ConfigPathForExecutable returns …/config.yml when exe lives in …/bin/pressbin (for tests).
func ConfigPathForExecutable(exe string) string {
	exe, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}
	if filepath.Base(exe) != "pressbin" {
		return ""
	}
	binDir := filepath.Dir(exe)
	if filepath.Base(binDir) != "bin" {
		return ""
	}
	return configInHome(filepath.Dir(binDir))
}

const (
	DefaultHomeName = ".pressbin"
	// LocalDevDir is the monorepo dev home (same layout as DefaultHomeName on servers).
	LocalDevDir = ".pressbin-dev"
)

// LocalDevHome returns $PWD/.pressbin-dev for engine development.
func LocalDevHome() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, LocalDevDir), nil
}

// Layout is the on-disk layout under the Pressbin home directory.
type Layout struct {
	Home         string
	Binary       string
	Config       string
	DB           string
	Assets       string
	AdminKeyFile string
	SyncKeyFile  string
}

func DefaultHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, DefaultHomeName), nil
}

func Paths(home string) Layout {
	home = filepath.Clean(home)
	return Layout{
		Home:         home,
		Binary:       filepath.Join(home, "bin", "pressbin"),
		Config:       filepath.Join(home, "config.yml"),
		DB:           filepath.Join(home, "data", "pressbin.db"),
		Assets:       filepath.Join(home, "assets"),
		AdminKeyFile: filepath.Join(home, "admin.key"),
		SyncKeyFile:  filepath.Join(home, "sync.key"),
	}
}
