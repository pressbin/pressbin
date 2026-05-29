package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/preflight"
	"pressbin.dev/pressbin/internal/server"
	"pressbin.dev/pressbin/internal/setup"
	"pressbin.dev/pressbin/internal/store"
)

//go:embed all:assets
var assetsFS embed.FS

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("pressbin %s\n", Version)
			return
		case "setup":
			os.Exit(runSetup(os.Args[2:]))
		case "check":
			os.Exit(runCheck(os.Args[2:]))
		case "serve":
			os.Exit(runServe(os.Args[2:]))
		case "-h", "--help", "help":
			printUsage()
			return
		}
	}
	os.Exit(runServe(os.Args[1:]))
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `pressbin %s

Usage:
  pressbin [serve] [--config path] [--allow-bootstrap]
  pressbin setup --site-url URL [--home dir] [--site-title title] [--host host] [--port port]
  pressbin check [--config path]
  pressbin version

Default config: .pressbin-dev/config.yml, then ~/.pressbin/config.yml, then ./config.yml
`, Version)
}

func defaultConfigPath() string {
	return setup.ConfigPath()
}

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	home := fs.String("home", "", "Pressbin home directory (default: ~/.pressbin)")
	siteURL := fs.String("site-url", "", "Public site URL (required)")
	siteTitle := fs.String("site-title", "My Blog", "Site title")
	host := fs.String("host", "127.0.0.1", "Listen host")
	port := fs.Int("port", 8080, "Listen port")
	logLevel := fs.String("log-level", "info", "Log level: debug, info, warn, error")
	_ = fs.Parse(args)

	if strings.TrimSpace(*siteURL) == "" {
		if v := strings.TrimSpace(os.Getenv("PRESSBIN_SITE_URL")); v != "" {
			*siteURL = v
		}
	}

	res, err := setup.Run(setup.Options{
		Home:      *home,
		SiteURL:   *siteURL,
		SiteTitle: *siteTitle,
		Host:      *host,
		Port:      *port,
		LogLevel:  *logLevel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup: %v\n", err)
		return 1
	}
	setup.PrintSummary(res.Layout, res.Config)
	return 0
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath(), "config file path")
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		return 1
	}
	if err := preflight.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "check: %v\n", err)
		return 1
	}
	fmt.Println("ok")
	return 0
}

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath(), "config file path")
	allowBootstrap := fs.Bool("allow-bootstrap", false, "create admin key on first run if missing (dev only)")
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		return 1
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel()})))

	if *allowBootstrap {
		st, err := store.New(cfg.Database.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "store: %v\n", err)
			return 1
		}
		if raw, err := st.Bootstrap(); err != nil {
			_ = st.Close()
			fmt.Fprintf(os.Stderr, "bootstrap: %v\n", err)
			return 1
		} else if raw != "" {
			fmt.Fprintf(os.Stderr, "warning: created admin key (allow-bootstrap): save it now\n%s\n", raw)
		}
		_ = st.Close()
	}

	if err := preflight.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "preflight: %v\n", err)
		return 1
	}

	st, err := store.New(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "store: %v\n", err)
		return 1
	}
	defer st.Close()

	if err := st.ApplySiteConfig(cfg.Site.Title, cfg.Site.Description, cfg.Site.URL, cfg.Site.PostsPerPage); err != nil {
		fmt.Fprintf(os.Stderr, "apply site config: %v\n", err)
		return 1
	}

	srv := server.New(st, cfg, assetsFS)
	addr := cfg.ListenAddr()
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server", "err", err)
		return 1
	}
	return 0
}
