package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/server"
	"pressbin.dev/pressbin/internal/store"
)

//go:embed all:assets
var assetsFS embed.FS

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("pressbin %s\n", Version)
		os.Exit(0)
	}

	configPath := flag.String("config", "config.yml", "optional config.yml (overridden by PRESSBIN_* env vars)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel()})))

	st, err := store.New(cfg.Database.Path)
	if err != nil {
		slog.Error("store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	if err := st.ApplySiteConfig(cfg.Site.Title, cfg.Site.Description, cfg.Site.URL, cfg.Site.PostsPerPage); err != nil {
		slog.Error("apply site config", "err", err)
		os.Exit(1)
	}

	if raw, err := st.Bootstrap(); err != nil {
		slog.Error("bootstrap", "err", err)
		os.Exit(1)
	} else if raw != "" {
		fmt.Printf(`
┌─────────────────────────────────────────────┐
│  Pressbin %s — First Run                    │
│                                             │
│  Admin API Key:                             │
│  %s                                         │
│                                             │
│  Save this. It will not be shown again.     │
└─────────────────────────────────────────────┘
`, Version, raw)
	}

	srv := server.New(st, cfg, assetsFS)
	addr := cfg.ListenAddr()
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
