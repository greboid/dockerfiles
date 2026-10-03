#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Test every recipe that carries a `test:` pipeline, using the packages
already built into dist/."""

import argparse
from pathlib import Path
import re
import subprocess
import sys

import yaml


ROOT = Path(__file__).resolve().parent.parent


def testable_names(root):
    names = []
    for path in sorted((root / "packages").glob("*.yaml")):
        name = path.stem
        if not re.fullmatch(r"[a-z0-9-]+", name):
            raise ValueError(f"invalid recipe name: {name}")
        with path.open() as source:
            recipe = yaml.safe_load(source)
        if recipe["package"]["name"] != name:
            raise ValueError(f"{path}: package name must match filename")
        # Recipes without a test section (e.g. os-release, go-licenses) pass
        # trivially in melange; skip them rather than building guests for nothing.
        if recipe.get("test"):
            names.append(name)
    return names


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dry-run", action="store_true", help="print testable recipes without testing")
    args = parser.parse_args()
    try:
        names = testable_names(ROOT)
    except (ValueError, KeyError, TypeError, AttributeError, OSError, yaml.YAMLError) as error:
        parser.exit(1, f"cannot plan package tests: {error}\n")
    if not names:
        parser.exit(1, "no recipes with test pipelines found\n")
    if args.dry_run:
        print("\n".join(names))
        return 0
    for name in names:
        print(f"=== Testing {name} ===", flush=True)
        result = subprocess.run([str(ROOT / "tooling/single-test.sh"), name], cwd=ROOT)
        if result.returncode:
            print(f"Test failed: {name}; stopping", file=sys.stderr)
            return result.returncode if result.returncode > 0 else 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
