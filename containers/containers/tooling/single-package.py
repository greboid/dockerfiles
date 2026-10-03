#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Build one package recipe with melange, planning its local build
dependencies with the same recipe graph tooling/all-packages.py uses for
ordered builds.

  tooling/single-package.py <name>                          # local: keys/local.rsa -> dist/
  tooling/single-package.py <name> --key keys/release.rsa --out-dir dist/release
                                                  # CI: per-run key, isolated repo
  tooling/single-package.py <name> --remote-repo URL --remote-key forgejo.rsa.pub
                                                  # fan-out child: resolve deps from
                                                  # the published Forgejo repository

Keys are generated when missing — they are cheap, so every build is signed
and its test guest verifies signatures. Release runs sign with the
organisational key; a throwaway CI or local key is trusted by nothing
outside its own build and test guests.

Remote mode (--remote-repo with --remote-key) is a publish DAG job's mode:
the target recipe is the only thing this invocation builds. Local build
dependencies resolve from the published repository instead — the caller's
needs edges guarantee affected ones are already there (published by the
jobs it depends on), and unaffected ones were published by earlier runs.
This keeps each recipe built exactly once per publish run: the Forgejo
repository is the shared repository, and the DAG's needs edges are the
dependency barrier.
"""

import argparse
import os
import re
import subprocess
import sys
from pathlib import Path

from planner import ROOT, build_graph


NAME_RE = re.compile(r"[a-z0-9-]+")


def run(*cmd):
    print("+", " ".join(str(part) for part in cmd), flush=True)
    subprocess.run([str(part) for part in cmd], check=True, cwd=ROOT)


def get_melange():
    binary = ROOT / "tooling" / "bin" / "melange"
    if not os.access(binary, os.X_OK):
        run("tooling/tools.sh", "melange")
    return binary


def built(repo_arch_dir, name):
    # The repository index must exist too: an apk without an index entry is
    # invisible to dependency resolution.
    if not (repo_arch_dir / "APKINDEX.tar.gz").is_file():
        return False
    return any(repo_arch_dir.glob(f"{name}-*.apk"))


def build_package(melange, name, out_dir, key, remote_repo, remote_key):
    """Build one recipe. In local mode, first build any missing local build
dependencies into the same signed repository; in remote mode, resolve them
from the published repository (the caller awaited affected ones)."""
    repo_arch_dir = ROOT / out_dir / "x86_64"
    if remote_repo is None:
        deps = sorted(graph_cache[name])
        # Dependencies land in the same repository; each build merges its apk
        # into the out-dir's index, so later builds resolve them from there.
        for dep in deps:
            if not built(repo_arch_dir, dep):
                build_package(melange, dep, out_dir, key, None, None)

    common = [
        "--empty-workspace",
        "--arch", "x86_64",
        "--namespace", "containers",
        # Describes the recipe file itself (the repo's convention); without
        # it melange emits an invalid-NOASSERTION SPDX warning.
        "--license", "MIT",
        # Stamps the recipe's origin into its SBOM entry (no remote
        # auto-detection exists in melange v0.61).
        "--git-repo-url", "https://git.mouse-lake.ts.net/containers/containers",
        # All inputs are fetched/generated; avoid scanning repository YAML
        # (notably go-licenses.yaml) as upstream license text.
    ]
    run(
        melange, "build", f"packages/{name}.yaml", *common,
        "--out-dir", out_dir,
        "--signing-key", key,
        "--keyring-append", f"{key}.pub",
        "--repository-append", out_dir,
        *( ["--keyring-append", remote_key, "--repository-append", remote_repo]
           if remote_repo else [] ),
    )


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("name", help="package name (packages/<name>.yaml must exist)")
    parser.add_argument("--key", default="keys/local.rsa", help="signing key, generated when missing")
    parser.add_argument("--out-dir", default="dist", help="signed output repository")
    parser.add_argument("--remote-repo", default=None,
                        help="published Alpine repository URL to resolve dependencies from "
                             "(remote mode: no local dependency builds)")
    parser.add_argument("--remote-key", default=None,
                        help="public key the remote repository's index is verified against")
    args = parser.parse_args()

    if bool(args.remote_repo) != bool(args.remote_key):
        parser.exit(2, "--remote-repo and --remote-key go together\n")

    if not NAME_RE.fullmatch(args.name):
        parser.exit(2, f"invalid package name: {args.name}\n")
    if not (ROOT / "packages" / f"{args.name}.yaml").is_file():
        parser.exit(1, f"no such recipe: packages/{args.name}.yaml\n")

    global graph_cache
    graph_cache = build_graph()

    # Same epoch source as the release workflow (release commit, not wall clock).
    epoch = subprocess.run(
        ["git", "log", "-1", "--format=%ct"], check=True, capture_output=True, text=True, cwd=ROOT
    ).stdout.strip()
    os.environ["SOURCE_DATE_EPOCH"] = epoch

    if not (ROOT / args.key).is_file():
        (ROOT / args.key).parent.mkdir(parents=True, exist_ok=True)
        run(get_melange(), "keygen", args.key)

    build_package(get_melange(), args.name, args.out_dir, args.key,
                  args.remote_repo, args.remote_key)
    return 0


if __name__ == "__main__":
    sys.exit(main())
