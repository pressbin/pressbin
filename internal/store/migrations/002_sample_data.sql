-- Sample posts and tags for local development (safe to delete from admin or DB)

INSERT OR IGNORE INTO posts (slug, title, content_md, content_html, summary, published_at, updated_at, status) VALUES
(
    'welcome-to-pressbin',
    'Welcome to Pressbin',
    '# Welcome to Pressbin

This is a **sample post** shipped with the dev database migration.

- Single Go binary
- SQLite + Markdown
- Git-friendly content sync
',
    '<h1>Welcome to Pressbin</h1>
<p>This is a <strong>sample post</strong> shipped with the dev database migration.</p>
<ul>
<li>Single Go binary</li>
<li>SQLite + Markdown</li>
<li>Git-friendly content sync</li>
</ul>
',
    'A quick tour of what Pressbin is for.',
    '2026-04-01T12:00:00Z',
    '2026-04-01T12:00:00Z',
    'published'
),
(
    'why-sqlite',
    'Why SQLite for a blog',
    '# Why SQLite

SQLite keeps deploys boring: one file, backups are a copy, and FTS5 handles search without extra services.

```go
db, _ := sql.Open("sqlite", "blog.db")
```

That is the whole database story.
',
    '<h1>Why SQLite</h1>
<p>SQLite keeps deploys boring: one file, backups are a copy, and FTS5 handles search without extra services.</p>
<pre><code class="language-go">db, _ := sql.Open("sqlite", "blog.db")
</code></pre>
<p>That is the whole database story.</p>
',
    'No Redis, no Postgres—just a file.',
    '2026-04-15T09:30:00Z',
    '2026-04-20T14:00:00Z',
    'published'
),
(
    'draft-notes',
    'Draft: future topics',
    '# Draft notes

This post is a **draft**—it should not appear on the public index.

Ideas: RSS-only posts, series, and image pipelines.
',
    '<h1>Draft notes</h1>
<p>This post is a <strong>draft</strong>—it should not appear on the public index.</p>
<p>Ideas: RSS-only posts, series, and image pipelines.</p>
',
    'Work in progress (draft).',
    '2026-04-28T08:00:00Z',
    '2026-04-28T08:00:00Z',
    'draft'
);

INSERT OR IGNORE INTO tags (slug, tag) VALUES
    ('welcome-to-pressbin', 'pressbin'),
    ('welcome-to-pressbin', 'intro'),
    ('why-sqlite', 'sqlite'),
    ('why-sqlite', 'architecture'),
    ('draft-notes', 'draft');
