#!/usr/bin/env python3
import os
import re
import sys

import requests
import yaml


def parse_frontmatter(content: str):
    match = re.match(r"^---\n(.*?)\n---\n", content, re.DOTALL)
    if not match:
        return {}, content
    meta = yaml.safe_load(match.group(1)) or {}
    body = content[match.end() :]
    return meta, body


def main():
    if len(sys.argv) < 2:
        print("usage: push.py <file.md>", file=sys.stderr)
        sys.exit(2)
    file_path = sys.argv[1]
    with open(file_path, encoding="utf-8") as f:
        raw = f.read()

    meta, _ = parse_frontmatter(raw)
    slug = meta.get("slug") or os.path.basename(file_path).replace(".md", "")

    payload = {
        "slug": slug,
        "title": meta.get("title", slug),
        "date": str(meta.get("date", "")),
        "tags": meta.get("tags", []),
        "summary": meta.get("summary", ""),
        "content": raw,
        "status": meta.get("status", "published"),
    }

    base = os.environ["PRESSBIN_URL"].rstrip("/")
    resp = requests.post(
        f"{base}/api/sync",
        headers={"Authorization": f"Bearer {os.environ['PRESSBIN_KEY']}"},
        json=payload,
        timeout=30,
    )

    if resp.status_code != 200:
        print(f"FAILED {slug}: {resp.status_code} — {resp.text}")
        sys.exit(1)

    print(f"OK: {slug}")


if __name__ == "__main__":
    main()
