#!/usr/bin/env python3
"""Upload a content-repo asset to Pressbin (images under assets/images/)."""
import base64
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
        print("usage: push_asset.py <file>", file=sys.stderr)
        sys.exit(2)

    file_path = sys.argv[1]
    with open(file_path, "rb") as f:
        data = f.read()

    payload = {
        "path": asset_rel_path(file_path),
        "content_base64": base64.b64encode(data).decode("ascii"),
    }

    base = os.environ["PRESSBIN_URL"].rstrip("/")
    resp = requests.post(
        f"{base}/api/sync/asset",
        headers={"Authorization": f"Bearer {os.environ['PRESSBIN_KEY']}"},
        json=payload,
        timeout=60,
    )

    if resp.status_code != 200:
        print(f"FAILED {payload['path']}: {resp.status_code} — {resp.text}")
        sys.exit(1)

    print(f"OK: {payload['path']}")


if __name__ == "__main__":
    main()
