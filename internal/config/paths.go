package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validatePaths rejects unsafe combinations of database, assets, and binary locations.
func validatePaths(cfg *Config) error {
	assetsPath := strings.TrimSpace(cfg.Assets.Path)
	if assetsPath == "" {
		return nil
	}

	assetsAbs, err := filepath.Abs(assetsPath)
	if err != nil {
		return fmt.Errorf("assets.path: %w", err)
	}
	cfg.Assets.Path = filepath.Clean(assetsAbs)

	dbAbs, err := filepath.Abs(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("database.path: %w", err)
	}
	cfg.Database.Path = filepath.Clean(dbAbs)

	if isWithin(cfg.Assets.Path, dbAbs) {
		return fmt.Errorf("assets.path must not contain the database file (would be served as a static file)")
	}
	if filepath.Dir(dbAbs) == cfg.Assets.Path {
		return fmt.Errorf("assets.path must not be the database directory")
	}

	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	exeDir, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		return nil
	}
	if isWithin(exeDir, cfg.Assets.Path) {
		return fmt.Errorf("assets.path must not be inside the pressbin binary directory (%s)", exeDir)
	}
	if cfg.Assets.Path == exeDir {
		return fmt.Errorf("assets.path must not be the pressbin binary directory")
	}
	return nil
}

// isWithin reports whether target is base or a descendant of base.
func isWithin(base, target string) bool {
	base = filepath.Clean(base)
	target = filepath.Clean(target)
	if base == target {
		return true
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
