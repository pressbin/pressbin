# Pressbin — Quick Start

## Install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/pressbin/pressbin/main/scripts/install.sh | bash -s -- \
  --site-url https://your-domain.com
export PATH="$HOME/.pressbin/bin:$PATH"
pressbin serve
```

Everything lives under `~/.pressbin/`:

| Path | Purpose |
|------|---------|
| `bin/pressbin` | Binary |
| `config.yml` | Configuration |
| `data/pressbin.db` | SQLite database |
| `assets/` | Synced images/files (served at `/assets/*`) |
| `admin.key` | Admin API key (`pb_admin_…`) — manage site, create keys |
| `sync.key` | Sync API key (`pb_sync_…`) — GitHub Actions publishing |

**Admin cannot sync. Sync cannot access admin APIs.**

## Manual install

1. Download `pressbin-linux-amd64` (or your platform) from [GitHub Releases](https://github.com/pressbin/pressbin/releases).
2. Verify with `checksums.txt` from the same release.
3. `chmod +x pressbin` and move to `~/.pressbin/bin/pressbin`.
4. Run: `pressbin setup --site-url https://your-domain.com`
5. `pressbin serve`

## Git publishing

Copy `.github/` from the [blog template](https://github.com/pressbin/pressbin-blog-template) into your content repo.

GitHub secrets:

| Secret | Value |
|--------|--------|
| `PRESSBIN_URL` | Your public site URL |
| `PRESSBIN_KEY` | Contents of `~/.pressbin/sync.key` |

## Commands

```bash
pressbin setup --site-url URL   # first-time setup
pressbin check                  # preflight validation
pressbin serve                  # run server
pressbin version
```

Upgrade: replace `~/.pressbin/bin/pressbin`, run `pressbin check`, then `pressbin serve`. Database and assets are unchanged.

Docs: https://pressbin.dev/docs
