#!/usr/bin/env python3
"""Remove an asset from Pressbin when its file was deleted from Git."""
import os
import sys

import requests


def asset_rel_path(file_path: str) -> str:
    p = file_path.replace("\\", "/")
    if p.startswith("assets/"):
        p = p[len("assets/") :]
    return p


def main():
    if len(sys.argv) < 2:
        print("usage: delete_asset.py <deleted-file>", file=sys.stderr)
        sys.exit(2)

    rel = asset_rel_path(sys.argv[1])
    base = os.environ["PRESSBIN_URL"].rstrip("/")
    resp = requests.delete(
        f"{base}/api/sync/asset/{rel}",
        headers={"Authorization": f"Bearer {os.environ['PRESSBIN_KEY']}"},
        timeout=30,
    )

    if resp.status_code not in (200, 404):
        print(f"FAILED delete {rel}: {resp.status_code} — {resp.text}")
        sys.exit(1)

    print(f"OK: deleted {rel}")


if __name__ == "__main__":
    main()
