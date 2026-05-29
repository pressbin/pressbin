# pressbin

Single-binary blog engine (**Go + SQLite + Markdown**) living in
`pressbin_api/`.

## Install (user, no root)

```bash
curl -fsSL https://raw.githubusercontent.com/pressbin/pressbin/main/scripts/install.sh | bash -s -- \
  --site-url https://blog.example.com
```

Installs to `~/.pressbin/` (binary, config, database, assets, `admin.key`, `sync.key`), then run:

```bash
export PATH="$HOME/.pressbin/bin:$PATH"
pressbin serve
```

Custom install directory:

```bash
curl -fsSL .../install.sh | bash -s -- \
  --site-url https://blog.example.com \
  --home ~/domains/blog.example.com/.pressbin
export PATH="$HOME/domains/blog.example.com/.pressbin/bin:$PATH"
pressbin serve   # loads …/.pressbin/config.yml from the binary path
```

Or explicitly: `pressbin serve --config ~/path/.pressbin/config.yml`.

- **Admin key** (`admin.key`): manage posts/settings and create keys via `/api/admin/*`. Cannot sync.
- **Sync key** (`sync.key`): GitHub Actions / `POST /api/sync*`. Cannot access admin APIs.

## Quick start (local dev, monorepo)

Dev data lives only in **`.pressbin-dev/`** — same layout as consumer `~/.pressbin/`:

```text
.pressbin-dev/
  config.yml
  data/pressbin.db
  assets/
  admin.key
  sync.key
```

```bash
cd pressbin_api
make setup-dev    # once (or make reset-dev to wipe)
make dev          # Air → http://127.0.0.1:8080
make keys         # print Bruno adminKey / syncKey
```

The API repo no longer uses `config.yml` or `pressbin.db` in the project root. Sync posts/assets into `.pressbin-dev` via Bruno or the blog repo’s GitHub Action pointed at your local server.

## Live reload (Air)

```bash
make install-air   # one-time
make dev
```

Air runs `pressbin serve --config .pressbin-dev/config.yml`.

## Configuration

Use **either** an optional `config.yml` **or** `PRESSBIN_*` environment variables
(env wins when both are set). No `.env` file is read — set vars in your shell,
systemd unit, or container.

| Variable | Purpose |
|----------|---------|
| `PRESSBIN_SERVER_PORT` | Listen port (default `8080`) |
| `PRESSBIN_SERVER_HOST` | Bind address (default `0.0.0.0`) |
| `PRESSBIN_DATABASE_PATH` | SQLite file path |
| `PRESSBIN_SITE_TITLE` | Site title (also synced to DB on start) |
| `PRESSBIN_SITE_DESCRIPTION` | Site description |
| `PRESSBIN_SITE_URL` | Public URL (RSS, etc.) |
| `PRESSBIN_SITE_POSTS_PER_PAGE` | Posts per index page |
| `PRESSBIN_ASSETS_PATH` | Blog assets directory (upload + serve at `/assets/*`) |
| `PRESSBIN_LOG_LEVEL` | `debug`, `info`, `warn`, `error` |

Local dev: use `make setup-dev` (not `config.yml` in the repo root). See `config.yml.example` for production layout reference.

CLI: `pressbin setup`, `pressbin check`, `pressbin serve` (default), `pressbin version`.

Default config resolution: `config.yml` next to the install tree (`…/bin/pressbin` →
`…/config.yml`), then `.pressbin-dev/`, `~/.pressbin/`, then `./config.yml`.

Production example (no config file):

```bash
export PRESSBIN_DATABASE_PATH=/var/lib/pressbin/data/pressbin.db
export PRESSBIN_ASSETS_PATH=/var/lib/pressbin/assets
export PRESSBIN_SITE_URL=https://blog.example.com
pressbin setup --site-url "$PRESSBIN_SITE_URL" --home /var/lib/pressbin
pressbin serve
```

Relative paths in YAML are resolved from the **config file’s directory**; without a
file, they resolve from the **process working directory**. Prefer absolute paths in env.

On every startup, resolved `site.*` values are written into the `settings` table.

## Migrations

Migrations are embedded in the binary and are applied automatically on startup.
SQL files live in `pressbin_api/internal/store/migrations/`. Posts are not seeded in
the database — sync them from your GitHub content repo so the DB stays 1:1 with Git.
Images sync via `POST /api/sync/asset` (same GitHub Action as posts). Set a writable
`PRESSBIN_ASSETS_PATH` (recommended) or `assets.path` in config. Must not be the binary or database directory.

Run migrations without starting the server:

```bash
cd pressbin_api
make migrate
```

## Releases

Pushes and PRs run [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

[`.github/workflows/release.yml`](.github/workflows/release.yml) (**Build and Release**) runs on:

- Tag push `v*` **without** `alpha` or `beta` in the name (e.g. `v1.0.0`, `v1.1.1`) → builds all platforms and publishes a GitHub Release
- Tags like `v1.0.0-beta` → workflow queues but skips the build (no release)
- **workflow_dispatch** on branch `main` → manual build; optional version input; no GitHub Release unless you pushed a tag

**If the UI says “Failed to queue workflow run”:** use branch `main`, enable Actions under Settings → Actions → General, push a stable tag (`git push origin v1.0.0`), or retry after a minute (GitHub glitch). Do not use a tag name containing `alpha` or `beta` if you expect a release.

Manual build:

```bash
# Actions → Build and Release → Run workflow
# Optional version: v1.0.0-rc1 or sha-abc1234
```

Download from [GitHub Releases](https://github.com/pressbin/pressbin/releases):
`pressbin-{os}-{arch}` (server) and `pressbin-sync-{os}-{arch}` (content-repo CI).
Use `config.yml.example` at the repo root as your starting config.

Local build:

```bash
make release              # all platforms → dist/
make release-bundle       # optional tarball with config + sync workflow
./bin/pressbin version    # after make build (shows dev unless ldflags set)
```

## Content sync (GitHub Action)

Content lives in a **separate Git repo** — use [`../pressbin_blog_template/`](../pressbin_blog_template/)
(or copy `templates/consumer/`). Tags are per-post YAML front matter; see the template README.

**Secrets** on the content repo:

| Secret | Value |
|--------|--------|
| `PRESSBIN_URL` | Your public blog URL (your instance, e.g. `https://blog.example.com`) |
| `PRESSBIN_KEY` | `pb_sync_...` from `sync.key` or `POST /api/admin/keys` |

**On push to `main`**, Actions:

1. `GET {PRESSBIN_URL}/api/version` on your server
2. Download matching `pressbin-sync-linux-amd64` from [GitHub Releases](https://github.com/pressbin/pressbin/releases) (checksum verified)
3. Run `pressbin-sync run` (incremental git diff)

**Manual full resync:** Actions → “Sync content to Pressbin” → **Run workflow** →
`pressbin-sync run --all` (every post and `assets/images/**` file).

Local full sync (same as manual CI):

```bash
export PRESSBIN_URL=https://blog.example.com
export PRESSBIN_KEY=pb_sync_...
pressbin-sync run --all /path/to/content-repo
```

Server exposes `GET /api/version` (public) so CI picks the matching sync client release.
Upgrade the server binary only; blog workflows stay thin and rarely need changes.

## Admin API

On first run, the server prints an **admin key** (`pb_admin_...`) once. Use it
as:

```
Authorization: Bearer pb_admin_...
```

Key endpoints:

- `POST /api/admin/keys` with `{ "label": "...", "type": "sync" }` to create a
  `pb_sync_...` key
- `GET /api/admin/posts`, `PUT /api/admin/posts/{slug}`, etc.
