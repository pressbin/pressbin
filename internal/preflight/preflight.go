package preflight

import (
	"fmt"
	"os"
	"strings"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/store"
)

// Run validates configuration and database readiness before serving.
func Run(cfg *config.Config) error {
	if strings.TrimSpace(cfg.Assets.Path) != "" {
		if err := checkDirWritable(cfg.Assets.Path, "assets.path"); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("assets.path is required; run: pressbin setup")
	}

	st, err := store.New(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer st.Close()

	hasAdmin, err := st.HasKeysWithPrefix("pb_admin_")
	if err != nil {
		return fmt.Errorf("api keys: %w", err)
	}
	if !hasAdmin {
		return fmt.Errorf("no admin API key found; run: pressbin setup")
	}

	hasSync, err := st.HasKeysWithPrefix("pb_sync_")
	if err != nil {
		return fmt.Errorf("api keys: %w", err)
	}
	if !hasSync {
		return fmt.Errorf("no sync API key found; run: pressbin setup")
	}

	return nil
}

func checkDirWritable(dir, label string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory: %s", label, dir)
	}
	f, err := os.CreateTemp(dir, ".pressbin-write-test-*")
	if err != nil {
		return fmt.Errorf("%s is not writable: %s", label, dir)
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return nil
}
