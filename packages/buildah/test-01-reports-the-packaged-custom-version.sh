#!/bin/sh
test -x /usr/bin/buildah
test -s /usr/share/licenses/buildah/LICENSE
buildah version | grep -q 'Version: *1\.45\.1'
# Alpine's v3.24 buildah is 1.44.0: this must be our source build.
! buildah version | grep -q 'Version: *1\.44'
