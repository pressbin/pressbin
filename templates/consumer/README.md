# Pressbin content repo

A **separate Git repository** for your Markdown posts. Pressbin itself runs elsewhere (VPS, etc.); this repo is only content plus a sync workflow.

## Repository layout

```
your-blog-content/
├── posts/
│   ├── hello-world.md
│   └── 2026/my-post.md          # nested paths are fine
├── .github/
│   ├── workflows/
│   │   └── sync.yml
│   └── scripts/
│       ├── push.py
│       └── delete.py
└── README.md                     # optional (this file)
```

Prefer the standalone template at [`pressbin_blog_template/`](../../../pressbin_blog_template/) in the stack (or “Use this template” on GitHub). You can also copy everything under `templates/consumer/` from this repo.

## Setup

1. Deploy Pressbin and save the **admin key** (`pb_admin_...`) from first run.
2. Create a sync API key:

   ```bash
   curl -s -X POST "$PRESSBIN_URL/api/admin/keys" \
     -H "Authorization: Bearer $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"label":"github action","type":"sync"}'
   ```

3. Add GitHub repository secrets:

   | Secret | Value |
   |--------|--------|
   | `PRESSBIN_URL` | Base URL (e.g. `https://yourdomain.com`) |
   | `PRESSBIN_KEY` | The `pb_sync_...` key from step 2 |

4. Commit posts under `posts/` and push to `main`.

## Tags (taxonomy)

Pressbin uses **tags only** — there are no categories. Tags are not stored in a central file; each post declares them in YAML front matter:

```yaml
---
title: Hello World
date: 2026-05-26
tags: [go, tutorial]
summary: Short blurb for the index page.
status: published
slug: hello-world          # optional; defaults to filename without .md
---
```

**Conventions:**

- Use short, lowercase tags (the server normalizes to lowercase).
- Edit `tags` in the file and push — the live site updates on the next sync.
- Optionally pick one tag as a broad “topic” (e.g. `dev`, `life`); that is convention only.
- Set `slug` explicitly on important posts so renames do not change the URL.

Published posts appear on `/tag/{name}`. All tags are listed at `/tags` on your site.

## How sync works

On push to `main` (when `posts/**/*.md` changes):

1. **First push** (no parent commit): syncs every `posts/**/*.md` file.
2. **Later pushes**: deletes removed posts, then syncs changed or added files.

| Action in Git | Effect on Pressbin |
|---------------|-------------------|
| Edit `tags:` and push | Tags replaced for that slug |
| Add new `.md` under `posts/` | Created on sync |
| Delete `.md` | Post removed via `delete.py` |
| Rename file without `slug:` | New slug; old post remains until you delete it in Git or via admin API |
| `status: draft` | Stored but hidden from index and tag pages |

## Post format

See `posts/example-post.md`. The filename is the slug when `slug` is omitted in front matter.
