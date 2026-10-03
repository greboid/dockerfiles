#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Offline checks for tooling/affected.py against a synthetic git repository.

The fixture graph mirrors the real conventions:

  os-release    no dependencies
  go-licenses   no dependencies
  golink        build-depends on go-licenses, runtime-depends on os-release
  miniflux      build-depends on go-licenses
  image golink   installs golink=1.0.0-r1 only (os-release arrives through
                golink's runtime closure)
"""

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

TOOLING = Path(__file__).resolve().parent
AFFECTED = TOOLING / "affected.py"

RECIPE = """package:
  name: {name}
  version: {version}
  epoch: 0
  description: test recipe
  target-architecture:
    - x86_64
  copyright:
    - paths:
        - "*"
      attestation: test
      license: MIT
{extra}"""


def build_env(dependencies):
    if not dependencies:
        return ""
    listed = "\n".join(f"        - {dep}" for dep in dependencies)
    return f"""environment:
  contents:
    repositories:
      - https://dl-cdn.alpinelinux.org/alpine/v3.24/main
    packages:
{listed}
"""


def recipe(name, version="1.0.0", build_deps=(), runtime_deps=()):
    # package.dependencies must stay inside the package block; it comes
    # before the environment section the helper appends.
    extra = ""
    if runtime_deps:
        listed = "\n".join(f"      - {dep}" for dep in runtime_deps)
        extra += f"  dependencies:\n    runtime:\n{listed}\n"
    extra += build_env(build_deps)
    return RECIPE.format(name=name, version=version, extra=extra)


IMAGE = """archs:
  - x86_64

contents:
  repositories:
    - https://dl-cdn.alpinelinux.org/alpine/v3.24/main
  packages:
    - golink=1.0.0-r0

entrypoint:
  command: /usr/bin/golink
"""


class AffectedSelection(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.repo = Path(self.temp.name) / "repo"
        (self.repo / "packages").mkdir(parents=True)
        (self.repo / "images").mkdir()
        (self.repo / "tooling").mkdir()
        subprocess.run(["git", "init", "-q", str(self.repo)], check=True)
        self.write("packages/os-release.yaml", recipe("os-release"))
        self.write("packages/go-licenses.yaml", recipe("go-licenses"))
        self.write("packages/golink.yaml", recipe("golink", build_deps=["go-licenses"], runtime_deps=["os-release"]))
        self.write("packages/miniflux.yaml", recipe("miniflux", build_deps=["go-licenses"]))
        self.write("images/golink.yaml", IMAGE)
        self.commit("fixture")

    def write(self, path, content):
        target = self.repo / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content)

    def commit(self, message) -> str:
        self.run_git("add", "-A")
        self.run_git("-c", "user.email=test@example.com", "-c", "user.name=test",
                     "commit", "-q", "--allow-empty", "-m", message)
        return self.run_git("rev-parse", "HEAD").strip()

    def run_git(self, *args):
        return subprocess.run(
            ["git", "-C", str(self.repo), *args], check=True, capture_output=True, text=True
        ).stdout

    def affected(self, base, head="HEAD", *extra):
        result = subprocess.run(
            ["uv", "run", "--quiet", "--script", str(AFFECTED),
             "--root", str(self.repo), "--base", base, "--head", head, *extra],
            capture_output=True, text=True,
        )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return json.loads(result.stdout), result.stderr

    def test_unchanged_range_selects_nothing(self):
        payload, _ = self.affected(self.commit("empty"))
        self.assertEqual(payload["packages"], [])
        self.assertEqual(payload["images"], [])
        self.assertFalse(payload["everything"])

    def test_recipe_change_selects_dependents_and_consumers(self):
        base = self.commit("base")
        self.write("packages/golink.yaml", recipe("golink", version="1.0.1", build_deps=["go-licenses"], runtime_deps=["os-release"]))
        self.commit("touch golink")
        payload, _ = self.affected(base)
        self.assertEqual(payload["packages"], ["golink"])
        self.assertEqual(payload["images"], [{"name": "golink", "deps": ["os-release", "golink"]}])

    def test_build_dependency_change_selects_reverse_closure(self):
        base = self.commit("base")
        self.write("packages/go-licenses.yaml", recipe("go-licenses", version="1.0.1"))
        self.commit("touch go-licenses")
        payload, _ = self.affected(base)
        self.assertEqual(payload["packages"], ["go-licenses", "golink", "miniflux"])
        self.assertEqual([image["name"] for image in payload["images"]], ["golink"])

    def test_image_change_needs_no_package_rebuild(self):
        base = self.commit("base")
        self.write("images/golink.yaml", IMAGE + "\n# touched\n")
        self.commit("touch image")
        payload, _ = self.affected(base)
        self.assertEqual(payload["packages"], [])
        self.assertEqual(payload["images"], [{"name": "golink", "deps": ["os-release", "golink"]}])

    def test_inert_change_selects_nothing(self):
        base = self.commit("base")
        self.write("README.md", "# docs\n")
        self.commit("docs only")
        payload, _ = self.affected(base)
        self.assertEqual(payload["packages"], [])
        self.assertEqual(payload["images"], [])

    def test_melange_tooling_selects_everything(self):
        base = self.commit("base")
        self.write("tooling/melange/patches/new.patch", "--- a\n+++ b\n")
        self.commit("melange patch")
        payload, _ = self.affected(base)
        self.assertEqual(
            payload["packages"], ["os-release", "go-licenses", "golink", "miniflux"]
        )
        self.assertEqual([image["name"] for image in payload["images"]], ["golink"])

    def test_apko_tooling_selects_images_only(self):
        base = self.commit("base")
        self.write("tooling/apko/patches/new.patch", "--- a\n+++ b\n")
        self.commit("apko patch")
        payload, _ = self.affected(base)
        self.assertEqual(payload["packages"], [])
        self.assertEqual([image["name"] for image in payload["images"]], ["golink"])

    def test_oras_tooling_selects_nothing(self):
        base = self.commit("base")
        self.write("tooling/oras/patches/new.patch", "--- a\n+++ b\n")
        self.commit("oras patch")
        payload, _ = self.affected(base)
        self.assertEqual(payload["packages"], [])
        self.assertEqual(payload["images"], [])

    def test_unusable_base_selects_everything(self):
        for base in ("", "0" * 40, "deadbeef"):
            payload, _ = self.affected(base)
            self.assertTrue(payload["everything"], f"base={base!r}")
            self.assertEqual(len(payload["packages"]), 4)
            self.assertEqual(len(payload["images"]), 1)

    def test_deleted_recipe_warns_but_selects_dependents(self):
        base = self.commit("base")
        (self.repo / "packages/golink.yaml").unlink()
        self.commit("delete golink")
        payload, stderr = self.affected(base)
        self.assertIn("no such recipe exists", stderr)
        self.assertNotIn("golink", payload["packages"])

    def test_lists_dir_format(self):
        base = self.commit("base")
        self.write("packages/golink.yaml", recipe("golink", version="1.0.1", build_deps=["go-licenses"], runtime_deps=["os-release"]))
        self.commit("touch golink")
        lists = self.repo / "affected"
        _, stderr = self.affected(base, "HEAD", "--explain", f"--lists-dir={lists}")
        self.assertIn("package golink:", stderr)
        self.assertEqual((lists / "packages.txt").read_text(), "golink\n")
        self.assertEqual((lists / "images.txt").read_text(), "golink\n")
        self.assertEqual((lists / "image-deps.txt").read_text(), "golink os-release golink\n")

    def test_invalid_image_structure_fails(self):
        self.write("images/golink.yaml", "contents:\n  packages: golink\n")
        self.commit("break image")
        result = subprocess.run(
            ["uv", "run", "--quiet", "--script", str(AFFECTED),
             "--root", str(self.repo), "--base", "HEAD"],
            capture_output=True, text=True,
        )
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("contents.packages", result.stderr)


if __name__ == "__main__":
    sys.exit(unittest.main())
