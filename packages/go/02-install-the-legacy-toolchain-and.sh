#!/bin/sh
dest="${DESTDIR}"
mkdir -p "$dest/usr/local/go"
# Copy only the distribution payload, never the workspace's melange-out,
# downloaded archive or build caches (the output directory is nested here).
cp -a bin pkg src lib VERSION go.env LICENSE PATENTS "$dest/usr/local/go/"
cp -a ./*.md "$dest/usr/local/go/"
rm -rf "$dest/usr/local/go/pkg/"*/cmd \
  "$dest/usr/local/go/pkg/bootstrap" "$dest/usr/local/go/pkg/obj" \
  "$dest/usr/local/go/pkg/tool/"*/api "$dest/usr/local/go/pkg/tool/"*/go_bootstrap \
  "$dest/usr/local/go/test" "$dest/usr/local/go/api" "$dest/usr/local/go/doc" \
  "$dest/usr/local/go/misc" "$dest/usr/local/go/src/cmd"
install -d -m 0777 "$dest/go" "$dest/go/src" "$dest/go/bin"
# Inherited from the legacy Alpine image, including its shell defaults.
mkdir -p "$dest/etc"
printf '[advice]\n    detachedHead = false\n' > "$dest/etc/gitconfig"
