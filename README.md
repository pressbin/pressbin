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

## Content sync (GitHub Action)

The repo includes:

- `pressbin_api/.github/workflows/sync.yml`
- `pressbin_api/.github/scripts/push.py`

Secrets needed in GitHub:

- `PRESSBIN_URL` (e.g. `https://pressbin.in`)
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
