#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Generate the publish DAG workflow from the affected selection.

One job per affected package, with needs edges equal to the affected part
of its local dependency closure; one job per affected image, needing
exactly its closure packages' jobs. The needs edges ARE the dependency
barrier: Forgejo dispatches a job only once every need has succeeded, so a
leg never holds a runner slot while waiting (the old ready-poll held one
and deadlocked a capacity-2 runner whenever the owner legs were still
queued), and a failed leg skips its dependents instantly instead of
cascading through 30-minute polling timeouts.

Why generation instead of a committed workflow: needs is resolved at
scheduling time over statically-named jobs, while the node set is the
per-commit affected list. ci.yml's plan job runs this generator, commits
the result to a per-run branch (ci/dag-<run_id>), and pushes it; the push
triggers the generated workflow, whose exact edges are only knowable at
run time. Each leg checks out that branch — identical recipes to the
pushed SHA (the extra commit adds only this workflow file) — and tags
images with the original SHA passed through env.SHA.

The readiness signal is publication itself: a need's job completes only
after its Publish and round-trip steps succeeded, so dependents resolve
their affected dependencies from the Forgejo Alpine repository without
any polling. Everything unaffected is already published.

  tooling/gen_dag.py --lists-dir dist/affected --scope all \
    --sha "$SHA" --branch "ci/dag-123" \
    --out .forgejo/workflows/publish-dag.yml

With nothing affected the generator writes no output file and exits 0;
the caller treats a missing file as "nothing to publish".
"""

from __future__ import annotations

import argparse
import copy
import re
import sys
from pathlib import Path

import yaml

from planner import ROOT, build_graph, closure_of, load_recipes, build_order

NAME_RE = re.compile(r"[a-z0-9-]+")

OWNER = "containers"
ALPINE_REPO_URL = "https://git.mouse-lake.ts.net/api/packages/containers/alpine/v3.24/main"
ALPINE_KEY_URL = "https://git.mouse-lake.ts.net/api/packages/containers/alpine/key"
OIDC_AUDIENCE = "u:1:ea30cbb9-71af-4074-9cf7-603f4e77208f"

CHECKOUT_STEP = {
    "name": "Checkout",
    # Full history: single-package.py and the tools build read the recipe
    # graph from the workspace; keep the same shape as every other job.
    "uses": "actions/toolkit/git-checkout@master",
    "with": {"fetch-depth": "0"},
}

PACKAGE_STEPS = [
    CHECKOUT_STEP,
    {
        "name": "Install uv and bubblewrap (shared tooling scripts)",
        # uv runs the shared tooling scripts (BusyBox env cannot exec their
        # `#!/usr/bin/env -S` shebangs); bubblewrap is melange's default
        # container runner for builds and tests.
        "run": "apk add --no-cache uv bubblewrap\n",
    },
    {
        "name": "Set up Go",
        # The job image has no Go toolchain; go-setup installs one into the
        # workspace and exports PATH/GOROOT for the tools build.
        "uses": "actions/toolkit/go-setup@master",
        "with": {"go-version": "stable"},
    },
    {
        "name": "Build tools",
        # Fresh from source at the recorded version, like every run; a
        # package leg only needs melange.
        "run": "tooling/tools.sh melange\n",
    },
    {
        "name": "Write release signing key (APK_SIGNING_KEY)",
        "env": {"APK_SIGNING_KEY": "${{ secrets.APK_SIGNING_KEY }}"},
        "run": """set -eu
mkdir -p keys
printf '%s\\n' "$APK_SIGNING_KEY" > keys/release.rsa
chmod 600 keys/release.rsa
command -v openssl >/dev/null 2>&1 || apk add --no-cache openssl
# melange and the test guests take the public key alongside the private one.
openssl rsa -in keys/release.rsa -pubout -out keys/release.rsa.pub 2>/dev/null
head -1 keys/release.rsa.pub | grep -q 'BEGIN PUBLIC KEY'
""",
    },
    {
        "name": "Fetch Forgejo repository signing key",
        # The published repository's index is signed by Forgejo itself; its
        # live public key is an explicit trust input for guests resolving
        # dependencies from there. The key endpoint is organisation-level.
        # melange matches index signatures to keyring entries by FILE NAME
        # (same as apko), and Forgejo signs the index as
        # .SIGN.RSA.<owner>@<sha256(key DER)>.rsa.pub — present the same key
        # under the name its signature carries, like the image jobs do.
        "run": """set -eu
curl -fsSL --retry 3 \\
  -o forgejo-alpine-key.rsa.pub "$ALPINE_KEY_URL" || {
    echo "key fetch failed: $ALPINE_KEY_URL" >&2
    exit 1
  }
head -1 forgejo-alpine-key.rsa.pub | grep -q 'BEGIN PUBLIC KEY'
command -v openssl >/dev/null 2>&1 || apk add --no-cache openssl
keyhash=$(openssl pkey -pubin -in forgejo-alpine-key.rsa.pub \\
  -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1)
printf '%s\\n' "${OWNER}@${keyhash}.rsa.pub" > forgejo-key-signature-name
cp forgejo-alpine-key.rsa.pub "$(cat forgejo-key-signature-name)"
""",
    },
    {
        "name": "Build and test package",
        # Needs edges guaranteed the affected dependencies are published;
        # everything else resolves from the repository as-is
        # (single-package.py --remote-repo). Recipes without a test section
        # pass trivially; tests run under bwrap like the builds: the docker
        # runner would bind-mount a workspace under this container's /tmp,
        # a path the daemon cannot see.
        "run": """set -eu
registry_key=$(cat forgejo-key-signature-name)
uv run --script tooling/single-package.py "$PACKAGE" \\
  --key keys/release.rsa \\
  --out-dir dist/release \\
  --remote-repo "$ALPINE_REPO_URL" \\
  --remote-key "$registry_key"
tooling/bin/melange test "packages/$PACKAGE.yaml" \\
  --arch x86_64 \\
  --keyring-append "$PWD/keys/release.rsa.pub" \\
  --keyring-append "$PWD/$registry_key" \\
  --repository-append "$PWD/dist/release" \\
  --repository-append "$ALPINE_REPO_URL"
""",
    },
    {
        "name": "Publish package (RELEASE_PACKAGE_TOKEN)",
        "env": {"RELEASE_PACKAGE_TOKEN": "${{ secrets.RELEASE_PACKAGE_TOKEN }}"},
        "run": """set -eu
# Match only the requested package: build dependencies share the
# published repository but are not this leg's artifacts.
# Publishing is the readiness signal dependent jobs' needs wait on.
for apk_file in dist/release/x86_64/${PACKAGE}-*.apk; do
  [ -e "$apk_file" ] || continue
  echo "publishing: $apk_file"
  sha256sum "$apk_file"
  curl -fsS --retry 3 \\
    --user "greboid:${RELEASE_PACKAGE_TOKEN}" \\
    -H "Content-Type: multipart/form-data" \\
    -F "file=@${apk_file}" \\
    -X PUT \\
    "$ALPINE_REPO_URL"
done
""",
    },
    {
        "name": "Verify anonymous round trip",
        "run": """set -eu
for apk_file in dist/release/x86_64/${PACKAGE}-*.apk; do
  [ -e "$apk_file" ] || continue
  want=$(sha256sum "$apk_file" | awk '{print $1}')
  name=$(basename "$apk_file")
  curl -fsSL -o roundtrip.apk "$ALPINE_REPO_URL/x86_64/$name"
  got=$(sha256sum roundtrip.apk | awk '{print $1}')
  test "$got" = "$want"
  echo "published ${name}: sha256:${got}"
done
""",
    },
]

IMAGE_STEPS = [
    CHECKOUT_STEP,
    {
        "name": "Set up Go",
        "uses": "actions/toolkit/go-setup@master",
        "with": {"go-version": "stable"},
    },
    {
        "name": "Build tools",
        # Publish legs need apko and oras.
        "run": "tooling/tools.sh\n",
    },
    {
        "name": "Fetch Forgejo repository signing key",
        # apko trusts the live repository key Forgejo signs the index with,
        # never a key an individual build signed with. apko matches index
        # signatures to keyring entries by FILE NAME, and Forgejo signs the
        # index as .SIGN.RSA.<owner>@<sha256(key DER)>.rsa.pub — so present
        # the same key under the name its signature carries.
        "run": """set -eu
curl -fsSL --retry 3 \\
  -o forgejo-alpine-key.rsa.pub "$ALPINE_KEY_URL" || {
    echo "key fetch failed: $ALPINE_KEY_URL" >&2
    exit 1
  }
head -1 forgejo-alpine-key.rsa.pub | grep -q 'BEGIN PUBLIC KEY'
command -v openssl >/dev/null 2>&1 || apk add --no-cache openssl
keyhash=$(openssl pkey -pubin -in forgejo-alpine-key.rsa.pub \\
  -outform DER 2>/dev/null | sha256sum | cut -d' ' -f1)
printf '%s\\n' "${OWNER}@${keyhash}.rsa.pub" > forgejo-key-signature-name
cp forgejo-alpine-key.rsa.pub "$(cat forgejo-key-signature-name)"
""",
    },
    {
        "name": "Log in to the Forgejo registry (authorized integration)",
        # Mint an ID token for this Authorized Integration and log in to
        # this instance's registry with it; credentials land in
        # .actions/docker/config.json for the publish step.
        "uses": "actions/toolkit/docker-login@master",
        "with": {"audience": OIDC_AUDIENCE},
    },
    {
        "name": "Publish image (authorized integration)",
        # The docker-login step wrote the config; this step has a fresh
        # shell, so point at it explicitly. TAG is the ORIGINAL pushed SHA
        # (env.SHA), not the DAG branch head, which only adds this file.
        "run": """set -eu
export DOCKER_CONFIG="$PWD/.actions/docker"
test -s "$DOCKER_CONFIG/config.json"
registry="git.mouse-lake.ts.net/${OWNER}/${IMAGE}"
out=$(tooling/bin/apko publish \\
  -r "$ALPINE_REPO_URL" \\
  -k "./$(cat forgejo-key-signature-name)" \\
  "images/${IMAGE}.yaml" \\
  "${registry}:${TAG}")
echo "$out"
digest=$(printf '%s\\n' "$out" | grep -o 'sha256:[0-9a-f]*' | tail -1)
test -n "$digest"
echo "image published: ${registry}@${digest}"
# Advance latest in the same leg: the digest is known here, there is no
# cross-job handoff.
tooling/bin/oras cp "${registry}@${digest}" "${registry}:latest"
""",
    },
]


def read_lists(lists_dir: Path):
    """packages.txt (build order), images.txt, and image-deps.txt entries.

    image-deps.txt lines are "image dep1 dep2 ..." — the image's full local
    package closure in build order, as written by affected.py.
    """
    packages = [
        line.strip()
        for line in (lists_dir / "packages.txt").read_text().splitlines()
        if line.strip()
    ]
    images = [
        line.strip()
        for line in (lists_dir / "images.txt").read_text().splitlines()
        if line.strip()
    ]
    deps = {}
    for line in (lists_dir / "image-deps.txt").read_text().splitlines():
        if not line.strip():
            continue
        tokens = line.split()
        deps[tokens[0]] = tokens[1:]
    return packages, images, deps


def plan_dag(lists_dir: Path, root: Path = ROOT, scope: str = "all"):
    """Compute the DAG: {package: needs} and {image: needs}.

    A package job needs the affected part of its local closure — the same
    set the previous design polled for; here it becomes needs edges. An
    image job needs its closure's affected packages (from image-deps.txt).
    Unaffected dependencies are already published and appear in no edge.
    """
    if scope not in ("all", "packages", "images"):
        raise ValueError(f"unknown scope: {scope}")
    packages, images, image_deps = read_lists(lists_dir)
    affected = set(packages)

    recipes, providers, runtimes = load_recipes(root)
    graph = build_graph(root)
    order = build_order(root)
    for name in packages:
        if name not in graph:
            raise ValueError(f"affected package has no recipe: {name}")
    position = {name: index for index, name in enumerate(order)}

    package_needs = {}
    for name in packages:
        closure = closure_of(name, graph, providers, runtimes, recipes)
        package_needs[name] = sorted(
            (dep for dep in closure if dep in affected and dep != name),
            key=lambda dep: position[dep],
        )

    image_needs = {}
    for name in images:
        if not NAME_RE.fullmatch(name):
            raise ValueError(f"invalid image name: {name}")
        if name not in image_deps:
            raise ValueError(f"image missing from image-deps.txt: {name}")
        image_needs[name] = sorted(
            (dep for dep in image_deps[name] if dep in affected),
            key=lambda dep: position[dep],
        )

    # Scope trimming drops whole sides of the DAG. A trimmed-away package
    # job that images still need cannot be waited on: drop those edges and
    # say so loudly — the image will fail resolution against the repository
    # unless the packages were already published.
    if scope == "packages":
        image_needs = {}
    if scope == "images":
        for name, needs in image_needs.items():
            if needs:
                print(
                    f"warning: scope=images drops edges to unbuilt packages "
                    f"for image {name}: {', '.join(needs)}",
                    file=sys.stderr,
                )
            image_needs[name] = []
        package_needs = {}

    # Edge integrity and acyclicity: cheap to prove here, and a mistake
    # would otherwise surface as an invalid workflow on the Forgejo side.
    # Job ids must mirror build_workflow exactly (image jobs prefixed);
    # otherwise an image sharing a package's name would read as a self-edge.
    # Deep-copy the edge lists: the check consumes edges as it goes and
    # must not mutate the plan it validates.
    edges = {
        **{name: list(needs) for name, needs in package_needs.items()},
        **{f"image-{name}": list(needs) for name, needs in image_needs.items()},
    }
    job_ids = set(edges)
    for name, needs in edges.items():
        missing = [dep for dep in needs if dep not in job_ids]
        if missing:
            raise ValueError(f"job {name} needs unknown jobs: {', '.join(missing)}")
    unresolved = dict(edges)
    while unresolved:
        ready = [name for name, needs in unresolved.items() if not needs]
        if not ready:
            raise ValueError(
                "dependency cycle in generated DAG: " + ", ".join(sorted(unresolved))
            )
        for name in ready:
            del unresolved[name]
        for needs in unresolved.values():
            for name in ready:
                if name in needs:
                    needs.remove(name)

    return {"packages": package_needs, "images": image_needs}


def build_workflow(plan, branch: str, sha: str) -> dict:
    """The generated workflow as a plain dict (yaml-dumpable).

    Image jobs are keyed image-<name>: package and image names overlap in
    this repository (miniflux, golink, ...), so both sides can be affected
    without colliding on a job id. The env inside an image job carries the
    raw name — the publish steps resolve images/<name>.yaml with it.
    """
    jobs = {}
    # Deep-copy the step templates: identical objects would serialize as
    # YAML anchors/aliases, which workflow parsers need not support.
    for name, needs in plan["packages"].items():
        job = {
            "runs-on": "alpine",
            "env": {"PACKAGE": name, "OWNER": OWNER},
            "steps": copy.deepcopy(PACKAGE_STEPS),
        }
        if needs:
            job = {"needs": needs, **job}
        jobs[name] = job
    for name, needs in plan["images"].items():
        job = {
            "runs-on": "alpine",
            "env": {"IMAGE": name, "TAG": "${{ env.SHA }}"},
            "steps": copy.deepcopy(IMAGE_STEPS),
        }
        if needs:
            job = {"needs": needs, **job}
        jobs[f"image-{name}"] = job
    if not jobs:
        raise ValueError("refusing to emit an empty workflow")
    jobs["cleanup"] = {
        "needs": list(jobs),
        # A failed build must not keep the branch alive either.
        "if": "always()",
        "runs-on": "alpine",
        "steps": [
            {
                "name": "Delete the DAG branch",
                "env": {"RELEASE_PACKAGE_TOKEN": "${{ secrets.RELEASE_PACKAGE_TOKEN }}"},
                "run": """set -eu
api="$GITHUB_SERVER_URL/api/v1/repos/$GITHUB_REPOSITORY"
# The branch name carries a slash (ci/dag-<run id>): the git refs
# endpoint cannot route it (405), the branches endpoint takes it
# percent-encoded.
if ! curl -fsS -X DELETE \\
    -H "Authorization: token $RELEASE_PACKAGE_TOKEN" \\
    "$api/branches/${DAG_BRANCH//\\//%2F}"; then
  echo "warning: could not delete $DAG_BRANCH (already deleted?)"
fi
""",
            }
        ],
    }
    return {
        "on": {"push": {"branches": [branch]}, "workflow_dispatch": None},
        "enable-openid-connect": True,
        "env": {
            "OWNER": OWNER,
            "ALPINE_REPO_URL": ALPINE_REPO_URL,
            "ALPINE_KEY_URL": ALPINE_KEY_URL,
            # The ORIGINAL pushed commit: images tag with it, and it is
            # what the DAG builds — the branch head only adds this file.
            "SHA": sha,
            "DAG_BRANCH": branch,
        },
        "jobs": jobs,
    }


class _LiteralDumper(yaml.SafeDumper):
    pass


def _represent_str(dumper, data):
    if "\n" in data:
        return dumper.represent_scalar("tag:yaml.org,2002:str", data, style="|")
    return dumper.represent_scalar("tag:yaml.org,2002:str", data)


_LiteralDumper.add_representer(str, _represent_str)


def render_workflow(workflow: dict, branch: str) -> str:
    text = yaml.dump(
        workflow, Dumper=_LiteralDumper, sort_keys=False, default_flow_style=False,
        width=10**6,
    )
    # Quote the trigger key: harmless for Forgejo (whose parser reads `on`
    # as a string either way), and it keeps the file honest under YAML 1.1
    # parsers where a bare `on` is a boolean.
    text = re.sub(r"(?m)^on:", '"on":', text)
    header = (
        f"# Generated by tooling/gen_dag.py from the affected selection — DO NOT EDIT.\n"
        f"# Triggered by the push to {branch}. One job per affected package; needs\n"
        f"# edges from the recipe graph are the dependency barrier. One job per\n"
        f"# affected image, needing exactly its closure packages' jobs. The cleanup\n"
        f"# job removes the branch afterwards; re-run a failed leg from its run page,\n"
        f"# or re-dispatch ci.yml (manual dispatch IS the republish path).\n"
    )
    return header + text


def main():
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--lists-dir", type=Path, required=True,
                        help="dist/affected with packages.txt, images.txt, image-deps.txt")
    parser.add_argument("--scope", default="all",
                        choices=("all", "packages", "images"),
                        help="trim a side of the DAG (manual dispatch scope)")
    parser.add_argument("--sha", required=True,
                        help="original pushed commit (embedded as env.SHA)")
    parser.add_argument("--branch", required=True,
                        help="branch the generated workflow is pushed to and triggered by")
    parser.add_argument("--root", type=Path, default=ROOT, help="repository root")
    parser.add_argument("--out", type=Path, required=True,
                        help="workflow file to write; not written when nothing is affected")
    args = parser.parse_args()

    try:
        plan = plan_dag(args.lists_dir, args.root, args.scope)
    except (ValueError, KeyError, TypeError, AttributeError, OSError) as error:
        parser.exit(1, f"cannot generate the publish DAG: {error}\n")

    if not plan["packages"] and not plan["images"]:
        print("nothing to publish: no affected packages or images")
        return 0

    workflow = build_workflow(plan, args.branch, args.sha)
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(render_workflow(workflow, args.branch))
    print(
        f"publish DAG: {len(plan['packages'])} package job(s), "
        f"{len(plan['images'])} image job(s) -> {args.out}"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
