package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"net/url"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const envPrefix = "PRESSBIN_"

// envToConfigKey maps PRESSBIN_* variables to koanf keys (dotted).
var envToConfigKey = map[string]string{
	"SERVER_PORT":          "server.port",
	"SERVER_HOST":          "server.host",
	"DATABASE_PATH":        "database.path",
	"SITE_TITLE":           "site.title",
	"SITE_DESCRIPTION":     "site.description",
	"SITE_URL":             "site.url",
	"SITE_POSTS_PER_PAGE":  "site.posts_per_page",
	"THEME_CUSTOM_CSS_URL": "theme.custom_css_url",
	// New asset upload configuration.
	"ASSETS_UPLOAD_DRIVER": "assets.upload.driver",
	"ASSETS_STORAGE_PATH":  "assets.upload.storage_path",
	"LOG_LEVEL":            "log.level",
}

type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	Site     SiteConfig     `koanf:"site"`
	Theme    ThemeConfig    `koanf:"theme"`
	Assets   AssetsConfig   `koanf:"assets"`
	Log      LogConfig      `koanf:"log"`
}

type ServerConfig struct {
	Port int    `koanf:"port"`
	Host string `koanf:"host"`
}

type DatabaseConfig struct {
	Path string `koanf:"path"`
}

type SiteConfig struct {
	Title        string `koanf:"title"`
	Description  string `koanf:"description"`
	URL          string `koanf:"url"`
	PostsPerPage int    `koanf:"posts_per_page"`
}

type ThemeConfig struct {
	// CustomCSSURL, if set, is linked after the default theme CSS.
	// Must be an absolute http(s) URL.
	CustomCSSURL string `koanf:"custom_css_url"`
}

type AssetsConfig struct {
	Upload AssetsUploadConfig `koanf:"upload"`
	// AllowedPrefixes controls which subfolders can be uploaded via /api/sync/asset.
	// Defaults to ["images/", "fonts/", "downloads/", "css/"].
	AllowedPrefixes []string `koanf:"allowed_prefixes"`
}

type AssetsUploadConfig struct {
	Driver      string `koanf:"driver"`       // local (v1). Future: scp, s3, r2, ftp.
	StoragePath string `koanf:"storage_path"` // root dir where assets are written
}

type LogConfig struct {
	Level string `koanf:"level"`
}

// Load merges configuration from an optional YAML file and PRESSBIN_* environment
// variables. Environment variables override YAML. If path is empty or the file is
// missing, only defaults and env vars apply.
func Load(path string) (*Config, error) {
	k := koanf.New(".")

	configDir, err := loadYAML(k, path)
	if err != nil {
		return nil, err
	}

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.TrimPrefix(k, envPrefix)
			ck, ok := envToConfigKey[k]
			if !ok {
				return "", nil
			}
			return ck, v
		},
	}), nil); err != nil {
		return nil, fmt.Errorf("load environment: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	applyDefaults(&cfg)
	if err := validateTheme(&cfg); err != nil {
		return nil, err
	}
	cfg.Database.Path = resolveAgainstDir(configDir, cfg.Database.Path)
	cfg.Assets.Upload.StoragePath = resolveAgainstDir(configDir, cfg.Assets.Upload.StoragePath)

	return &cfg, nil
}

func loadYAML(k *koanf.Koanf, path string) (configDir string, err error) {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		wd = "."
	}

	if strings.TrimSpace(path) == "" {
		return wd, nil
	}

	absConfig, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("config path: %w", err)
	}

	if _, err := os.Stat(absConfig); err != nil {
		if os.IsNotExist(err) {
			return wd, nil
		}
		return "", fmt.Errorf("config file %s: %w", absConfig, err)
	}

	if err := k.Load(file.Provider(absConfig), yaml.Parser()); err != nil {
		return "", fmt.Errorf("load config %s: %w", absConfig, err)
	}
	return filepath.Dir(absConfig), nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./pressbin.db"
	}
	if cfg.Site.PostsPerPage == 0 {
		cfg.Site.PostsPerPage = 10
	}
	if strings.TrimSpace(cfg.Assets.Upload.Driver) == "" {
		cfg.Assets.Upload.Driver = "local"
	}
	if len(cfg.Assets.AllowedPrefixes) == 0 {
		cfg.Assets.AllowedPrefixes = []string{"images/", "fonts/", "downloads/", "css/"}
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
}

func validateTheme(cfg *Config) error {
	raw := strings.TrimSpace(cfg.Theme.CustomCSSURL)
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return fmt.Errorf("theme.custom_css_url: invalid url: %q", raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("theme.custom_css_url: url scheme must be http or https: %q", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("theme.custom_css_url: url must be absolute: %q", raw)
	}
	cfg.Theme.CustomCSSURL = raw
	return nil
}

// resolveAgainstDir makes relative paths relative to configDir (YAML file directory,
// or process working directory when no config file was loaded).
func resolveAgainstDir(configDir, p string) string {
	if p == "" {
		return p
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(configDir, filepath.Clean(p)))
}

func (c *Config) LogLevel() slog.Level {
	switch c.Log.Level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// EnvVars documents supported PRESSBIN_* variables (for help text and docs).
func EnvVars() []struct{ Var, ConfigKey string } {
	out := make([]struct{ Var, ConfigKey string }, 0, len(envToConfigKey))
	for envKey, cfgKey := range envToConfigKey {
		out = append(out, struct{ Var, ConfigKey string }{
			Var:       envPrefix + envKey,
			ConfigKey: cfgKey,
		})
	}
	return out
}
