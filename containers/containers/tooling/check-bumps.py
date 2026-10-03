#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Reject recipe changes that do not bump version or epoch.

The fan-out children resolve their affected build dependencies from the
published Forgejo Alpine repository, waiting for exactly the APK file name
the owning recipe produces (`name-version-rN.apk`, epoch as pkgrel). That
barrier only separates the fresh build from the previous one if the file
name moved: a recipe edited without bumping version or epoch republishes the
same file name, and a dependent child cannot tell the new build from the old
one — it may resolve the stale package and the fan-out rebuilds nothing.

The repository's discipline already expects this: image definitions pin
exact `name=version-r epoch` versions, so a change that consumers should
pick up has to move the pin, which means moving version or epoch. This
check makes the requirement explicit at dispatch time, before any child
run is queued, instead of as a silently stale rebuild.

Runs only where a base exists (push dispatches). `republish` has no base —
everything rebuilds against whatever is currently published, and recovery
builds of unchanged versions are the point there.

  tooling/check-bumps.py --base <sha> [--head <ref>]
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path

import yaml

from planner import ROOT

NAME_PATH = re.compile(r"([a-z0-9-]+)\.yaml")


def git(root, *args):
    result = subprocess.run(
        ["git", "-C", str(root), *args], capture_output=True, text=True
    )
    if result.returncode:
        raise RuntimeError(result.stderr.strip() or f"git {args[0]} failed")
    return result.stdout


def changed_recipes(root, base, head):
    merge_base = git(root, "merge-base", base, head).strip()
    listing = git(root, "diff", "--name-only", "-z", f"{merge_base}..{head}")
    for path in listing.split("\0"):
        match = NAME_PATH.fullmatch(Path(path).name) if path else None
        if path.startswith("packages/") and match:
            yield match.group(1)


def version_epoch(root, ref, name):
    """Parse the recipe at a ref; return (document, version, epoch).

    The parsed document (comments and formatting stripped) is what a build
    consumes, so a bump is required only when it changes.
    """
    source = git(root, "show", f"{ref}:packages/{name}.yaml")
    document = yaml.safe_load(source)
    package = document["package"]
    return document, package["version"], package.get("epoch", 0)


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--base", required=True, help="base commit or ref")
    parser.add_argument("--head", default="HEAD", help="head commit or ref")
    parser.add_argument("--root", type=Path, default=ROOT)
    args = parser.parse_args()

    # An empty, all-zeros or unresolvable base (new branch, orphan history)
    # has nothing to compare against: selection treats it as everything and
    # the bump requirement cannot be evaluated — skip instead of failing.
    if not args.base or set(args.base) == {"0"}:
        print("no usable base commit; skipping the version bump check")
        return 0
    try:
        git(args.root, "rev-parse", "--verify", "--quiet", f"{args.base}^{{commit}}")
    except RuntimeError:
        print(f"base {args.base} is not a known commit; skipping the version bump check")
        return 0

    try:
        offenders = []
        for name in sorted(set(changed_recipes(args.root, args.base, args.head))):
            old_doc, old_version, old_epoch = version_epoch(args.root, args.base, name)
            new_doc, new_version, new_epoch = version_epoch(args.root, args.head, name)
            if (old_version, old_epoch) == (new_version, new_epoch) and old_doc != new_doc:
                offenders.append((name, new_version, new_epoch))
    except (RuntimeError, yaml.YAMLError, KeyError, TypeError) as error:
        parser.exit(1, f"cannot check version bumps: {error}\n")

    if offenders:
        print(
            "recipes changed without bumping version or epoch; the fan-out "
            "children would not be able to tell fresh builds from published "
            "ones:\n",
            file=sys.stderr,
        )
        for name, version, epoch in offenders:
            print(f"  packages/{name}.yaml (still {version}-r{epoch})", file=sys.stderr)
        print(
            "\nbump package.version or package.epoch in each listed recipe.",
            file=sys.stderr,
        )
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
