package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"pressbin.in/pressbin/internal/config"
	"pressbin.in/pressbin/internal/server"
	"pressbin.in/pressbin/internal/store"
)

//go:embed all:assets
var assetsFS embed.FS

func main() {
	configPath := flag.String("config", "config.yml", "path to config.yml")
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

	if raw, err := st.Bootstrap(); err != nil {
		slog.Error("bootstrap", "err", err)
		os.Exit(1)
	} else if raw != "" {
		fmt.Printf(`
┌─────────────────────────────────────────────┐
│  Pressbin — First Run                       │
│                                             │
│  Admin API Key:                             │
│  %s                                         │
│                                             │
│  Save this. It will not be shown again.     │
└─────────────────────────────────────────────┘
`, raw)
	}

	srv := server.New(st, cfg, assetsFS)
	addr := cfg.ListenAddr()
	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
