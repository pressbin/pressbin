# Pressbin — Solo Build Guide

> Single binary blog engine. Go + SQLite + Markdown + Git push-to-live. Domain:
> pressbin.dev

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Tech Stack](#2-tech-stack)
3. [Project Structure](#3-project-structure)
4. [Database Schema](#4-database-schema)
5. [Module: Store](#5-module-store)
6. [Module: Parser](#6-module-parser)
7. [Module: Server](#7-module-server)
8. [Module: Sync API](#8-module-sync-api)
9. [Module: Admin API](#9-module-admin-api)
10. [Module: Render](#10-module-render)
11. [Auth & API Keys](#11-auth--api-keys)
12. [Config](#12-config)
13. [GitHub Action](#13-github-action)
14. [Build & Deploy](#14-build--deploy)
15. [Build Order](#15-build-order)

---

## 1. Project Overview

Pressbin is a self-contained blog engine compiled into a single binary. You
write posts in Markdown, store them in a GitHub repo, and a GitHub Action pushes
them to the binary via a signed API call. The binary stores everything in SQLite
and serves it fast.

**Core promises:**

- One binary + one config file = full deployment
- No PHP, Node, Docker, or separate DB on the server
- Git is the source of truth for content
- Full REST API so you can build any UI on top

---

## 2. Tech Stack

| Layer          | Library                                   | Version   | Why                                  |
| -------------- | ----------------------------------------- | --------- | ------------------------------------ |
| Language       | Go                                        | 1.22+     | Single binary compilation            |
| HTTP Router    | `github.com/go-chi/chi/v5`                | v5        | Standard net/http, composable        |
| HTML Rendering | `maragu.dev/gomponents`                   | latest    | Pure Go, no runtime panics           |
| CSS            | TailwindCSS                               | CDN (dev) | Utility classes in Go components     |
| JS             | HTMX                                      | CDN       | Search box, zero custom JS           |
| Icons          | `github.com/eduardolat/gomponents-lucide` | latest    | Clean SVG icons                      |
| SQLite         | `modernc.org/sqlite`                      | latest    | Pure Go, no CGO, truly single binary |
| Markdown       | `github.com/yuin/goldmark`                | latest    | Extensible, standard                 |
| Front Matter   | `github.com/yuin/goldmark-meta`           | latest    | Pairs with goldmark                  |
| Config         | `github.com/knadh/koanf/v2`               | v2        | Lighter than Viper                   |
| Logging        | `log/slog`                                | stdlib    | Built-in since Go 1.21               |
| Auth           | `crypto/hmac`                             | stdlib    | HMAC signing, no deps                |
| Password Hash  | `golang.org/x/crypto/bcrypt`              | latest    | For API key hashing                  |
| IDs            | `github.com/jaevor/go-nanoid`             | latest    | Short unique IDs for keys            |

### go.mod starting point

```
module github.com/yourname/pressbin

go 1.22

require (
    github.com/go-chi/chi/v5 v5.1.0
    maragu.dev/gomponents v1.0.0
    maragu.dev/gomponents-htmx v0.5.0
    github.com/eduardolat/gomponents-lucide v0.1.0
    modernc.org/sqlite v1.29.0
    github.com/yuin/goldmark v1.7.1
    github.com/yuin/goldmark-meta v1.1.0
    github.com/knadh/koanf/v2 v2.1.1
    golang.org/x/crypto v0.22.0
    github.com/jaevor/go-nanoid v1.3.0
)
```

---

## 3. Project Structure

```
pressbin/
├── main.go                  # Entry point — wire everything together
├── config.yml               # User config (not committed to binary)
│
├── internal/
│   ├── config/
│   │   └── config.go        # Load + validate config.yml
│   │
│   ├── store/
│   │   ├── store.go         # DB connection, migrations
│   │   ├── posts.go         # Post CRUD
│   │   ├── tags.go          # Tag queries
│   │   ├── keys.go          # API key CRUD
│   │   └── settings.go      # Site settings
│   │
│   ├── parser/
│   │   └── parser.go        # Markdown + front matter → HTML
│   │
│   ├── server/
│   │   ├── server.go        # chi router setup, middleware
│   │   ├── blog.go          # Public blog handlers
│   │   └── search.go        # FTS5 search handler
│   │
│   ├── sync/
│   │   └── sync.go          # POST /api/sync handler + HMAC validation
│   │
│   ├── admin/
│   │   ├── posts.go         # Admin post handlers
│   │   ├── keys.go          # API key management handlers
│   │   └── settings.go      # Settings handlers
│   │
│   ├── auth/
│   │   └── auth.go          # Key extraction, bcrypt verify, middleware
│   │
│   └── render/
│       ├── layout.go        # Base layout component
│       ├── post.go          # Single post page
│       ├── index.go         # Post list page
│       ├── tag.go           # Tag filtered page
│       └── search.go        # Search results page
│
├── assets/
│   └── style.css            # Custom CSS on top of Tailwind (embedded)
│
└── .github/
    └── workflows/
        └── sync.yml         # GitHub Action
```

---

## 4. Database Schema

Three files to create in `internal/store/store.go` as migration strings run on
startup.

```sql
-- Posts
CREATE TABLE IF NOT EXISTS posts (
    slug          TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    content_md    TEXT NOT NULL,
    content_html  TEXT NOT NULL,
    summary       TEXT DEFAULT '',
    published_at  DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    status        TEXT DEFAULT 'published'  -- published | draft
);

-- Tags (many-to-many via join)
CREATE TABLE IF NOT EXISTS tags (
    slug    TEXT NOT NULL,
    tag     TEXT NOT NULL,
    PRIMARY KEY (slug, tag),
    FOREIGN KEY (slug) REFERENCES posts(slug) ON DELETE CASCADE
);

-- Full text search (FTS5 virtual table)
CREATE VIRTUAL TABLE IF NOT EXISTS fts_index USING fts5(
    slug UNINDEXED,
    title,
    content_md,
    content=posts,
    content_rowid=rowid
);

-- FTS triggers to keep index in sync
CREATE TRIGGER IF NOT EXISTS posts_ai AFTER INSERT ON posts BEGIN
    INSERT INTO fts_index(rowid, slug, title, content_md)
    VALUES (new.rowid, new.slug, new.title, new.content_md);
END;

CREATE TRIGGER IF NOT EXISTS posts_ad AFTER DELETE ON posts BEGIN
    INSERT INTO fts_index(fts_index, rowid, slug, title, content_md)
    VALUES ('delete', old.rowid, old.slug, old.title, old.content_md);
END;

CREATE TRIGGER IF NOT EXISTS posts_au AFTER UPDATE ON posts BEGIN
    INSERT INTO fts_index(fts_index, rowid, slug, title, content_md)
    VALUES ('delete', old.rowid, old.slug, old.title, old.content_md);
    INSERT INTO fts_index(rowid, slug, title, content_md)
    VALUES (new.rowid, new.slug, new.title, new.content_md);
END;

-- API Keys
CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT PRIMARY KEY,        -- nanoid
    label        TEXT NOT NULL,           -- "github action", "mobile app"
    prefix       TEXT NOT NULL,           -- pb_admin_ or pb_sync_
    hash         TEXT NOT NULL,           -- bcrypt hash, never plaintext
    permissions  TEXT NOT NULL,           -- JSON: ["*"] or ["posts:write"]
    created_at   DATETIME NOT NULL,
    last_used_at DATETIME
);

-- Site settings (key-value)
CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Default settings
INSERT OR IGNORE INTO settings (key, value) VALUES
    ('site_title',       'My Blog'),
    ('site_description', ''),
    ('site_url',         'http://localhost:8080'),
    ('posts_per_page',   '10');
```

---

## 5. Module: Store

### `internal/store/store.go`

```go
package store

import (
    "database/sql"
    "log/slog"
    _ "modernc.org/sqlite"
)

type Store struct {
    db *sql.DB
}

func New(path string) (*Store, error) {
    db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
    if err != nil {
        return nil, err
    }
    s := &Store{db: db}
    if err := s.migrate(); err != nil {
        return nil, err
    }
    slog.Info("store ready", "path", path)
    return s, nil
}

func (s *Store) migrate() error {
    _, err := s.db.Exec(schema) // schema = the SQL above as a const string
    return err
}
```

### `internal/store/posts.go`

```go
package store

import "time"

type Post struct {
    Slug        string
    Title       string
    ContentMD   string
    ContentHTML string
    Summary     string
    PublishedAt time.Time
    UpdatedAt   time.Time
    Status      string
    Tags        []string
}

// Upsert — used by both sync API and admin API
func (s *Store) UpsertPost(p Post) error { ... }

// List — paginated, published only for public; all for admin
func (s *Store) ListPosts(page, limit int) ([]Post, int, error) { ... }

// Get single post by slug
func (s *Store) GetPost(slug string) (Post, error) { ... }

// Get posts by tag
func (s *Store) GetPostsByTag(tag string, page, limit int) ([]Post, int, error) { ... }

// Delete post
func (s *Store) DeletePost(slug string) error { ... }

// Patch status (publish/unpublish)
func (s *Store) SetPostStatus(slug, status string) error { ... }

// Full-text search via FTS5
func (s *Store) SearchPosts(query string) ([]Post, error) { ... }
```

### `internal/store/keys.go`

```go
package store

import "time"

type APIKey struct {
    ID          string
    Label       string
    Prefix      string
    Hash        string
    Permissions []string
    CreatedAt   time.Time
    LastUsedAt  *time.Time
}

func (s *Store) CreateKey(k APIKey) error               { ... }
func (s *Store) ListKeys() ([]APIKey, error)             { ... }
func (s *Store) FindKeyByRaw(raw string) (APIKey, error) { ... } // prefix match then bcrypt
func (s *Store) RevokeKey(id string) error               { ... }
func (s *Store) TouchKey(id string) error                { ... } // update last_used_at
```

---

## 6. Module: Parser

### `internal/parser/parser.go`

```go
package parser

import (
    "bytes"
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark-meta"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/renderer/html"
)

type ParseResult struct {
    Title       string
    Summary     string
    Tags        []string
    PublishedAt string
    HTML        string
}

var md = goldmark.New(
    goldmark.WithExtensions(
        extension.GFM,      // GitHub Flavoured Markdown
        extension.Footnote,
        meta.Meta,          // Front matter parsing
    ),
    goldmark.WithRendererOptions(
        html.WithUnsafe(), // Allow raw HTML in posts
    ),
)

func Parse(content []byte) (ParseResult, error) {
    var buf bytes.Buffer
    ctx := parser.NewContext()

    if err := md.Convert(content, &buf, goldmark_parser.WithContext(ctx)); err != nil {
        return ParseResult{}, err
    }

    metaData := meta.Get(ctx)

    return ParseResult{
        Title:       getString(metaData, "title"),
        Summary:     getString(metaData, "summary"),
        Tags:        getStringSlice(metaData, "tags"),
        PublishedAt: getString(metaData, "date"),
        HTML:        buf.String(),
    }, nil
}
```

### Expected Markdown front matter format

```markdown
---
title: My First Post
date: 2026-04-30
tags: [go, blog, pressbin]
summary: A short description shown in post listings.
status: published
---

# My First Post

Content starts here...
```

---

## 7. Module: Server

### `internal/server/server.go`

```go
package server

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
    store  *store.Store
    config *config.Config
    router chi.Router
}

func New(st *store.Store, cfg *config.Config) *Server {
    s := &Server{store: st, config: cfg}
    s.router = chi.NewRouter()
    s.routes()
    return s
}

func (s *Server) routes() {
    r := s.router

    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Compress(5))

    // Static assets (embedded)
    r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(assetsFS)))

    // ── Public blog routes ────────────────────────────────────────
    r.Get("/",             s.handleIndex)
    r.Get("/p/{slug}",     s.handlePost)
    r.Get("/tag/{tag}",    s.handleTag)
    r.Get("/search",       s.handleSearch)   // HTMX target
    r.Get("/feed.xml",     s.handleRSS)

    // ── Sync API (pb_sync_ key) ───────────────────────────────────
    r.Group(func(r chi.Router) {
        r.Use(s.requireAuth("posts:write"))
        r.Post("/api/sync", s.handleSync)
    })

    // ── Admin API (pb_admin_ key) ─────────────────────────────────
    r.Group(func(r chi.Router) {
        r.Use(s.requireAuth("*"))

        // Posts
        r.Get("/api/admin/posts",                 s.handleAdminListPosts)
        r.Get("/api/admin/posts/{slug}",          s.handleAdminGetPost)
        r.Put("/api/admin/posts/{slug}",          s.handleAdminUpdatePost)
        r.Delete("/api/admin/posts/{slug}",       s.handleAdminDeletePost)
        r.Patch("/api/admin/posts/{slug}/status", s.handleAdminSetStatus)

        // Keys
        r.Get("/api/admin/keys",          s.handleAdminListKeys)
        r.Post("/api/admin/keys",         s.handleAdminCreateKey)
        r.Delete("/api/admin/keys/{id}",  s.handleAdminRevokeKey)

        // Settings
        r.Get("/api/admin/settings", s.handleAdminGetSettings)
        r.Put("/api/admin/settings", s.handleAdminUpdateSettings)

        // Stats
        r.Get("/api/admin/stats", s.handleAdminStats)
    })
}

func (s *Server) Start(addr string) error {
    slog.Info("server starting", "addr", addr)
    return http.ListenAndServe(addr, s.router)
}
```

---

## 8. Module: Sync API

### `POST /api/sync`

This is the only endpoint the GitHub Action calls.

**Request**

```
POST /api/sync
Authorization: Bearer pb_sync_xxxxxxxxxxxxxxxx
Content-Type: application/json
```

```json
{
    "slug": "my-first-post",
    "title": "My First Post",
    "date": "2026-04-30",
    "tags": ["go", "blog"],
    "summary": "A short summary shown in listings.",
    "content": "# My First Post\n\nContent here...",
    "status": "published"
}
```

**Response 200**

```json
{ "status": "ok", "slug": "my-first-post", "action": "created" }
```

**Response 400**

```json
{ "error": "slug and content are required" }
```

### Handler logic

```go
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
    var req SyncRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, 400, "invalid request body")
        return
    }

    if req.Slug == "" || req.Content == "" {
        writeError(w, 400, "slug and content are required")
        return
    }

    parsed, err := parser.Parse([]byte(req.Content))
    if err != nil {
        writeError(w, 422, "failed to parse markdown")
        return
    }

    post := store.Post{
        Slug:        req.Slug,
        Title:       req.Title,
        ContentMD:   req.Content,
        ContentHTML: parsed.HTML,
        Summary:     req.Summary,
        Tags:        req.Tags,
        Status:      req.Status,
        PublishedAt: parseDate(req.Date),
        UpdatedAt:   time.Now(),
    }

    if err := s.store.UpsertPost(post); err != nil {
        writeError(w, 500, "failed to save post")
        return
    }

    writeJSON(w, 200, map[string]string{"status": "ok", "slug": req.Slug})
}
```

---

## 9. Module: Admin API

All endpoints require `Authorization: Bearer pb_admin_xxxxxxxx`.

### Posts

| Method   | Path                            | Description                       |
| -------- | ------------------------------- | --------------------------------- |
| `GET`    | `/api/admin/posts`              | List all posts including drafts   |
| `GET`    | `/api/admin/posts/:slug`        | Get single post with full content |
| `PUT`    | `/api/admin/posts/:slug`        | Update post (same body as sync)   |
| `DELETE` | `/api/admin/posts/:slug`        | Delete post permanently           |
| `PATCH`  | `/api/admin/posts/:slug/status` | Publish or unpublish              |

**GET /api/admin/posts response**

```json
{
    "posts": [
        {
            "slug": "my-first-post",
            "title": "My First Post",
            "summary": "A short summary.",
            "tags": ["go", "blog"],
            "status": "published",
            "published_at": "2026-04-30T00:00:00Z",
            "updated_at": "2026-04-30T12:00:00Z"
        }
    ],
    "total": 1,
    "page": 1
}
```

**PATCH status body**

```json
{ "status": "draft" }
```

### Keys

| Method   | Path                  | Description                         |
| -------- | --------------------- | ----------------------------------- |
| `GET`    | `/api/admin/keys`     | List all keys (hash never returned) |
| `POST`   | `/api/admin/keys`     | Create new key                      |
| `DELETE` | `/api/admin/keys/:id` | Revoke a key                        |

**POST /api/admin/keys request**

```json
{
    "label": "github action",
    "type": "sync",
    "permissions": ["posts:write"]
}
```

**POST /api/admin/keys response** — only time the raw key is ever shown

```json
{
    "id": "abc123",
    "key": "pb_sync_x7k2mN9qR4vL8wP3jT6yA1nF5eH0",
    "label": "github action",
    "note": "Save this key. It will not be shown again."
}
```

**GET /api/admin/keys response**

```json
{
    "keys": [
        {
            "id": "abc123",
            "label": "github action",
            "prefix": "pb_sync_",
            "permissions": ["posts:write"],
            "created_at": "2026-04-30T00:00:00Z",
            "last_used_at": "2026-05-01T08:00:00Z"
        }
    ]
}
```

### Settings

| Method | Path                  | Description      |
| ------ | --------------------- | ---------------- |
| `GET`  | `/api/admin/settings` | Get all settings |
| `PUT`  | `/api/admin/settings` | Update settings  |

**Settings shape**

```json
{
    "site_title": "My Blog",
    "site_description": "Writing about Go and things.",
    "site_url": "https://pressbin.dev",
    "posts_per_page": 10
}
```

### Stats

| Method | Path               | Description      |
| ------ | ------------------ | ---------------- |
| `GET`  | `/api/admin/stats` | Basic site stats |

**Stats response**

```json
{
    "total_posts": 42,
    "published_posts": 38,
    "draft_posts": 4,
    "total_tags": 15,
    "last_sync_at": "2026-05-01T08:00:00Z"
}
```

---

## 10. Module: Render

All pages are pure Go functions using gomponents. No template files, no runtime
errors.

### `internal/render/layout.go`

```go
package render

import (
    g "maragu.dev/gomponents"
    h "maragu.dev/gomponents/html"
)

type PageData struct {
    SiteTitle string
    SiteURL   string
}

func Layout(title string, pd PageData, content g.Node) g.Node {
    return h.HTML(
        h.Head(
            h.Meta(h.Charset("utf-8")),
            h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
            h.Title(g.Text(title + " — " + pd.SiteTitle)),
            h.Script(h.Src("https://cdn.tailwindcss.com")),
            h.Script(h.Src("https://unpkg.com/htmx.org@1.9.12")),
            h.Link(h.Rel("stylesheet"), h.Href("/assets/style.css")),
            h.Link(h.Rel("alternate"), h.Type("application/rss+xml"), h.Href("/feed.xml")),
        ),
        h.Body(
            h.Class("bg-gray-50 text-gray-900 min-h-screen"),
            navbar(pd),
            h.Main(h.Class("max-w-2xl mx-auto px-4 py-12"), content),
            footer(pd),
        ),
    )
}
```

### `internal/render/post.go`

```go
func PostPage(post store.Post, pd PageData) g.Node {
    return Layout(post.Title, pd,
        h.Article(
            h.H1(h.Class("text-3xl font-bold mb-2"), g.Text(post.Title)),
            h.Div(
                h.Class("text-gray-500 text-sm mb-8"),
                g.Text(post.PublishedAt.Format("January 2, 2006")),
                tagList(post.Tags),
            ),
            h.Div(
                h.Class("prose prose-gray max-w-none"),
                g.Raw(post.ContentHTML), // safe — generated by our own parser
            ),
        ),
    )
}
```

### `internal/render/index.go` — search with HTMX

```go
func searchBox() g.Node {
    return h.Input(
        h.Class("w-full border rounded px-4 py-2"),
        h.Placeholder("Search posts..."),
        hx.Get("/search"),
        hx.Target("#search-results"),
        hx.Trigger("keyup changed delay:300ms"),
    )
}
```

---

## 11. Auth & API Keys

### First Boot Bootstrap

On first run with an empty DB, generate the admin key automatically and print it
once:

```go
func (s *Store) Bootstrap() (string, error) {
    count, _ := s.countKeys()
    if count > 0 {
        return "", nil // not first run
    }

    raw := "pb_admin_" + generateSecureRandom(32)
    hash, _ := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)

    s.CreateKey(APIKey{
        ID:          nanoid.Generate(),
        Label:       "default admin",
        Prefix:      "pb_admin_",
        Hash:        string(hash),
        Permissions: []string{"*"},
        CreatedAt:   time.Now(),
    })

    return raw, nil
}
```

In `main.go`:

```go
key, _ := store.Bootstrap()
if key != "" {
    fmt.Printf(`
┌─────────────────────────────────────────────┐
│  Pressbin — First Run                       │
│                                             │
│  Admin API Key:                             │
│  %s                                         │
│                                             │
│  Save this. It will not be shown again.     │
└─────────────────────────────────────────────┘
`, key)
}
```

### Auth Middleware

```go
func (s *Server) requireAuth(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            if raw == "" {
                writeError(w, 401, "missing api key")
                return
            }

            key, err := s.store.FindKeyByRaw(raw)
            if err != nil || !key.HasPermission(permission) {
                writeError(w, 403, "forbidden")
                return
            }

            go s.store.TouchKey(key.ID) // async, non-blocking

            next.ServeHTTP(w, r)
        })
    }
}
```

### Key Lookup Strategy

Never store keys in plaintext. Lookup works in 3 steps:

1. Extract prefix from raw key (`pb_sync_` or `pb_admin_`)
2. Query DB for keys with that prefix (usually 1–3 rows)
3. `bcrypt.CompareHashAndPassword` against each hash

```go
func (s *Store) FindKeyByRaw(raw string) (APIKey, error) {
    prefix := extractPrefix(raw)
    keys, err := s.getKeysByPrefix(prefix)
    if err != nil {
        return APIKey{}, err
    }
    for _, k := range keys {
        if err := bcrypt.CompareHashAndPassword([]byte(k.Hash), []byte(raw)); err == nil {
            return k, nil
        }
    }
    return APIKey{}, errors.New("key not found")
}
```

### Permission Model

```go
func (k APIKey) HasPermission(required string) bool {
    for _, p := range k.Permissions {
        if p == "*" || p == required {
            return true
        }
    }
    return false
}
```

| Key Type | Prefix      | Permissions       | Used by       |
| -------- | ----------- | ----------------- | ------------- |
| Admin    | `pb_admin_` | `["*"]`           | You, manually |
| Sync     | `pb_sync_`  | `["posts:write"]` | GitHub Action |

### Key Rotation (zero downtime)

```
1. POST /api/admin/keys  { label: "github v2", type: "sync" }
2. Copy new key → update GitHub secret PRESSBIN_KEY
3. DELETE /api/admin/keys/:old_id
```

No restart needed.

---

## 12. Config

### `config.yml`

```yaml
server:
    port: 8080
    host: 0.0.0.0

database:
    path: ./pressbin.db

site:
    title: "My Blog"
    description: "Writing about Go and things."
    url: "https://pressbin.dev"
    posts_per_page: 10

log:
    level: info # debug | info | warn | error
```

### `internal/config/config.go`

```go
package config

import "github.com/knadh/koanf/v2"

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Site     SiteConfig
    Log      LogConfig
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

func Load(path string) (*Config, error) {
    k := koanf.New(".")
    // load yaml, set defaults, unmarshal into Config{}
    ...
}
```

---

## 13. GitHub Action

### `.github/workflows/sync.yml`

```yaml
name: Sync posts to Pressbin

on:
    push:
        branches: [main]
        paths:
            - "posts/**.md"

jobs:
    sync:
        runs-on: ubuntu-latest
        steps:
            - uses: actions/checkout@v4
              with:
                  fetch-depth: 2

            - name: Sync changed posts
              env:
                  PRESSBIN_URL: ${{ secrets.PRESSBIN_URL }}
                  PRESSBIN_KEY: ${{ secrets.PRESSBIN_KEY }}
              run: |
                  git diff --name-only HEAD~1 HEAD -- 'posts/*.md' | while read file; do
                    echo "Syncing $file..."
                    python3 .github/scripts/push.py "$file"
                  done
```

### `.github/scripts/push.py`

```python
#!/usr/bin/env python3
import sys, os, re, json, requests

def parse_frontmatter(content):
    match = re.match(r'^---\n(.*?)\n---\n', content, re.DOTALL)
    if not match:
        return {}, content
    import yaml
    meta = yaml.safe_load(match.group(1))
    body = content[match.end():]
    return meta, body

file_path = sys.argv[1]
with open(file_path, 'r') as f:
    raw = f.read()

meta, _ = parse_frontmatter(raw)
slug = meta.get('slug') or os.path.basename(file_path).replace('.md', '')

payload = {
    "slug":    slug,
    "title":   meta.get('title', slug),
    "date":    str(meta.get('date', '')),
    "tags":    meta.get('tags', []),
    "summary": meta.get('summary', ''),
    "content": raw,          # send full file including front matter
    "status":  meta.get('status', 'published'),
}

resp = requests.post(
    f"{os.environ['PRESSBIN_URL']}/api/sync",
    headers={"Authorization": f"Bearer {os.environ['PRESSBIN_KEY']}"},
    json=payload,
    timeout=30
)

if resp.status_code != 200:
    print(f"FAILED {slug}: {resp.status_code} — {resp.text}")
    sys.exit(1)

print(f"OK: {slug}")
```

### GitHub Secrets to configure

| Secret         | Value                          |
| -------------- | ------------------------------ |
| `PRESSBIN_URL` | `https://pressbin.dev`         |
| `PRESSBIN_KEY` | `pb_sync_xxxxxxxxxxxxxxxxxxxx` |

---

## 14. Build & Deploy

### Build single binary

```bash
# Local build
go build -o pressbin .

# Cross-compile for Linux server from Mac/Windows
GOOS=linux GOARCH=amd64 go build -o pressbin-linux .

# Optimised production build (smaller binary, strip debug info)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o pressbin-linux .
```

### Deploy to any Linux server

```bash
# Copy binary + config
scp pressbin-linux user@server:/opt/pressbin/pressbin
scp config.yml     user@server:/opt/pressbin/config.yml

# SSH in, run it
ssh user@server
cd /opt/pressbin
chmod +x pressbin
./pressbin
# Save the admin key printed on first run
```

### systemd service

```ini
# /etc/systemd/system/pressbin.service
[Unit]
Description=Pressbin Blog Engine
After=network.target

[Service]
Type=simple
User=pressbin
WorkingDirectory=/opt/pressbin
ExecStart=/opt/pressbin/pressbin --config /opt/pressbin/config.yml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable pressbin
systemctl start pressbin
systemctl status pressbin
```

### Nginx reverse proxy (optional)

```nginx
server {
    listen 80;
    server_name pressbin.dev www.pressbin.dev;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
    }
}
```

---

## 15. Build Order

Work through this in sequence. Each step is independently testable.

### Week 1 — Core loop working end to end

- [ ] `go mod init`, install all dependencies
- [ ] `internal/config` — load and parse config.yml
- [ ] `internal/store/store.go` — SQLite connection + schema migration on
      startup
- [ ] `internal/store/posts.go` — UpsertPost + GetPost
- [ ] `internal/parser/parser.go` — parse a .md file → HTML
- [ ] Smoke test: small main.go that reads a .md file and saves it to SQLite
- [ ] `internal/render/layout.go` — base gomponents layout
- [ ] `internal/render/post.go` — single post page
- [ ] `internal/server/server.go` — chi router, serve one post at `/p/:slug`

**Checkpoint:** `go run .` → hit `localhost:8080/p/test` → see a rendered post.

### Week 2 — Sync pipeline live

- [ ] `internal/auth/auth.go` — key lookup, bcrypt verify, middleware
- [ ] First-boot bootstrap — generate + print admin key once
- [ ] `internal/sync/sync.go` — POST /api/sync handler
- [ ] `internal/render/index.go` — post listing page at `/`
- [ ] `internal/store/tags.go` — tag queries
- [ ] `internal/render/tag.go` — tag filtered page at `/tag/:tag`
- [ ] GitHub Action + push.py
- [ ] Test full loop: push MD to GitHub → Action fires → post appears live

**Checkpoint:** Push a Markdown file → see it on the site within 30 seconds.

### Week 3 — Admin API + polish

- [ ] All admin post endpoints (list, get, update, delete, patch status)
- [ ] Key management endpoints (create, list, revoke)
- [ ] Settings endpoints (get, update)
- [ ] Stats endpoint
- [ ] FTS5 search — `SearchPosts()` in store
- [ ] `internal/render/search.go` — HTMX search box on index page
- [ ] RSS feed at `/feed.xml`
- [ ] Embed assets with `//go:embed`

**Checkpoint:** Full API working. Any UI can be built against it.

### Week 4 — Ship

- [ ] Production build script
- [ ] systemd service file
- [ ] Nginx config
- [ ] README with 5-minute setup guide
- [ ] Deploy to pressbin.dev
- [ ] Push your first real post via GitHub

---

_Built with Go. Deployed as one file. No nonsense._
