# Bruno API collection

Set `baseUrl`, `adminKey`, and `syncKey` in `environments/local.bru`.

After `make setup-dev` from `pressbin_api/`:

```bash
make keys
```

Copy the printed values into `local.bru`. Keys live in `.pressbin-dev/admin.key`
and `.pressbin-dev/sync.key` (same as consumer `~/.pressbin/`).

- **admin/** — `adminKey` → `/api/admin/*` (posts, settings, keys)
- **sync/posts** — `syncKey` → `POST /api/sync`, `DELETE /api/sync/{slug}`
- **sync/assets** — `syncKey` → asset upload/delete

Use `assetPath` like `images/foo.svg` (under `.pressbin-dev/assets/` on disk).
