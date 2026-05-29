package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigPath resolves the default config file for serve/check.
// Order: ./.pressbin-dev/config.yml (monorepo dev), ~/.pressbin/config.yml, ./config.yml.
func ConfigPath() string {
	if home, err := LocalDevHome(); err == nil {
		p := filepath.Join(home, "config.yml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if home, err := DefaultHome(); err == nil {
		p := Paths(home).Config
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "config.yml"
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
