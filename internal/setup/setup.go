package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/store"
)

type Options struct {
	Home      string
	SiteURL   string
	SiteTitle string
	Host      string
	Port      int
	LogLevel  string
}

type Result struct {
	Layout Layout
	Config *config.Config
}

func Run(opts Options) (*Result, error) {
	home := strings.TrimSpace(opts.Home)
	if home == "" {
		var err error
		home, err = DefaultHome()
		if err != nil {
			return nil, err
		}
	}
	home, err := filepath.Abs(home)
	if err != nil {
		return nil, err
	}

	siteURL := strings.TrimSpace(opts.SiteURL)
	if siteURL == "" {
		return nil, fmt.Errorf("--site-url is required")
	}
	siteTitle := strings.TrimSpace(opts.SiteTitle)
	if siteTitle == "" {
		siteTitle = "My Blog"
	}
	host := strings.TrimSpace(opts.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := opts.Port
	if port == 0 {
		port = 8080
	}
	logLevel := strings.TrimSpace(opts.LogLevel)
	if logLevel == "" {
		logLevel = "info"
	}

	layout := Paths(home)
	for _, dir := range []string{
		filepath.Dir(layout.Binary),
		filepath.Dir(layout.DB),
		layout.Assets,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	if _, err := os.Stat(layout.Config); os.IsNotExist(err) {
		if err := writeConfig(layout.Config, siteTitle, siteURL, host, port, layout.DB, layout.Assets, logLevel); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	cfg, err := config.Load(layout.Config)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	st, err := store.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	defer st.Close()

	if err := ensureAdminKey(st, layout.AdminKeyFile); err != nil {
		return nil, err
	}
	if err := ensureSyncKey(st, layout.SyncKeyFile); err != nil {
		return nil, err
	}

	if err := st.ApplySiteConfig(cfg.Site.Title, cfg.Site.Description, cfg.Site.URL, cfg.Site.PostsPerPage); err != nil {
		return nil, fmt.Errorf("apply site config: %w", err)
	}

	return &Result{Layout: layout, Config: cfg}, nil
}

func writeConfig(path, title, siteURL, host string, port int, dbPath, assetsPath, logLevel string) error {
	body := fmt.Sprintf(`server:
  port: %d
  host: %s

database:
  path: %s

site:
  title: %q
  description: ""
  url: %q
  posts_per_page: 10

assets:
  path: %s

log:
  level: %s
`, port, host, dbPath, title, siteURL, assetsPath, logLevel)
	return os.WriteFile(path, []byte(body), 0o644)
}

func ensureAdminKey(st *store.Store, keyFile string) error {
	has, err := st.HasKeysWithPrefix("pb_admin_")
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	raw, err := st.CreateAdminKey("setup admin")
	if err != nil {
		return fmt.Errorf("create admin key: %w", err)
	}
	return writeKeyFile(keyFile, raw)
}

func ensureSyncKey(st *store.Store, keyFile string) error {
	has, err := st.HasKeysWithPrefix("pb_sync_")
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	raw, err := st.CreateSyncKey("setup sync")
	if err != nil {
		return fmt.Errorf("create sync key: %w", err)
	}
	return writeKeyFile(keyFile, raw)
}

func writeKeyFile(path, raw string) error {
	if err := os.WriteFile(path, []byte(raw+"\n"), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// PrintSummary writes post-setup instructions to stdout.
func PrintSummary(layout Layout, cfg *config.Config) {
	fmt.Printf(`
Pressbin is ready.

  Home:     %s
  Config:   %s
  Database: %s
  Assets:   %s
  Admin:    %s
  Sync:     %s

Start the server:
  %s serve

Or add to PATH (serve loads config next to the binary):
  export PATH=%s/bin:$PATH
  pressbin serve

GitHub Actions (blog repo secrets):
  PRESSBIN_URL=%s
  PRESSBIN_KEY=<contents of %s>
`, layout.Home, layout.Config, cfg.Database.Path, cfg.Assets.Path,
		layout.AdminKeyFile, layout.SyncKeyFile,
		layout.Binary, layout.Home,
		cfg.Site.URL, layout.SyncKeyFile)
}
