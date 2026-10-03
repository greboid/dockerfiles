#!/bin/sh
# grep -q (not >/dev/null) would exit on first match and SIGPIPE the
# still-writing vips -l: steps run with pipefail, so that surfaces as
# exit status 141. See the melange test notes in MIGRATION.md.
test -x /usr/bin/vips
/usr/bin/vips --version | grep "${PACKAGE_VERSION}" >/dev/null
vips -l | grep heifload >/dev/null
vips -l | grep jpegsave >/dev/null
