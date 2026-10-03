#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Select the packages and images affected by a commit range.

CI's plan job runs this against the event's base and head commits. Selection
over-approximates, never under-approximates:

  - packages/<name>.yaml changed      -> that recipe plus every recipe that
                                         transitively build-depends on it
  - images/<name>.yaml changed        -> that image
  - a local package a recipe or image installs changed -> the recipe/image
  - melange tooling changed (versions.env, tools.sh, melange patches,
    planner.py, single-package.py)    -> all packages (images follow through
                                         their local package dependencies)
  - apko tooling changed              -> all images
  - oras tooling changed              -> nothing (publish-only copy tool)
  - an un-anchorable base (empty, all-zeros, unknown, or no merge base)
                                      -> everything; selective builds are
                                         only ever an optimization

Every run also validates what selection depends on: the recipe graph
(cycles, duplicate providers, name mismatches) and every image definition
(structure; local provider resolution of its package list).

  tooling/affected.py --base <ref> [--head <ref>] [--root <dir>]
  tooling/affected.py --base <ref> --explain                  # why each item
  tooling/affected.py --base <ref> --lists-dir dist/affected  # CI loop inputs
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

import yaml

from planner import ROOT, build_graph, build_order, load_recipes

VERSION_SPLIT = re.compile(r"[<>=~@]")
NAME_PATH = re.compile(r"([a-z0-9-]+)\.yaml")
TOOL_PATCHES = re.compile(r"tooling/(melange|apko|oras)/patches/")

# Changes to these tooling files make every package (or image) suspect, no
# matter which recipes changed alongside them.
ALL_PACKAGES_TRIGGERS = (
    "tooling/versions.env",
    "tooling/tools.sh",
    "tooling/planner.py",
    "tooling/single-package.py",
)
ALL_IMAGES_TRIGGERS = (
    "tooling/single-image.sh",
)


def git(root, *args):
    result = subprocess.run(
        ["git", "-C", str(root), *args], capture_output=True, text=True
    )
    if result.returncode:
        raise RuntimeError(result.stderr.strip() or f"git {args[0]} failed")
    return result.stdout


def changed_files(root, base, head):
    """Return the changed paths between base and head, or None for everything.

    Three-dot semantics: the diff from the merge base, so a pull request is
    charged only with its own changes, not with anything that landed on the
    base branch meanwhile.
    """
    if not base or set(base) == {"0"}:
        return None
    try:
        git(root, "rev-parse", "--verify", "--quiet", f"{base}^{{commit}}")
        merge_base = git(root, "merge-base", base, head).strip()
        listing = git(root, "diff", "--name-only", "-z", f"{merge_base}..{head}")
    except RuntimeError:
        return None
    return {path for path in listing.split("\0") if path}


def load_images(root):
    """Parse every image definition; return {name: contents.packages list}.

    Doubles as the plan job's validation gate: structure errors and invalid
    names fail the run before anything builds or publishes.
    """
    images = {}
    for path in sorted(Path(root, "images").glob("*.yaml")):
        name = path.stem
        if not re.fullmatch(r"[a-z0-9-]+", name):
            raise ValueError(f"invalid image name: {name}")
        with path.open() as source:
            image = yaml.safe_load(source)
        packages = image.get("contents", {}).get("packages", [])
        if not isinstance(packages, list) or not all(
            isinstance(package, str) for package in packages
        ):
            raise ValueError(
                f"images/{name}.yaml: contents.packages must be a list of package names"
            )
        images[name] = packages
    if not images:
        raise ValueError("no image definitions found")
    return images


def local_closure(seed_packages, providers, runtimes):
    """Local recipe names providing the seed packages plus their local
    runtime closure. This is what an image needs in a local repository to
    build: whatever it installs, plus whatever those packages need at
    runtime. Build dependencies are not included — single-package.py plans
    those itself from the recipes."""
    closure = set()
    work = list(seed_packages)
    while work:
        package = VERSION_SPLIT.split(work.pop(), maxsplit=1)[0]
        recipe = providers.get(package)
        if recipe is None or recipe in closure:
            continue
        closure.add(recipe)
        work.extend(runtimes.get(package, []))
    return closure


def transitive_dependents(graph, seeds):
    """Expand seed recipes with every recipe that transitively depends on them."""
    reverse = {name: set() for name in graph}
    for name, dependencies in graph.items():
        for dependency in dependencies:
            reverse[dependency].add(name)
    affected = set(seeds)
    work = list(seeds)
    while work:
        for dependent in reverse[work.pop()]:
            if dependent not in affected:
                affected.add(dependent)
                work.append(dependent)
    return affected


def select(changed, root):
    """Map changed paths to affected packages (build order) and images with
    their local package closures. `changed` of None means everything."""
    _, providers, runtimes = load_recipes(root)
    graph = build_graph(root)
    order = build_order(root)
    images = load_images(root)

    package_reasons = {}
    image_reasons = {}
    all_packages = False
    all_images = False
    changed_packages = set()
    changed_images = set()

    if changed is None:
        all_packages = True
        all_images = True
        package_reasons["*"] = "no usable base commit; selecting everything"
        image_reasons["*"] = "no usable base commit; selecting everything"

    for path in changed or ():
        match = NAME_PATH.fullmatch(Path(path).name)
        if path.startswith("packages/") and match:
            name = match.group(1)
            if name in graph:
                changed_packages.add(name)
                package_reasons.setdefault(name, f"{path} changed")
            else:
                print(f"warning: {path} changed but no such recipe exists", file=sys.stderr)
            continue
        if path.startswith("images/") and match:
            name = match.group(1)
            if name in images:
                changed_images.add(name)
                image_reasons.setdefault(name, f"{path} changed")
            else:
                print(f"warning: {path} changed but no such image definition exists", file=sys.stderr)
            continue
        tool = TOOL_PATCHES.match(path)
        if tool and tool.group(1) == "apko":
            all_images = True
            image_reasons.setdefault("*", f"{path} changed")
            continue
        if tool and tool.group(1) == "oras":
            continue  # publish-only copy tool; tools.sh rebuilds it regardless
        if path in ALL_PACKAGES_TRIGGERS or tool:
            all_packages = True
            package_reasons.setdefault("*", f"{path} changed")
            continue
        if path in ALL_IMAGES_TRIGGERS:
            all_images = True
            image_reasons.setdefault("*", f"{path} changed")
            continue
        # Everything else (docs, workflows, CI test scripts) changes no artifact.

    if all_packages:
        affected = set(graph)
    else:
        affected = transitive_dependents(graph, changed_packages)
        for name in affected - changed_packages:
            # Any changed dependency in the closure explains the selection.
            cause = sorted(graph[name] & affected)
            package_reasons.setdefault(name, f"depends on {', '.join(cause)}")

    if all_images:
        selected_images = set(images)
    else:
        selected_images = set(changed_images)
        for name, packages in images.items():
            closure = local_closure(packages, providers, runtimes)
            hits = sorted(closure & affected)
            if hits:
                selected_images.add(name)
                image_reasons.setdefault(name, f"installs affected package {hits[0]}")

    package_list = [name for name in order if name in affected]
    image_list = []
    for name in sorted(selected_images):
        closure = local_closure(images[name], providers, runtimes)
        image_list.append({
            "name": name,
            "deps": [recipe for recipe in order if recipe in closure],
        })
    return {
        "packages": package_list,
        "images": image_list,
        "package_reasons": package_reasons,
        "image_reasons": image_reasons,
    }


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--base", default="", help="base commit or ref; empty or all-zeros selects everything")
    parser.add_argument("--head", default="HEAD", help="head commit or ref")
    parser.add_argument("--root", type=Path, default=ROOT, help="repository root")
    parser.add_argument("--explain", action="store_true", help="print why each item was selected")
    parser.add_argument("--lists-dir", type=Path, help="write packages.txt, images.txt and image-deps.txt for CI loops")
    args = parser.parse_args()

    try:
        changed = changed_files(args.root, args.base, args.head)
        result = select(changed, args.root)
    except (ValueError, KeyError, TypeError, AttributeError, OSError, yaml.YAMLError) as error:
        parser.exit(1, f"cannot select affected builds: {error}\n")

    payload = {
        "base": args.base or None,
        "head": args.head,
        "everything": changed is None,
        "packages": result["packages"],
        "images": result["images"],
    }
    print(json.dumps(payload, indent=2))
    if args.explain:
        for name, reason in result["package_reasons"].items():
            print(f"package {name}: {reason}", file=sys.stderr)
        for name, reason in result["image_reasons"].items():
            print(f"image {name}: {reason}", file=sys.stderr)
    if args.lists_dir:
        # Shell-loop inputs: one name per line; image-deps.txt prefixes each
        # image with its local package closure for PR image builds.
        lists = args.lists_dir
        lists.mkdir(parents=True, exist_ok=True)
        (lists / "packages.txt").write_text(
            "".join(f"{name}\n" for name in payload["packages"]))
        (lists / "images.txt").write_text(
            "".join(f"{entry['name']}\n" for entry in payload["images"]))
        (lists / "image-deps.txt").write_text("".join(
            f"{entry['name']}{' ' + ' '.join(entry['deps']) if entry['deps'] else ''}\n"
            for entry in payload["images"]))
    return 0


if __name__ == "__main__":
    sys.exit(main())
