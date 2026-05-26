// Command migrate opens the SQLite database from config and applies pending
// embedded migrations (same as on server startup), then exits.
package main

import (
	"flag"
	"fmt"
	"os"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/store"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to config.yml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	st, err := store.New(cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	fmt.Printf("migrations applied: %s\n", cfg.Database.Path)
}
