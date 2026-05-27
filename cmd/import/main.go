package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/parser"
	"pressbin.dev/pressbin/internal/store"
)

func main() {
	configPath := flag.String("config", "config.yml", "Pressbin config.yml (for database + site settings)")
	contentDir := flag.String("content", "", "content repo directory (expects posts/**/*.md)")
	dryRun := flag.Bool("dry-run", false, "parse and print what would be imported without writing to the DB")
	flag.Parse()

	if strings.TrimSpace(*contentDir) == "" {
		fmt.Fprintln(os.Stderr, "missing -content (e.g. ../pressbin_blog)")
		os.Exit(2)
	}

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

	postsRoot := filepath.Join(*contentDir, "posts")
	var imported int
	var failed int

	err = filepath.WalkDir(postsRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".md" {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			failed++
			slog.Error("read post", "path", path, "err", err)
			return nil
		}

		parsed, err := parser.Parse(b)
		if err != nil {
			failed++
			slog.Error("parse post", "path", path, "err", err)
			return nil
		}

		slug := strings.TrimSpace(parsed.Slug)
		if slug == "" {
			slug = strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		}
		slug = strings.TrimSpace(slug)
		if slug == "" {
			failed++
			slog.Error("invalid slug", "path", path)
			return nil
		}

		title := strings.TrimSpace(parsed.Title)
		if title == "" {
			title = slug
		}

		status := strings.TrimSpace(parsed.Status)
		if status == "" {
			status = "published"
		}

		post := store.Post{
			Slug:        slug,
			Title:       title,
			ContentMD:   string(b),
			ContentHTML: parsed.HTML,
			Summary:     strings.TrimSpace(parsed.Summary),
			Tags:        parsed.Tags,
			Status:      status,
			PublishedAt: parseDate(parsed.PublishedAt),
			UpdatedAt:   time.Now().UTC(),
		}

		if *dryRun {
			imported++
			slog.Info("would import", "slug", slug, "path", path, "status", post.Status)
			return nil
		}

		if err := st.UpsertPost(post); err != nil {
			failed++
			slog.Error("upsert post", "slug", slug, "path", path, "err", err)
			return nil
		}

		imported++
		slog.Info("imported", "slug", slug, "path", path, "status", post.Status)
		return nil
	})
	if err != nil {
		slog.Error("walk posts", "root", postsRoot, "err", err)
		os.Exit(1)
	}

	if *dryRun {
		slog.Info("dry-run complete", "posts_found", imported, "failed", failed)
		return
	}
	slog.Info("import complete", "imported", imported, "failed", failed, "db", cfg.Database.Path)
	if failed > 0 {
		os.Exit(1)
	}
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().UTC()
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

