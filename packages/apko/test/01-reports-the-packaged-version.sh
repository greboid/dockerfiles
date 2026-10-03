#!/bin/sh
# The patched build must report the packaged version and the pinned upstream
# commit it was built from.
apko version | grep -q 'GitVersion: *v1\.4\.4'
apko version | grep -q 'GitCommit: *7e72102a642f1ce74cbca37ef85f8b925d02a11b'
