#!/usr/bin/env python3
"""Remove a post from Pressbin when its Markdown file was deleted from Git."""
import os
import re
import subprocess
import sys

import requests
import yaml


def parse_frontmatter(content: str):
    match = re.match(r"^---\n(.*?)\n---\n", content, re.DOTALL)
    if not match:
        return {}
    return yaml.safe_load(match.group(1)) or {}


def slug_from_path(file_path: str) -> str:
    meta = {}
    try:
        out = subprocess.run(
            ["git", "show", f"HEAD~1:{file_path}"],
            capture_output=True,
            text=True,
            check=True,
        )
        meta = parse_frontmatter(out.stdout)
    except subprocess.CalledProcessError:
        pass
    return meta.get("slug") or os.path.basename(file_path).replace(".md", "")


def main():
    if len(sys.argv) < 2:
        print("usage: delete.py <deleted-file.md>", file=sys.stderr)
        sys.exit(2)

    file_path = sys.argv[1]
    slug = slug_from_path(file_path)

    base = os.environ["PRESSBIN_URL"].rstrip("/")
    resp = requests.delete(
        f"{base}/api/sync/{slug}",
        headers={"Authorization": f"Bearer {os.environ['PRESSBIN_KEY']}"},
        timeout=30,
    )

    if resp.status_code not in (200, 404):
        print(f"FAILED delete {slug}: {resp.status_code} — {resp.text}")
        sys.exit(1)

    print(f"OK: deleted {slug}")


if __name__ == "__main__":
    main()
