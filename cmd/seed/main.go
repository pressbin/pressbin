// Command seed inserts synthetic posts into the Pressbin database for load testing.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"pressbin.dev/pressbin/internal/config"
	"pressbin.dev/pressbin/internal/parser"
	"pressbin.dev/pressbin/internal/store"
)

const slugPrefix = "load-test-"

var tagPool = []string{
	"performance", "load-test", "sample", "go", "sqlite", "search",
	"pagination", "dev", "benchmark", "blog",
}

var bodyParagraphs = []string{
	"This post exists to exercise list, search, and pagination under realistic volume.",
	"Pressbin stores markdown in SQLite and renders HTML on sync. Full-text search uses FTS5.",
	"Try searching for unique words in these titles, or browse by tag once enough posts exist.",
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt.",
	"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip.",
	"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat.",
}

func main() {
	configPath := flag.String("config", ".pressbin-dev/config.yml", "Pressbin config.yml (database path)")
	count := flag.Int("count", 55, "number of sample posts to insert")
	prefix := flag.String("prefix", slugPrefix, "slug prefix for generated posts")
	clear := flag.Bool("clear", false, "delete existing posts with the slug prefix before seeding")
	flag.Parse()

	if *count < 1 {
		fmt.Fprintln(os.Stderr, "count must be at least 1")
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

	if *clear {
		removed, err := clearPosts(st, *prefix)
		if err != nil {
			slog.Error("clear", "err", err)
			os.Exit(1)
		}
		slog.Info("cleared seed posts", "removed", removed, "prefix", *prefix)
	}

	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	now := time.Now().UTC()
	var inserted int

	for i := 1; i <= *count; i++ {
		slug := fmt.Sprintf("%s%03d", *prefix, i)
		title := fmt.Sprintf("Load Test Post #%d: %s", i, topicFor(i))
		tags := tagsFor(i)
		summary := fmt.Sprintf("Synthetic post %d for load and pagination testing.", i)
		md := buildMarkdown(title, summary, tags, i)

		parsed, err := parser.Parse([]byte(md))
		if err != nil {
			slog.Error("parse", "slug", slug, "err", err)
			os.Exit(1)
		}

		post := store.Post{
			Slug:        slug,
			Title:       title,
			ContentMD:   md,
			ContentHTML: parsed.HTML,
			Summary:     summary,
			Tags:        tags,
			Status:      "published",
			PublishedAt: start.AddDate(0, 0, i-1),
			UpdatedAt:   now,
		}

		if err := st.UpsertPost(post); err != nil {
			slog.Error("upsert", "slug", slug, "err", err)
			os.Exit(1)
		}
		inserted++
	}

	slog.Info("seed complete", "inserted", inserted, "db", cfg.Database.Path, "prefix", *prefix)
}

func clearPosts(st *store.Store, prefix string) (int, error) {
	posts, _, err := st.ListAllPosts(1, 10000)
	if err != nil {
		return 0, err
	}
	var removed int
	for _, p := range posts {
		if !strings.HasPrefix(p.Slug, prefix) {
			continue
		}
		if err := st.DeletePost(p.Slug); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func topicFor(i int) string {
	topics := []string{
		"Benchmarking SQLite", "Search Under Load", "Pagination Patterns",
		"HTMX Live Search", "Markdown Rendering", "Tag Clouds at Scale",
		"RSS Feed Size", "Draft vs Published", "Sync Throughput", "Index Performance",
	}
	return topics[i%len(topics)]
}

func tagsFor(i int) []string {
	a := tagPool[i%len(tagPool)]
	b := tagPool[(i*3+1)%len(tagPool)]
	if a == b {
		return []string{a}
	}
	return []string{a, b}
}

func buildMarkdown(title, summary string, tags []string, n int) string {
	tagYAML := fmt.Sprintf("[%s]", strings.Join(tags, ", "))
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: %s\nsummary: %s\ndate: 2024-01-01\ntags: %s\n---\n\n", title, summary, tagYAML)
	fmt.Fprintf(&b, "# %s\n\n", title)
	for j := 0; j < 3; j++ {
		p := bodyParagraphs[(n+j)%len(bodyParagraphs)]
		fmt.Fprintf(&b, "%s\n\n", p)
	}
	fmt.Fprintf(&b, "Unique token for search: **loadtoken-%04d**.\n", n)
	return b.String()
}
