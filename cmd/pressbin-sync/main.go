package main

import (
	"flag"
	"fmt"
	"os"

	"pressbin.dev/pressbin/internal/syncclient"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf("pressbin-sync %s\n", Version)
		return
	case "run":
		os.Exit(run(os.Args[2:]))
	case "-h", "--help", "help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `pressbin-sync %s

Sync a content repository to Pressbin (for CI or local use).

Usage:
  pressbin-sync run [--all] [repo-dir]
  pressbin-sync version

  --all   Upload every post and image (full resync; ignores git history)

Environment:
  PRESSBIN_URL   Public blog URL (required)
  PRESSBIN_KEY   Sync API key pb_sync_... (required)

Default repo directory: current working directory
`, Version)
}

func run(args []string) int {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	all := fs.Bool("all", false, "sync every post and asset (full resync)")
	_ = fs.Parse(args)

	repo := "."
	if fs.NArg() > 0 {
		repo = fs.Arg(0)
	}

	client, err := syncclient.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pressbin-sync: %v\n", err)
		return 1
	}
	if *all {
		if err := syncclient.RunAll(client, repo); err != nil {
			fmt.Fprintf(os.Stderr, "pressbin-sync: %v\n", err)
			return 1
		}
		return 0
	}
	if err := syncclient.Run(client, repo); err != nil {
		fmt.Fprintf(os.Stderr, "pressbin-sync: %v\n", err)
		return 1
	}
	return 0
}
