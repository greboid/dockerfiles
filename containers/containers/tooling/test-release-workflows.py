#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["PyYAML==6.0.3"]
# ///
"""Offline checks for .forgejo/workflows/ci.yml and the generated publish DAG.

ci.yml is the entry workflow: the plan job validates the recipe graph,
selects the affected set and checks version bumps, then either generates
the publish DAG — tooling/gen_dag.py writes a per-run workflow whose needs
edges ARE the dependency barrier, pushed to ci/dag-<run_id> where the push
triggers it — or, on pull requests, emits matrices for the packages/images
jobs, which build locally and reference no secrets.

Static checks pin the security shape and the constraints of this Forgejo
instance: no granted permissions, every used action from the instance's
docker action toolkit, checkouts with full history, secrets only in
push-gated steps (pull-request jobs reference no secrets at all), uv
scripts invoked explicitly (BusyBox env cannot exec their
`#!/usr/bin/env -S` shebangs). The generator is checked directly against
the real recipe graph: edge computation (affected closure -> needs), scope
trimming, job-id collision avoidance (image jobs are image- prefixed),
acyclicity, and a YAML render round-trip. Behavioral checks run the
generated publish steps and the matrix emission with fake network tools in
fresh step shells, proving globs, auth files, digest handoffs and the
emitted JSON work without Forgejo.
"""

import contextlib
import io
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

import yaml

import gen_dag

ROOT = Path(__file__).resolve().parent.parent
DIGEST = "sha256:" + "a" * 64
AUDIENCE = "u:1:ea30cbb9-71af-4074-9cf7-603f4e77208f"
REPO_URL = "https://git.mouse-lake.ts.net/api/packages/containers/alpine/v3.24/main"
KEY_URL = "https://git.mouse-lake.ts.net/api/packages/containers/alpine/key"
PUSH = "github.event_name != 'pull_request'"
BASE_EXPRESSION = "inputs.base || github.event.pull_request.base.sha || github.event.before"


def load_workflow(name="ci.yml"):
    return yaml.safe_load((ROOT / ".forgejo/workflows" / name).read_text())


def steps_of(config, job_name):
    return config["jobs"][job_name].get("steps", [])


def steps_referencing(config, needle):
    """Job/step pairs whose env or run text mentions needle."""
    hits = []
    for job_name, job in config["jobs"].items():
        for step in job.get("steps", []):
            if needle in yaml.safe_dump(step):
                hits.append((job_name, step.get("name", "<unnamed>")))
    return hits


def step_by_name(config, job_name, step_name):
    return next(
        s for s in config["jobs"][job_name]["steps"]
        if s.get("name", "<unnamed>") == step_name
    )


def all_steps(config=None):
    config = config if config is not None else load_workflow()
    for job_name, job in config["jobs"].items():
        for step in job.get("steps", []):
            yield job_name, step


def generated_workflow():
    """A generated publish DAG shaped like a real publish run: one leaf
    package and one image whose name collides with nothing."""
    plan = {"packages": {"golink": []}, "images": {"miniflux": []}}
    return gen_dag.build_workflow(plan, "ci/dag-test", "test-commit")


class WorkflowHygiene(unittest.TestCase):
    def test_no_permissions_key(self):
        # Forgejo does not implement workflow-level permissions; the file
        # must not declare one. Isolation is enforced by gated secret steps
        # instead.
        self.assertNotIn("permissions", load_workflow())

    def test_only_toolkit_actions_are_used(self):
        # The instance runs docker actions from actions/toolkit; there is no
        # Node runtime and no GitHub-hosted cache/checkout/artifact actions.
        config = load_workflow()
        used = [
            step.get("uses")
            for job in config["jobs"].values()
            for step in job.get("steps", [])
            if step.get("uses")
        ]
        self.assertTrue(used, "expected at least one action")
        for action in used:
            self.assertTrue(
                re.fullmatch(r"actions/toolkit/[a-z-]+@[A-Za-z0-9._-]+", action),
                f"unsupported action reference: {action}")

    def test_checkouts_have_full_history(self):
        # Every job re-derives its work from the repository, so every
        # checkout needs the merge-base history.
        for job_name, step in all_steps():
            with self.subTest(job=job_name):
                if str(step.get("uses", "")).startswith("actions/toolkit/git-checkout"):
                    self.assertEqual(step["with"]["fetch-depth"], "0", job_name)

    def test_jobs_run_on_alpine(self):
        config = load_workflow()
        for job_name, job in config["jobs"].items():
            self.assertEqual(job.get("runs-on"), "alpine", job_name)

    def test_only_known_secrets_are_referenced(self):
        # Publish-run secrets live in the generated DAG only; the entry
        # workflow needs just the branch-push token.
        used = set(re.findall(r"secrets\.([A-Z_]+)", yaml.safe_dump(load_workflow())))
        self.assertEqual(used, {"RELEASE_PACKAGE_TOKEN"})

    def test_dynamic_matrix_wiring(self):
        # The pull-request fan-out is native: the plan job emits matrices
        # through $GITHUB_OUTPUT, and the build jobs expand them with
        # fromJSON of the plan outputs. Pins the verified-on-this-instance
        # wiring so a refactor cannot quietly fall back to sequential loops.
        config = load_workflow()
        plan = config["jobs"]["plan"]
        outputs = plan.get("outputs", {})
        self.assertEqual(
            outputs,
            {"packages": "${{ steps.matrices.outputs.packages }}",
             "images": "${{ steps.matrices.outputs.images }}"})
        emit = step_by_name(config, "plan", "Emit matrices")
        self.assertEqual(emit.get("id"), "matrices",
                         "job outputs must name the step that writes them")
        self.assertIn("GITHUB_OUTPUT", emit["run"])
        self.assertIn("{package: .}", emit["run"])
        self.assertIn("{image: .}", emit["run"])

        packages = config["jobs"]["packages"]
        self.assertEqual(
            packages["strategy"]["matrix"],
            "${{ fromJSON(needs.plan.outputs.packages) }}")
        self.assertEqual(packages["strategy"]["fail-fast"], False)

        images = config["jobs"]["images"]
        self.assertEqual(
            images["strategy"]["matrix"],
            "${{ fromJSON(needs.plan.outputs.images) }}")
        self.assertEqual(images["strategy"]["fail-fast"], False)
        # Images never wait on the packages matrix job: a zero-leg matrix
        # job has no runtime result, so needs on it would block forever
        # (observed live). Irrelevant on publish runs, which run the
        # generated DAG instead — this wiring is pull-request only.
        self.assertEqual(images["needs"], "plan")

    def test_matrix_jobs_are_pull_request_only(self):
        # Publish runs run the generated DAG; the matrix jobs exist for the
        # secret-free pull-request shape. Scope trimming happens in the
        # generator (the DAG simply has no trimmed side), not in job gates.
        config = load_workflow()
        packages_gate = str(config["jobs"]["packages"].get("if", ""))
        images_gate = str(config["jobs"]["images"].get("if", ""))
        for gate, job in ((packages_gate, "packages"), (images_gate, "images")):
            self.assertIn("github.event_name == 'pull_request'", gate, job)
        self.assertIn("inputs.scope != 'images'", packages_gate)
        self.assertIn("inputs.scope != 'packages'", images_gate)
        generate = step_by_name(config, "plan", "Generate and push the publish DAG")
        self.assertIn("inputs.scope || 'all'", yaml.safe_dump(generate))

    def test_python_steps_go_through_uv(self):
        # The alpine job image's BusyBox env cannot exec the scripts'
        # `#!/usr/bin/env -S` shebangs; run blocks must invoke uv scripts
        # explicitly.
        for job_name, step in all_steps():
            with self.subTest(job=job_name, step=step.get("name")):
                run = step.get("run")
                if not run:
                    continue
                for line in run.splitlines():
                    stripped = line.strip()
                    if stripped.startswith("#") or ".py" not in stripped:
                        continue
                    self.assertIn(
                        "uv run --script", stripped,
                        f"{job_name}/{step.get('name')}: {stripped}")

    def test_go_setup_precedes_tool_builds(self):
        # The alpine job image has no Go toolchain; the tools build needs
        # go-setup's exported PATH/GOROOT before it runs.
        config = load_workflow()
        for job_name in ("packages", "images"):
            names = [step.get("name") for step in config["jobs"][job_name]["steps"]]
            with self.subTest(job=job_name):
                self.assertIn("Build tools", names)
                self.assertIn("Set up Go", names)
                self.assertLess(names.index("Set up Go"), names.index("Build tools"))

    def test_bubblewrap_installed_for_melange(self):
        # melange's default container runner needs bwrap on PATH; the alpine
        # job image does not ship it.
        installs = [
            step.get("run", "")
            for _, step in all_steps()
            if "apk add" in step.get("run", "")
        ]
        self.assertTrue(any("bubblewrap" in run for run in installs))

    def test_offline_check_job_installs_jq(self):
        # The plan job runs the offline suite (whose behavioral tests
        # exercise jq-driven steps) and compacts matrices with jq.
        config = load_workflow()
        installs = [
            step.get("run", "")
            for step in config["jobs"]["plan"]["steps"]
            if "apk add" in step.get("run", "")
        ]
        self.assertTrue(any("jq" in run for run in installs), "plan must install jq")

    def test_melange_tests_use_the_default_runner(self):
        # melange assembles its build/test workspace under the caller's
        # /tmp; from inside the job container the docker runner cannot
        # bind-mount that path (the daemon sees the host filesystem).
        # Builds and tests therefore both run under bwrap in CI.
        text = (ROOT / ".forgejo/workflows/ci.yml").read_text()
        self.assertNotIn("--runner docker", text)

    def test_nothing_polls_for_publications(self):
        # The poll design held runner slots while waiting and deadlocked a
        # capacity-2 runner; needs edges in the generated DAG replaced it.
        # The wait script stays for manual use only.
        text = (ROOT / ".forgejo/workflows/ci.yml").read_text()
        self.assertNotIn("await-published.py", text)


class SecretGating(unittest.TestCase):
    def test_secret_steps_are_push_gated(self):
        # Manual dispatches and pushes publish; pull requests never touch a
        # secret. Every secret-bearing step must carry the gate.
        config = load_workflow()
        referencing = steps_referencing(config, "secrets.")
        self.assertTrue(referencing, "expected secret steps")
        for job_name, step_name in referencing:
            condition = str(step_by_name(config, job_name, step_name).get("if", ""))
            self.assertEqual(
                condition, PUSH,
                f"{job_name}/{step_name} uses a secret but is not push-gated")

    def test_pull_request_steps_are_secret_free(self):
        # The PR-shape steps (throwaway key, local image build) run only on
        # pull requests and must reference no secrets.
        config = load_workflow()
        for job_name, step in all_steps():
            condition = str(step.get("if", ""))
            if "github.event_name == 'pull_request'" in condition:
                self.assertNotIn(
                    "secrets.", yaml.safe_dump(step),
                    f"{job_name}/{step.get('name')}")

    def test_ci_has_no_publish_steps(self):
        # Publishing lives in the generated DAG (whose jobs are inherently
        # publish-only, triggered only by the branch push). The entry
        # workflow must not carry release keys, uploads or registry logins;
        # RELEASE_PACKAGE_TOKEN appears solely in the push-gated branch
        # push (and, in the generated DAG, the package upload and cleanup).
        text = yaml.safe_dump(load_workflow())
        self.assertNotIn("APK_SIGNING_KEY", text)
        self.assertNotIn("docker-login", text)
        self.assertNotIn("apko publish", text)

    def test_pull_request_packages_use_the_throwaway_key(self):
        # PR legs sign with a leg-local throwaway key trusted by nothing
        # outside their own build and test guests.
        config = load_workflow()
        keygen = step_by_name(config, "packages", "Generate throwaway signing key")
        self.assertIn("melange keygen", keygen["run"])
        build = step_by_name(config, "packages", "Build and test package")
        self.assertIn("keys/ci.rsa", build["run"])
        self.assertNotIn("--remote-repo", build["run"],
                         "PR legs resolve everything locally")

    def test_dag_packages_sign_with_the_org_key(self):
        # Publish jobs sign with the organisational APK_SIGNING_KEY and
        # resolve dependencies from the remote repository; the needs edges
        # guarantee the affected ones are published by then.
        build = next(
            step for step in gen_dag.PACKAGE_STEPS
            if step.get("name") == "Build and test package")
        self.assertIn("keys/release.rsa", build["run"])
        self.assertIn("--remote-repo", build["run"])

    def test_plan_runs_the_bump_check(self):
        # DAG jobs wait on exact published file names; a recipe changed
        # without a version or epoch bump would be indistinguishable from
        # its previous build. The check skips itself when the base is empty
        # (republish dispatch) or unusable.
        bumps = step_by_name(load_workflow(), "plan", "Check version bumps")
        self.assertIn("check-bumps.py", bumps["run"])

    def test_base_expression_covers_every_event(self):
        # Manual dispatch passes the base explicitly, pull requests diff
        # from their base, pushes from the previous tip.
        config = load_workflow()
        self.assertEqual(
            config["env"]["BASE"],
            "${{ " + BASE_EXPRESSION + " }}")

    def test_dag_generation_is_push_gated(self):
        # The publish DAG only exists for publish runs; pull requests must
        # not push branches (they have no token either).
        config = load_workflow()
        generate = step_by_name(config, "plan", "Generate and push the publish DAG")
        self.assertEqual(generate.get("if"), PUSH)
        self.assertIn("gen_dag.py", generate["run"])
        self.assertIn("--branch", generate["run"])
        self.assertIn("refs/heads/", generate["run"])


class GeneratedDag(unittest.TestCase):
    """tooling/gen_dag.py against the real recipe graph."""

    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.lists = Path(self.temp.name) / "affected"
        self.lists.mkdir(parents=True)

    def write_lists(self, packages=(), images=(), image_deps=()):
        (self.lists / "packages.txt").write_text(
            "".join(f"{name}\n" for name in packages))
        (self.lists / "images.txt").write_text(
            "".join(f"{name}\n" for name in images))
        (self.lists / "image-deps.txt").write_text(
            "".join(f"{line}\n" for line in image_deps))

    def plan(self, **kwargs):
        return gen_dag.plan_dag(self.lists, ROOT, kwargs.get("scope", "all"))

    def test_leaf_package_needs_nothing(self):
        # An affected leaf whose local closure is unaffected waits on
        # nobody: everything it consumes is already published.
        self.write_lists(packages=["golink"])
        self.assertEqual(self.plan(), {"packages": {"golink": []}, "images": {}})

    def test_dependent_needs_its_affected_dependency(self):
        # golink build-depends on go-licenses; when both are affected the
        # edge makes Forgejo schedule golink only after go-licenses
        # published. The edge replaces the old publication poll.
        self.write_lists(packages=["go-licenses", "golink"])
        plan = self.plan()
        self.assertEqual(plan["packages"]["go-licenses"], [])
        self.assertEqual(plan["packages"]["golink"], ["go-licenses"])

    def test_unaffected_dependency_appears_in_no_edge(self):
        # The closure is intersected with the affected set: an unaffected
        # dependency is published already, so the edge would only add
        # queueing, not freshness.
        self.write_lists(packages=["golink"])
        plan = self.plan()
        self.assertNotIn("go-licenses", plan["packages"]["golink"])

    def test_image_needs_only_its_affected_closure(self):
        # From image-deps.txt (the image's full local closure), only the
        # affected part becomes edges — an image is only ever queued on the
        # packages it installs that are actually being rebuilt.
        self.write_lists(
            packages=["go-licenses", "golink"],
            images=["miniflux"],
            image_deps=["miniflux go-licenses golink"],
        )
        plan = self.plan()
        self.assertEqual(plan["images"]["miniflux"], ["go-licenses", "golink"])

    def test_scope_packages_drops_images(self):
        self.write_lists(
            packages=["golink"], images=["miniflux"],
            image_deps=["miniflux golink"])
        plan = self.plan(scope="packages")
        self.assertEqual(plan["images"], {})
        self.assertIn("golink", plan["packages"])

    def test_scope_images_drops_package_edges_with_a_warning(self):
        # The trimmed side cannot be waited on; the edges are dropped and
        # the image will fail resolution unless the packages were already
        # published — which is the dispatcher's mistake, said loudly.
        self.write_lists(
            packages=["golink"], images=["miniflux"],
            image_deps=["miniflux golink"])
        stderr = self.enterContext(_capture_stderr())
        plan = self.plan(scope="images")
        self.assertEqual(plan["packages"], {})
        self.assertEqual(plan["images"]["miniflux"], [])
        self.assertIn("warning", stderr.getvalue())

    def test_unknown_affected_package_rejected(self):
        self.write_lists(packages=["not-a-recipe"])
        with self.assertRaises(ValueError):
            self.plan()

    def test_image_without_deps_line_rejected(self):
        self.write_lists(packages=[], images=["miniflux"], image_deps=[])
        with self.assertRaises(ValueError):
            self.plan()

    def test_invalid_image_name_rejected(self):
        self.write_lists(packages=[], images=["../evil"], image_deps=["../evil"])
        with self.assertRaises(ValueError):
            self.plan()

    def test_empty_selection_generates_nothing(self):
        # A docs-only push: no file, exit 0 — the plan step treats a
        # missing output as "nothing to publish".
        self.write_lists()
        self.assertEqual(self.plan(), {"packages": {}, "images": {}})
        out = Path(self.temp.name) / "publish-dag.yml"
        argv = [
            "gen_dag.py", "--lists-dir", str(self.lists), "--scope", "all",
            "--sha", "test-commit", "--branch", "ci/dag-test", "--out", str(out),
        ]
        with mock.patch.object(sys, "argv", argv), \
                contextlib.redirect_stdout(io.StringIO()):
            self.assertEqual(gen_dag.main(), 0)
        self.assertFalse(out.exists())

    def test_workflow_shape(self):
        self.write_lists(
            packages=["go-licenses", "golink"],
            images=["miniflux"],
            image_deps=["miniflux go-licenses golink"],
        )
        plan = self.plan()
        workflow = gen_dag.build_workflow(plan, "ci/dag-7", "test-commit")
        jobs = workflow["jobs"]
        # Job ids: bare recipe names for packages, image- prefixed for
        # images (names overlap across the two: miniflux, golink, ...).
        self.assertEqual(
            list(jobs),
            ["go-licenses", "golink", "image-miniflux", "cleanup"])
        self.assertEqual(jobs["golink"]["needs"], ["go-licenses"])
        self.assertEqual(jobs["image-miniflux"]["needs"], ["go-licenses", "golink"])
        self.assertEqual(
        jobs["golink"]["env"], {"PACKAGE": "golink", "OWNER": gen_dag.OWNER})
        self.assertEqual(jobs["image-miniflux"]["env"]["IMAGE"], "miniflux")
        self.assertEqual(jobs["image-miniflux"]["env"]["TAG"], "${{ env.SHA }}")
        self.assertEqual(jobs["cleanup"]["needs"], list(jobs)[:-1])
        self.assertEqual(jobs["cleanup"]["if"], "always()")
        self.assertTrue(workflow["enable-openid-connect"])
        self.assertNotIn("permissions", workflow)
        self.assertEqual(workflow["env"]["SHA"], "test-commit")
        self.assertEqual(workflow["env"]["DAG_BRANCH"], "ci/dag-7")
        self.assertEqual(workflow["on"]["push"]["branches"], ["ci/dag-7"])
        for job_name, job in jobs.items():
            with self.subTest(job=job_name):
                self.assertEqual(job["runs-on"], "alpine")
                for step in job["steps"]:
                    action = step.get("uses")
                    if action:
                        self.assertRegex(
                            action, r"actions/toolkit/[a-z-]+@[A-Za-z0-9._-]+")

    def test_render_round_trip(self):
        # The rendered YAML must parse back to the same structure: no
        # anchors/aliases, no boolean-coerced keys, literal blocks intact.
        self.write_lists(
            packages=["go-licenses", "golink"],
            images=["miniflux"],
            image_deps=["miniflux go-licenses golink"],
        )
        workflow = gen_dag.build_workflow(self.plan(), "ci/dag-7", "test-commit")
        rendered = gen_dag.render_workflow(workflow, "ci/dag-7")
        self.assertEqual(yaml.safe_load(rendered), workflow)
        self.assertNotIn("&id", rendered)
        self.assertNotIn("runner docker", rendered)

    def test_generated_secrets(self):
        workflow = generated_workflow()
        used = set(re.findall(r"secrets\.([A-Z_]+)", yaml.safe_dump(workflow)))
        self.assertEqual(
            used, {"APK_SIGNING_KEY", "RELEASE_PACKAGE_TOKEN"})

    def test_registry_login_uses_authorized_integration(self):
        # Registry pushes authenticate with the Authorized Integration JWT
        # minted from the runner's OIDC endpoint — never a shared secret.
        logins = [
            step for step in gen_dag.IMAGE_STEPS
            if str(step.get("uses", "")).startswith("actions/toolkit/docker-login")
        ]
        self.assertEqual(len(logins), 1)
        self.assertNotIn("password", logins[0].get("with", {}), "shared-secret login")
        self.assertEqual(logins[0]["with"].get("audience"), AUDIENCE)

@contextlib.contextmanager
def _capture_stderr():
    import io
    stream = io.StringIO()
    with contextlib.redirect_stderr(stream):
        yield stream


class EmittedSteps(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.output = self.root / "github-output"
        self.ci = load_workflow()
        self.dag = generated_workflow()
        self.env = {
            "PATH": f"{self.root}/bin:{os.environ['PATH']}",
            "HOME": str(self.root),
            "GITHUB_OUTPUT": str(self.output),
        }
        self.values = {
            "github.sha": "test-commit",
            BASE_EXPRESSION: "base-sha",
            "env.SHA": "test-commit",
            "secrets.RELEASE_PACKAGE_TOKEN": "dummy-token",
            "secrets.APK_SIGNING_KEY": "dummy-key",
        }

    def begin(self):
        """Fresh workspace for one subTest iteration."""
        shutil.rmtree(self.root, ignore_errors=True)
        (self.root / "bin").mkdir(parents=True)
        self.output.parent.mkdir(parents=True, exist_ok=True)
        self.output.write_text("")

    def write_affected(self, packages=("miniflux",), images=("miniflux",)):
        affected = self.root / "dist/affected"
        affected.mkdir(parents=True, exist_ok=True)
        (affected / "packages.txt").write_text(
            "".join(f"{name}\n" for name in packages))
        (affected / "images.txt").write_text(
            "".join(f"{name}\n" for name in images))

    def mock(self, path, script):
        target = self.root / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text("#!/bin/sh\nset -eu\n" + script)
        target.chmod(0o755)

    def run_step(self, config, job, name, expect_failure=False):
        step = step_by_name(config, job, name)
        env = self.env.copy()
        for layer in (config.get("env", {}), config["jobs"][job].get("env", {}), step.get("env", {})):
            for key, value in layer.items():
                env[key] = re.sub(
                    r"\$\{\{\s*(.*?)\s*\}\}",
                    lambda m: self.values[m[1]],
                    str(value),
                )
        # Deliberately no shell exports survive between steps.
        result = subprocess.run(
            ["sh", "-c", step["run"]], cwd=self.root, env=env,
            capture_output=True, text=True,
        )
        if expect_failure:
            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
        else:
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def test_emit_matrices_compacts_selection(self):
        # The matrices ride through $GITHUB_OUTPUT as single-line JSON:
        # pretty-printed values would break the output file format.
        self.begin()
        self.write_affected(packages=("miniflux", "golink"), images=("miniflux",))
        self.run_step(self.ci, "plan", "Emit matrices")
        lines = dict(
            line.split("=", 1)
            for line in self.output.read_text().splitlines()
        )
        self.assertEqual(
            lines["packages"], '{"package":["miniflux","golink"]}')
        self.assertEqual(lines["images"], '{"image":["miniflux"]}')

    def test_emit_matrices_tolerate_empty_selection(self):
        # A docs-only push selects nothing: empty matrices, zero legs.
        self.begin()
        self.write_affected(packages=(), images=())
        self.run_step(self.ci, "plan", "Emit matrices")
        lines = dict(
            line.split("=", 1)
            for line in self.output.read_text().splitlines()
        )
        self.assertEqual(lines["packages"], '{"package":[]}')
        self.assertEqual(lines["images"], '{"image":[]}')

    def test_package_key_fetch_targets_org_level_endpoint(self):
        # The repository signing key is served at the organisation level
        # (/api/packages/{owner}/alpine/key, no branch/repository segments);
        # fetching it from the branch/repo-level URL is a 404.
        self.begin()
        self.mock("bin/curl", '''
out=''
url=''
while [ "$#" -gt 0 ]; do
    case "$1" in
        -o) out=$2; shift 2 ;;
        https://*) url=$1; shift ;;
        *) shift ;;
    esac
done
printf '%s\\n' "$url" > key-url
printf -- '-----BEGIN PUBLIC KEY-----\\nkey\\n-----END PUBLIC KEY-----\\n' > "$out"
''')
        self.run_step(self.dag, "golink", "Fetch Forgejo repository signing key")
        self.assertEqual(
            (self.root / "key-url").read_text().strip(), KEY_URL)

    def test_package_publish_and_roundtrip(self):
        self.begin()
        repo = self.root / "dist/release/x86_64"
        repo.mkdir(parents=True)
        (repo / "golink-1.0.0-r0.apk").write_bytes(b"application package")
        (repo / "go-licenses-1.0.0-r0.apk").write_bytes(b"build dependency")
        self.mock("bin/curl", '''
output=''
file=''
url=''
while [ "$#" -gt 0 ]; do
    case "$1" in
        -o) output=$2; shift 2 ;;
        -F) file=${2#file=@}; shift 2 ;;
        https://*) url=$1; shift ;;
        *) shift ;;
    esac
done
if [ -n "$file" ]; then
    test -f "$file"
    printf '%s\\n' "$file" >> uploads
    cp "$file" published.apk
else
    test "$url" = 'https://git.mouse-lake.ts.net/api/packages/containers/alpine/v3.24/main/x86_64/golink-1.0.0-r0.apk'
    cp published.apk "$output"
fi
''')
        self.run_step(self.dag, "golink", "Publish package (RELEASE_PACKAGE_TOKEN)")
        self.run_step(self.dag, "golink", "Verify anonymous round trip")
        # Only the requested package uploads: its build dependencies share
        # the repository but are not release artifacts.
        self.assertEqual((self.root / "uploads").read_text().splitlines(), [
            "dist/release/x86_64/golink-1.0.0-r0.apk",
        ])

    def test_image_publish_and_latest(self):
        self.begin()
        # The docker-login step (a separate action, not run here) leaves the
        # registry credentials in the workspace; the key fetch step leaves
        # the repository key under its signature name.
        docker_config = self.root / ".actions/docker"
        docker_config.mkdir(parents=True)
        (docker_config / "config.json").write_text('{"auths":{}}')
        (self.root / "forgejo-key-signature-name").write_text(
            "containers@abc.rsa.pub\n")
        (self.root / "containers@abc.rsa.pub").write_text("public key\n")
        self.mock("tooling/bin/apko", f"printf '%s\\n' '{DIGEST}'\n")
        self.mock("tooling/bin/oras", '''
test "$DOCKER_CONFIG" = "$PWD/.actions/docker"
test -s "$DOCKER_CONFIG/config.json"
printf '%s\\n' "$@" > oras-args
''')
        self.run_step(self.dag, "image-miniflux", "Publish image (authorized integration)")
        # Publishing and advancing latest happen in the same leg: the
        # digest is known locally, there is no cross-job handoff. TAG is
        # the original pushed SHA (env.SHA), not the DAG branch head.
        self.assertEqual((self.root / "oras-args").read_text().splitlines(), [
            "cp", f"git.mouse-lake.ts.net/containers/miniflux@{DIGEST}",
            "git.mouse-lake.ts.net/containers/miniflux:latest",
        ])

    def test_image_publish_trusts_signature_named_key(self):
        # apko matches index signatures to keyring entries by file name;
        # the publish step must pass the signature-named key, not the
        # downloaded alias.
        self.begin()
        docker_config = self.root / ".actions/docker"
        docker_config.mkdir(parents=True)
        (docker_config / "config.json").write_text('{"auths":{}}')
        (self.root / "forgejo-key-signature-name").write_text(
            "containers@abc.rsa.pub\n")
        (self.root / "containers@abc.rsa.pub").write_text("public key\n")
        self.mock("tooling/bin/apko", '''
key=''
while [ "$#" -gt 0 ]; do
    case "$1" in
        -k) key=$2; shift 2 ;;
        *) shift ;;
    esac
done
test -f "$key" || { echo "keyring file missing: $key" >&2; exit 1; }
test "$key" != "./forgejo-alpine-key.rsa.pub" || { echo "alias passed" >&2; exit 1; }
printf '%s\\n' 'sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
''')
        self.mock("tooling/bin/oras", "true")
        self.run_step(self.dag, "image-miniflux", "Publish image (authorized integration)")


if __name__ == "__main__":
    unittest.main()
