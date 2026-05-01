package config

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	Site     SiteConfig     `koanf:"site"`
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
	Title          string `koanf:"title"`
	Description    string `koanf:"description"`
	URL            string `koanf:"url"`
	PostsPerPage   int    `koanf:"posts_per_page"`
}

type LogConfig struct {
	Level string `koanf:"level"`
}

func Load(path string) (*Config, error) {
	absConfig, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("config path: %w", err)
	}
	configDir := filepath.Dir(absConfig)

	k := koanf.New(".")

	if err := k.Load(file.Provider(absConfig), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load config %s: %w", absConfig, err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./pressbin.db"
	}
	cfg.Database.Path = resolveAgainstDir(configDir, cfg.Database.Path)
	if cfg.Site.PostsPerPage == 0 {
		cfg.Site.PostsPerPage = 10
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}

	return &cfg, nil
}

// resolveAgainstDir makes relative paths (e.g. ./pressbin.db) relative to the
// config file directory, not the process working directory, so Air/IDE/monorepo
// cwd does not silently point at a different or empty database.
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
