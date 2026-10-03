#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Plan local recipe dependencies, then use the normal signed package builder."""

import argparse
import subprocess
import sys

import yaml

from planner import ROOT, build_order


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dry-run", action="store_true", help="print build order without building")
    args = parser.parse_args()
    try:
        order = build_order(ROOT)
    except (ValueError, KeyError, TypeError, AttributeError, OSError, yaml.YAMLError) as error:
        parser.exit(1, f"cannot plan package builds: {error}\n")
    if args.dry_run:
        print("\n".join(order))
        return 0
    for name in order:
        print(f"=== Building {name} ===", flush=True)
        result = subprocess.run([str(ROOT / "tooling/single-package.py"), name], cwd=ROOT)
        if result.returncode:
            print(f"Build failed: {name}; stopping", file=sys.stderr)
            return result.returncode if result.returncode > 0 else 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
