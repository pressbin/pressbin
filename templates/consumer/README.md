# Pressbin content repo

A **separate Git repository** for your Markdown posts. Pressbin itself runs elsewhere (VPS, etc.); this repo is only content plus a sync workflow.

## Repository layout

```
your-blog-content/
├── posts/
│   ├── hello-world.md
│   └── 2026/my-post.md
├── assets/
│   └── images/              # synced to Pressbin (see workflow)
├── .github/
│   └── workflows/sync.yml
└── README.md
```

Prefer the standalone template at [`pressbin_blog_template/`](../../../pressbin_blog_template/) in the stack (or “Use this template” on GitHub). You can also copy everything under `templates/consumer/` from this repo.

## Setup

1. **Install Pressbin** on your server:

   ```bash
   curl -fsSL https://raw.githubusercontent.com/pressbin/pressbin/main/scripts/install.sh | bash -s -- \
     --site-url https://your-domain.com
   ```

2. **Add GitHub repository secrets** on this content repo:

   | Secret | Value |
   |--------|--------|
   | `PRESSBIN_URL` | Your public blog URL |
   | `PRESSBIN_KEY` | Contents of `~/.pressbin/sync.key` (`pb_sync_...`) |

   Use **`sync.key` only** — admin keys cannot call the sync API.

3. Commit posts under `posts/`, images under `assets/images/`, and push to `main`.

To create extra sync keys later, use `admin.key` with `POST /api/admin/keys` and `{"label":"…","type":"sync"}`.

## Tags (taxonomy)

Pressbin uses **tags only** — there are no categories. Tags are declared per post in YAML front matter:

```yaml
---
title: Hello World
date: 2026-05-26
tags: [go, tutorial]
summary: Short blurb for the index page.
status: published
slug: hello-world
---
```

Published posts appear on `/tag/{name}`. All tags are listed at `/tags`.

## How sync works

On push to `main` when `posts/` or `assets/images/` changes, GitHub Actions:

1. Calls `GET {PRESSBIN_URL}/api/version` on **your** instance.
2. Downloads matching `pressbin-sync-linux-amd64` from [github.com/pressbin/pressbin](https://github.com/pressbin/pressbin) releases (checksum verified).
3. Runs `pressbin-sync run` on the repo checkout.

**First push:** syncs every `posts/**/*.md` and `assets/images/**` file.

**Later pushes:** deletes removed posts/images, then syncs changed or added files.

| Action in Git | Effect on Pressbin |
|---------------|-------------------|
| Edit `tags:` and push | Tags replaced for that slug |
| Add new `.md` under `posts/` | Created on sync |
| Delete `.md` | Post removed |
| Add / change file under `assets/images/` | Uploaded |
| Delete image | Removed |
| `status: draft` | Hidden from public index |

Upgrade Pressbin on your server (`install.sh` or a new release binary); the next blog push picks up the matching sync client automatically. No need to update this workflow when sync logic changes.

If your server reports version `dev`, set repository variable `PRESSBIN_SYNC_VERSION` to a release tag (e.g. `v1.0.0`).

## Post format

See `posts/example-post.md`. Use `/assets/images/...` in Markdown (root-absolute paths), not `raw.githubusercontent.com`.
