# pressbin

Single-binary blog engine (**Go + SQLite + Markdown**) living in
`pressbin_api/`.

## Quick start (local)

```bash
cd pressbin_api
go mod tidy
make migrate
make run
```

Then open `http://localhost:8080/`.

## Live reload (Air)

```bash
cd pressbin_api
make install-air   # one-time
make dev
```

Air watches Go code, `assets/style.css`, and `config.yml` (embedded assets are
rebuilt automatically).

## Config / DB path note

`pressbin_api/config.yml` defaults to `database.path: ./pressbin.db`. This path
is resolved **relative to the config file**, not the current working directory,
so running from the monorepo root still uses the same database.

## Migrations + sample data

Migrations are embedded in the binary and are applied automatically on startup.

- SQL files live in `pressbin_api/internal/store/migrations/`
- `002_sample_data.sql` inserts a few sample posts (safe to delete)

Run migrations without starting the server:

```bash
cd pressbin_api
make migrate
```

## Releases

Pushes and PRs run [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

[`.github/workflows/release.yml`](.github/workflows/release.yml) (**Build and Release**) runs on:

- Tag push `v*` **without** `alpha` or `beta` in the name (e.g. `v1.0.0`, `v1.1.1`) → builds all platforms and publishes a GitHub Release
- Tags like `v1.0.0-beta` or `v1.0.0-alpha.1` → workflow does not run
- **workflow_dispatch** → builds with a custom embedded version (optional input); no release unless you tagged

Manual build:

```bash
# Actions → Build and Release → Run workflow
# Optional version: v1.0.0-rc1 or sha-abc1234
```

Download from [GitHub Releases](https://github.com/pressbin/pressbin/releases).
Use `config.yml.example` at the repo root as your starting config.

Local build:

```bash
make release              # all platforms → dist/
make release-bundle       # optional tarball with config + sync workflow
./bin/pressbin version    # after make build (shows dev unless ldflags set)
```

## Content sync (GitHub Action)

For a **separate content repo**, use [`../pressbin_blog_template/`](../pressbin_blog_template/)
(or copy `templates/consumer/` — kept in sync with the template). Tags live in
each post’s YAML front matter; see the template README for layout and conventions.

Secrets needed in the **content repo** on GitHub:

- `PRESSBIN_URL` (e.g. `https://pressbin.dev`)
- `PRESSBIN_KEY` (a `pb_sync_...` key created via the admin API)

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
