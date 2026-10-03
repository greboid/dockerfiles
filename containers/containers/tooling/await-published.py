#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Wait until the published Forgejo Alpine repository holds fresh builds of
the local packages a target needs.

For a target recipe this computes the transitive local closure — its build
dependencies, their runtime closures, and its own local runtime closure —
intersects it with the affected set, and polls the repository's anonymous
HTTP endpoint for each expected APK file (name, version and epoch from the
recipes, same filenames the owning child publishes). Unaffected packages are
already published and resolve immediately; packages with no local provider
are Alpine's business.

Status: CI no longer waits this way. Polling made a leg hold a runner slot
while waiting, which deadlocked small runners (capacity 2) whenever the
dependency legs were still queued, and a failed owner surfaced only as a
full-timeout cascade. Publish runs now generate a workflow whose needs
edges ARE the dependency barrier (tooling/gen_dag.py, run by ci.yml's plan
job); this script stays for manually waiting out a local or re-run publish
against the shared repository.

Exit 0 when everything appeared within the timeout; exit 1 listing what
never did (the owning leg failed or was never scheduled).

  tooling/await-published.py golink \
    --affected dist/affected/packages.txt \
    --repo https://git.mouse-lake.ts.net/api/packages/containers/alpine/v3.24/main

For an image child the awaited set is not one recipe's closure but the
image's local package closure (which affected.py already writes into
image-deps.txt); pass it with --list instead of a target name and the same
wait applies to every recipe on it.
"""

from __future__ import annotations

import argparse
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

from planner import ROOT, build_graph, closure_of, load_recipes

NAME_RE = re.compile(r"[a-z0-9-]+")


def apk_filenames(recipe):
    """Expected APK file names for a recipe's outputs (main + subpackages).

    The filename comes from the recipe itself: name, version, epoch as pkgrel
    — exactly what the owning child builds and publishes. A changed recipe
    must bump version or epoch (tooling/check-bumps.py enforces this before
    the fan-out) so a published file is unambiguously the fresh build.
    """
    names = []
    for package in [recipe["package"], *recipe.get("subpackages", [])]:
        version = package.get("version", recipe["package"]["version"])
        epoch = package.get("epoch", recipe["package"].get("epoch", 0))
        names.append(f"{package['name']}-{version}-r{epoch}.apk")
    return names


def published(url):
    """Whether the repository serves the file, without downloading it.

    A ranged GET (two bytes when the server honors Range, a discarded body
    when it does not) instead of HEAD: the package API is only proven for
    GET — the download route is what apk itself and the round-trip check
    use — and a HEAD on an unregistered method would read as a missing
    file. 200 and 206 both mean present.
    """
    request = urllib.request.Request(url, headers={"Range": "bytes=0-0"})
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            return response.status in (200, 206)
    except urllib.error.HTTPError as error:
        if error.code == 404:
            return False
        if error.code == 416:
            # Range Not Satisfiable: the file is there, just empty.
            return True
        raise
    except urllib.error.URLError:
        # Transient network trouble counts as "not yet".
        return False


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("name", nargs="?", default=None,
                        help="target recipe name (packages/<name>.yaml)")
    parser.add_argument("--list", type=Path, default=None,
                        help="file of recipe names to await; replaces the target's closure")
    parser.add_argument("--affected", type=Path, required=True,
                        help="dist/affected/packages.txt; closure entries on it are awaited")
    parser.add_argument("--repo", required=True,
                        help="Alpine repository API URL (branch/repo level, no arch)")
    parser.add_argument("--arch", default="x86_64")
    parser.add_argument("--timeout", type=int, default=7200,
                        help="give up after this many seconds (default 7200)")
    parser.add_argument("--interval", type=int, default=20)
    args = parser.parse_args()

    if bool(args.name) == bool(args.list):
        parser.exit(2, "give a target recipe name or --list, not both\n")
    if args.list is not None:
        recipes = {
            line.strip()
            for line in args.list.read_text().splitlines()
            if line.strip()
        }
        unknown = sorted(name for name in recipes if not NAME_RE.fullmatch(name))
        if unknown:
            parser.exit(2, f"invalid recipe names in --list: {', '.join(unknown)}\n")
    else:
        if not NAME_RE.fullmatch(args.name):
            parser.exit(2, f"invalid package name: {args.name}\n")

    recipes, providers, runtimes = load_recipes()
    graph = build_graph()
    if args.list is None and args.name not in graph:
        parser.exit(1, f"no such recipe: packages/{args.name}.yaml\n")

    affected = {
        line.strip()
        for line in args.affected.read_text().splitlines()
        if line.strip()
    }

    if args.list is not None:
        closure = recipes
    else:
        closure = closure_of(args.name, graph, providers, runtimes, recipes)
    wanted = {}
    for name in sorted(closure):
        if name not in affected or name == args.name:
            continue  # unaffected: published already; target: this child builds it
        for filename in apk_filenames(recipes[name]):
            wanted[filename] = name

    if not wanted:
        print("nothing to await")
        return 0

    urls = {
        filename: f"{args.repo}/{args.arch}/{filename}" for filename in wanted
    }
    deadline = time.monotonic() + args.timeout
    print(f"awaiting {len(wanted)} published file(s):")
    for filename, name in sorted(wanted.items()):
        print(f"  {filename} (from {name})")

    while True:
        missing = [f for f in sorted(wanted) if not published(urls[f])]
        if not missing:
            print("all awaited files are published")
            return 0
        if time.monotonic() > deadline:
            print(f"timed out after {args.timeout}s waiting for:", file=sys.stderr)
            for filename in missing:
                print(f"  {filename} (from {wanted[filename]})", file=sys.stderr)
            print(
                "the owning child run failed or is still queued behind runner "
                "capacity; inspect its run in Forgejo Actions",
                file=sys.stderr,
            )
            return 1
        time.sleep(args.interval)


if __name__ == "__main__":
    sys.exit(main())
