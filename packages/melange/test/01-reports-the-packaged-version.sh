#!/bin/sh
# The patched build must report the packaged version and the pinned upstream
# commit it was built from.
melange version | grep -q 'GitVersion: *v0\.61\.1'
melange version | grep -q 'GitCommit: *d062cf25ce373bf8045d79b9897878ab207d8d36'
