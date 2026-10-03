"""Shared recipe dependency planning for the tooling scripts.

Recipes are self-describing: a package that needs another local package
(e.g. golink needs go-licenses) lists it in environment.contents.packages.
The graph here resolves those references to recipe names; all-packages.py
uses it for ordered builds, single-package.py for one package's build-deps.
"""

from pathlib import Path
import re

import yaml


ROOT = Path(__file__).resolve().parent.parent


def load_recipes(root=ROOT):
    """Load every recipe; return (recipes, providers, runtimes).

    recipes maps recipe name -> parsed YAML. providers maps every package
    output (main packages and subpackages) -> owning recipe name. runtimes
    maps package output -> its declared runtime dependencies. Used both for
    the forward build graph and for reverse lookups (which recipes or images
    consume a changed package).
    """
    recipes = {}
    providers = {}
    runtimes = {}
    for path in sorted(Path(root, "packages").glob("*.yaml")):
        name = path.stem
        if not re.fullmatch(r"[a-z0-9-]+", name):
            raise ValueError(f"invalid recipe name: {name}")
        with path.open() as source:
            recipe = yaml.safe_load(source)
        if recipe["package"]["name"] != name:
            raise ValueError(f"{path}: package name must match filename")
        recipes[name] = recipe
        for package in [recipe["package"], *recipe.get("subpackages", [])]:
            output = package["name"]
            if output in providers:
                raise ValueError(f"duplicate local package: {output}")
            providers[output] = name
            runtimes[output] = package.get("dependencies", {}).get("runtime", [])
    if not recipes:
        raise ValueError("no package recipes found")
    return recipes, providers, runtimes


def build_graph(root=ROOT):
    """Return {recipe name: set of local recipe names it build-depends on}."""
    recipes, providers, runtimes = load_recipes(root)
    graph = {}
    for name, recipe in recipes.items():
        dependencies = list(recipe.get("environment", {}).get("contents", {}).get("packages", []))
        graph[name] = set()
        installed = set()
        # Installing a local build dependency also installs its runtime closure.
        # Runtime dependencies alone do not prevent packaging an application.
        while dependencies:
            dependency = dependencies.pop()
            if dependency.startswith("!"):
                continue  # APK exclusion, not a dependency
            package = re.split(r"[<>=~@]", dependency, maxsplit=1)[0]
            provider = providers.get(package)
            if provider is not None and package not in installed:
                installed.add(package)
                graph[name].add(provider)
                dependencies.extend(runtimes[package])
    return graph


def build_order(root=ROOT):
    """Return recipe names in dependency order (os-release first)."""
    graph = build_graph(root)
    order = []
    visiting = []
    visited = set()

    def visit(name):
        if name in visiting:
            raise ValueError("local dependency cycle: " + " -> ".join(visiting + [name]))
        if name in visited:
            return
        visiting.append(name)
        for dependency in sorted(graph[name]):
            visit(dependency)
        visiting.pop()
        visited.add(name)
        order.append(name)

    # OS metadata is shared by all images, although it is not a build dependency.
    for name in sorted(graph, key=lambda name: (name != "os-release", name)):
        visit(name)
    return order


def closure_of(target, graph, providers, runtimes, recipes):
    """Every local recipe the target consumes, directly or transitively.

    Build dependencies of the target's build environment, plus the runtime
    closures of everything in the set: a build guest resolving the target's
    package list walks runtimes too, so a missing runtime provider would
    break it just the same. This is exactly the set a publish run must have
    freshly published (the affected part of it) before the target builds —
    the DAG generator turns it into needs edges.
    """
    seen = set()
    work = [target, *graph[target]]
    while work:
        name = work.pop()
        if name in seen:
            continue
        seen.add(name)
        recipe = recipes[name]
        for package in [recipe["package"], *recipe.get("subpackages", [])]:
            for dependency in package.get("dependencies", {}).get("runtime", []):
                provider = providers.get(dependency)
                if provider is not None:
                    work.append(provider)
    return seen
