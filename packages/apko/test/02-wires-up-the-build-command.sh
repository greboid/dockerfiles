#!/bin/sh
# Smoke beyond `version`: the build command must be wired up in the CLI.
apko build --help > /dev/null
