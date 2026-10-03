#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Build every local image using already-built packages in dist/."""

import argparse
from pathlib import Path
import re
import subprocess
import sys


ROOT = Path(__file__).resolve().parent.parent


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dry-run", action="store_true", help="print image names without building")
    args = parser.parse_args()
    names = [path.stem for path in sorted((ROOT / "images").glob("*.yaml"))]
    if not names:
        parser.exit(1, "no image recipes found\n")
    for name in names:
        if not re.fullmatch(r"[a-z0-9-]+", name):
            parser.exit(1, f"invalid image name: {name}\n")
    if args.dry_run:
        print("\n".join(names))
        return 0
    for name in names:
        print(f"=== Building {name} ===", flush=True)
        result = subprocess.run([str(ROOT / "tooling/single-image.sh"), name], cwd=ROOT)
        if result.returncode:
            print(f"Build failed: {name}; stopping", file=sys.stderr)
            return result.returncode if result.returncode > 0 else 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
