-- Pressbin schema (baseline)

CREATE TABLE IF NOT EXISTS posts (
    slug          TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    content_md    TEXT NOT NULL,
    content_html  TEXT NOT NULL,
    summary       TEXT DEFAULT '',
    published_at  DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    status        TEXT DEFAULT 'published'
);

CREATE TABLE IF NOT EXISTS tags (
    slug    TEXT NOT NULL,
    tag     TEXT NOT NULL,
    PRIMARY KEY (slug, tag),
    FOREIGN KEY (slug) REFERENCES posts(slug) ON DELETE CASCADE
);

CREATE VIRTUAL TABLE IF NOT EXISTS fts_index USING fts5(
    slug UNINDEXED,
    title,
    content_md,
    content='posts',
    content_rowid='rowid'
);

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

CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT PRIMARY KEY,
    label        TEXT NOT NULL,
    prefix       TEXT NOT NULL,
    hash         TEXT NOT NULL,
    permissions  TEXT NOT NULL,
    created_at   DATETIME NOT NULL,
    last_used_at DATETIME
);

CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT OR IGNORE INTO settings (key, value) VALUES
    ('site_title',       'My Blog'),
    ('site_description', ''),
    ('site_url',         'http://localhost:8080'),
    ('posts_per_page',   '10');
